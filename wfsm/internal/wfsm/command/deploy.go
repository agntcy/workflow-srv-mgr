// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/cisco-eti/wfsm/internal/platforms"
	"github.com/cisco-eti/wfsm/internal/wfsm/config"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/builder"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/cisco-eti/wfsm/internal/wfsm/manifest"
)

var deployLongHelp = `
This command takes two required flags: --manifestPath path/to/acpManifest

Optional flags:
	--envFilePath path/to/envConfigFile user provided environment file
  --configPath path/to/configFile user provided config file
  --showConfig if true, prints out config (defaults and user provided values merged together)
	--baseImage can be set to determine which base image is used as the workflowserver for the agent.
	--deleteBuildFolders can be set to true or false to determine if the build folders should be deleted after deployment.
	--deploymentOption can be set to determine which deployment option to use from the manifest. It defaults to the first deployment option.
	--dryRun if set to true, the deployment will not be executed, instead deployment artifacts will be printed to the console.
	--forceBuild can be set to true or false to determine if the build should be forced even if the image already exists.
	--platform specify the platform to deploy the agent(s) to [docker, k8s].
  --namespace specify the namespace to deploy the agent(s) to. This is only used for k8s deployments.

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
const showConfigFlag string = "showConfig"
const envFilePathFlag string = "envFilePath"
const forceBuild string = "forceBuild"
const manifestPathFlag string = "manifestPath"
const configPathFlag string = "configPath"

type DeployParams struct {
	ManifestPath       string
	EnvFilePath        string
	AgentConfigPath    string
	Platform           string
	DryRun             bool
	ShowConfig         bool
	DeleteBuildFolders bool
	ForceBuild         bool
	BaseImage          string
	DeploymentOption   *string
}

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
		showConfig, _ := cmd.Flags().GetBool(showConfigFlag)
		envFilePath, _ := cmd.Flags().GetString(envFilePathFlag)
		configPathFlag, _ := cmd.Flags().GetString(configPathFlag)
		forceBuild, _ := cmd.Flags().GetBool(forceBuild)
		manifestPath, _ := cmd.Flags().GetString(manifestPathFlag)
		platform, _ := cmd.Flags().GetString(platformsFlag)

		params := DeployParams{
			ManifestPath:       manifestPath,
			EnvFilePath:        envFilePath,
			AgentConfigPath:    configPathFlag,
			Platform:           platform,
			DryRun:             dryRun,
			ShowConfig:         showConfig,
			DeleteBuildFolders: deleteBuildFolders,
			ForceBuild:         forceBuild,
			BaseImage:          baseImage,
			DeploymentOption:   &deploymentOption,
		}

		err := runDeploy(getContextWithLogger(cmd), params)
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
	deployCmd.Flags().BoolP(dryRunFlag, "r", true, "By default set to true, meaning the deployment artifacts are generated, but not executed")
	deployCmd.Flags().BoolP(showConfigFlag, "s", false, "If true, prints out config (defaults and user provided values merged together)")
	deployCmd.Flags().StringP(envFilePathFlag, "e", "", "User provided environment file")
	deployCmd.Flags().StringP(configPathFlag, "c", "", "User provided config file")
	deployCmd.Flags().BoolP(forceBuild, "f", false, "If set to true, the build will be forced even if the image already exists")
	deployCmd.Flags().StringP(manifestPathFlag, "m", "", "Manifest file for the application")

	deployCmd.MarkFlagRequired(manifestPathFlag)
}

func runDeploy(ctx context.Context, params DeployParams) error {
	log := zerolog.Ctx(ctx)

	agentSpecBuilder := manifest.NewAgentSpecBuilder()
	err := agentSpecBuilder.BuildAgentSpec(ctx, params.ManifestPath, "", params.DeploymentOption, nil)
	if err != nil {
		return err
	}

	hostStorageFolder, err := getHostStorageFolder(agentSpecBuilder.DeploymentName)
	if err != nil {
		return err
	}

	// load env vars from env file
	envFile, err := manifest.LoadEnvVars(params.EnvFilePath)
	if err != nil {
		return err
	}

	// merge default agent config with user provided config
	agentConfig, err := config.GenerateDefaultConfig(agentSpecBuilder.AgentSpecs, params.Platform, agentSpecBuilder.DeploymentName, envFile)
	if err != nil {
		return fmt.Errorf("failed to generate default agent config: %v", err)
	}

	if params.AgentConfigPath != "" {
		// if no config is provided, generate a default one
		userConfig, err := config.LoadConfig(params.AgentConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load user config: %v", err)
		}
		agentConfig = config.MergeConfigs(agentConfig, userConfig, params.Platform)
	}

	if params.ShowConfig {
		err = config.PrintConfig(ctx, agentConfig)
		if err != nil {
			return fmt.Errorf("failed to print config: %v", err)
		}
	}

	// load spec from config
	// env vars are merged with the ones from the manifest and env file and OS env vars
	agentSpecBuilder.LoadFromConfig(ctx, agentConfig, envFile)

	errs := agentSpecBuilder.ValidateEnvVars(ctx)
	if len(errs) > 0 {
		// concatenate errors
		errStr := ""
		for _, err := range errs {
			errStr += fmt.Sprintf("%s\n", err.Error())
		}
		return fmt.Errorf("failed validating env vars: %s", errStr)
	}

	// run agent builder
	agDeploymentSpecs := make(map[string]internal.AgentDeploymentBuildSpec, len(agentSpecBuilder.AgentSpecs))
	for depName, agentSpec := range agentSpecBuilder.AgentSpecs {
		deployment := manifest.GetDeployment(agentSpec.Manifest)
		builder := builder.GetAgentBuilder(deployment.DeploymentOptions[agentSpec.SelectedDeploymentOption],
			params.DeleteBuildFolders, params.ForceBuild, params.BaseImage)
		agdbSpec, err := builder.Build(ctx, agentSpec)
		if err != nil {
			return fmt.Errorf("failed to build agent: %v", err)
		}
		agDeploymentSpecs[depName] = agdbSpec
	}

	// run deployment of agent(s)
	runner := platforms.GetPlatformRunner(params.Platform, hostStorageFolder)

	afs, err := runner.Deploy(ctx, agentSpecBuilder.DeploymentName, agDeploymentSpecs, agentSpecBuilder.Dependencies, params.DryRun)
	if err != nil {
		return fmt.Errorf("failed to deploy agent: %v", err)
	}
	if params.DryRun {
		log.Debug().Msg(string(afs))
	}
	return nil
}

func getHostStorageFolder(deploymentName string) (string, error) {
	hostStorageFolder := os.Getenv("WFSM_HOST_STORAGE_FOLDER")
	if hostStorageFolder == "" {
		homeDir, err := util.GetHomeDir()
		if err != nil {
			return "", errors.New("failed to get home directory")
		}

		hostStorageFolder = path.Join(homeDir, ".wfsm", deploymentName)
		// make sure the folder exists
		if _, err := os.Stat(hostStorageFolder); os.IsNotExist(err) {
			if err := os.MkdirAll(hostStorageFolder, 0755); err != nil {
				return "", errors.New("failed to create host storage folder")
			}
		}

	}
	return hostStorageFolder, nil
}
