// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/cisco-eti/wfsm/internal/platforms"
	"github.com/cisco-eti/wfsm/internal/wfsm/config"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/builder"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/cisco-eti/wfsm/internal/wfsm/manifest"
)

var deployLongHelp = `
This command takes two required flags: --manifestPath path/to/acpManifest

Optional flags:
	--envFilePath path/to/envConfigFile 
  --configPath path/to/configFile
	--baseImage can be set to determine which base image is used as the workflowserver for the agent.
	--deleteBuildFolders can be set to true or false to determine if the build folders should be deleted after deployment.
	--deploymentOption can be set to determine which deployment option to use from the manifest. It defaults to the first deployment option.
	--dryRun if set to true, the deployment will not be executed, instead deployment artifacts will be printed to the console.
	--forceBuild can be set to true or false to determine if the build should be forced even if the image already exists.
	--platform specify the platform to deploy the agent(s) to [docker, k8s].

Env config file should be a yaml file in the format of 'EnvVarValues' (see manifest format).
Example:

values:
  ENV_VAR_1: "sample value 1"
dependencies:
  - name: <agent_dependency_name>
    values:
      ENV_VAR_2: "sample value 2"
		
Examples:
- Build an agent with a manifest and environment file:
	wfsm deploy --manifestPath path/to/acpManifest --envFilePath path/to/envConfigFile
`

const deployFail = "Deploy Status: Failed - %s"
const deployError string = "deploy failed"

const baseImageFlag string = "baseImage"
const deleteBuildFoldersFlag string = "deleteBuildFolders"
const deploymentOptionFlag string = "deploymentOption"
const dryRunFlag string = "dryRun"
const envFilePathFlag string = "envFilePath"
const forceBuild string = "forceBuild"
const manifestPathFlag string = "manifestPath"
const configPathFlag string = "configPath"

// deployCmd represents the image build and run docker commands
var deployCmd = &cobra.Command{
	Use:   "deploy --manifestPath path/to/acpManifest --envFilePath path/to/envFile",
	Short: "Build an ACP agent",
	Long:  deployLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {

		baseImage, _ := cmd.Flags().GetString(baseImageFlag)
		deleteBuildFolders, _ := cmd.Flags().GetBool(deleteBuildFoldersFlag)
		deploymentOption, _ := cmd.Flags().GetString(deploymentOptionFlag)
		dryRun, _ := cmd.Flags().GetBool(dryRunFlag)
		envFilePath, _ := cmd.Flags().GetString(envFilePathFlag)
		configPathFlag, _ := cmd.Flags().GetString(configPathFlag)
		forceBuild, _ := cmd.Flags().GetBool(forceBuild)
		manifestPath, _ := cmd.Flags().GetString(manifestPathFlag)
		platform, _ := cmd.Flags().GetString(platformsFlag)

		err := runDeploy(getContextWithLogger(cmd), manifestPath, envFilePath, configPathFlag, platform,
			dryRun, deleteBuildFolders, forceBuild, baseImage, &deploymentOption)
		if err != nil {
			util.OutputMessage(deployFail, err.Error())
			return fmt.Errorf(CmdErrorHelpText, deployError)
		}
		return nil
	},
}

func init() {
	deployCmd.Flags().StringP(baseImageFlag, "b", "", "Base image to be used as the workflowserver for the agent, repo is at ghcr.io/agntcy/acp/wfsrv")
	deployCmd.Flags().BoolP(deleteBuildFoldersFlag, "d", true, "Delete build folders after deployment")
	deployCmd.Flags().StringP(deploymentOptionFlag, "o", "", "Deployment option to use from the manifest")
	deployCmd.Flags().BoolP(dryRunFlag, "r", false, "If set to true, the deployment will not be executed, instead deployment artifacts will be printed to the console")
	deployCmd.Flags().StringP(envFilePathFlag, "e", "", "Environment file for the application")
	deployCmd.Flags().StringP(configPathFlag, "c", "", "Config file for the application")
	deployCmd.Flags().BoolP(forceBuild, "f", false, "If set to true, the build will be forced even if the image already exists")
	deployCmd.Flags().StringP(manifestPathFlag, "m", "", "Manifest file for the application")

	deployCmd.MarkFlagRequired(manifestPathFlag)
}

func runDeploy(ctx context.Context, manifestPath string, envFilePath string, agentConfigPath string, platform string, dryRun bool, deleteBuildFolders bool, forceBuild bool, baseImage string, deploymentOption *string) error {
	log := zerolog.Ctx(ctx)

	envVarValues, err := manifest.LoadEnvVars(envFilePath)
	if err != nil {
		return err
	}

	agsb := manifest.NewAgentSpecBuilder()
	err = agsb.BuildAgentSpec(ctx, manifestPath, "", deploymentOption, envVarValues)
	if err != nil {
		return err
	}

	hostStorageFolder, err := getHostStorage()
	if err != nil {
		return err
	}

	// merge default agent config with user provided config
	agentConfig, err := config.GenerateDefaultConfig(agsb.AgentSpecs, platform, agsb.DeploymentName)
	if err != nil {
		return fmt.Errorf("failed to generate default agent config: %v", err)
	}

	if agentConfigPath != "" {
		// if no config is provided, generate a default one
		userConfig, err := config.LoadConfig(agentConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load user config: %v", err)
		}
		agentConfig = config.MergeConfigs(agentConfig, userConfig)
	}

	// load spec from config
	agsb.LoadFromConfig(agentConfig)

	// run agent builder
	agDeploymentSpecs := make(map[string]internal.AgentDeploymentBuildSpec, len(agsb.AgentSpecs))
	for depName, agentSpec := range agsb.AgentSpecs {
		builder := builder.GetAgentBuilder(agentSpec.Manifest.Deployment.DeploymentOptions[agentSpec.SelectedDeploymentOption], deleteBuildFolders, forceBuild, baseImage)
		agdbSpec, err := builder.Build(ctx, agentSpec)
		if err != nil {
			return fmt.Errorf("failed to build agent: %v", err)
		}
		agDeploymentSpecs[depName] = agdbSpec
	}

	// run deployment of agent(s)
	runner := platforms.GetPlatformRunner(platform, hostStorageFolder)

	afs, err := runner.Deploy(ctx, agsb.DeploymentName, agDeploymentSpecs, agsb.Dependencies, 0, dryRun)
	if err != nil {
		return fmt.Errorf("failed to deploy agent: %v", err)
	}
	if dryRun {
		log.Info().Msg(string(afs))
	}
	return nil
}

func getHostStorage() (string, error) {
	//TODO get this from a command line options or env variable
	hostStorageFolder := os.Getenv("HOST_STORAGE_FOLDER")
	if hostStorageFolder == "" {
		homeDir, err := util.GetHomeDir()
		if err != nil {
			return "", errors.New("failed to get home directory")
		}

		hostStorageFolder = path.Join(homeDir, ".wfsm")
		// make sure the folder exists
		if _, err := os.Stat(hostStorageFolder); os.IsNotExist(err) {
			if err := os.Mkdir(hostStorageFolder, 0755); err != nil {
				return "", errors.New("failed to create host storage folder")
			}
		}
	}
	return hostStorageFolder, nil
}
