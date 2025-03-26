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

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/platforms/docker_compose"

	"github.com/cisco-eti/wfsm/internal/builder"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/cisco-eti/wfsm/internal/wfsm/manifest"
)

var deployLongHelp = `
This command takes two required flags: --manifestPath path/to/acpManifest
                                       --envFilePath path/to/envConfigFile
Optional flags:
	--platform specify the platform to deploy the agent(s) to. Currently only 'docker' is supported.
	--dryRun if set to true, the deployment will not be executed, instead deployment artifacts will be printed to the console.
	--deleteBuildFolders can be set to true or false to determine if the build folders should be deleted after deployment.

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
const deployError string = "get failed"

const baseImageFlag string = "baseImage"
const deleteBuildFoldersFlag string = "deleteBuildFolders"
const dryRunFlag string = "dryRun"
const envFilePathFlag string = "envFilePath"
const manifestPathFlag string = "manifestPath"
const platformsFlag string = "platform"

// deployCmd represents the image build and run docker commands
var deployCmd = &cobra.Command{
	Use:   "deploy --manifestPath path/to/acpManifest --envFilePath path/to/envFile",
	Short: "Build an ACP agent",
	Long:  deployLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {

		baseImage, _ := cmd.Flags().GetString(baseImageFlag)
		deleteBuildFolders, _ := cmd.Flags().GetBool(deleteBuildFoldersFlag)
		dryRun, _ := cmd.Flags().GetBool(dryRunFlag)
		envFilePath, _ := cmd.Flags().GetString(envFilePathFlag)
		manifestPath, _ := cmd.Flags().GetString(manifestPathFlag)
		platform, _ := cmd.Flags().GetString(platformsFlag)

		err := runDeploy(getContextWithLogger(cmd), manifestPath, envFilePath, platform, dryRun, deleteBuildFolders, baseImage)
		if err != nil {
			util.OutputMessage(deployFail, err.Error())
			return fmt.Errorf(CmdErrorHelpText, deployError)
		}
		return nil
	},
}

func init() {
	deployCmd.Flags().StringP(baseImageFlag, "b", "ghcr.io/agntcy/acp/wfsrv:v0.2.0-dev.1", "Base image to be used as the workflowserver for the agent")
	deployCmd.Flags().StringP(envFilePathFlag, "e", "", "Environment file for the application")
	deployCmd.Flags().StringP(manifestPathFlag, "m", "", "Manifest file for the application")

	deployCmd.Flags().BoolP(deleteBuildFoldersFlag, "d", true, "Delete build folders after deployment")
	deployCmd.Flags().BoolP(dryRunFlag, "r", false, "If set to true, the deployment will not be executed, instead deployment artifacts will be printed to the console")
	deployCmd.Flags().StringP(platformsFlag, "p", "docker", "Environment file for the application")

	deployCmd.MarkFlagRequired(envFilePathFlag)
	deployCmd.MarkFlagRequired(manifestPathFlag)

}

func runDeploy(ctx context.Context, manifestPath string, envFilePath string, platform string, dryRun bool, deleteBuildFolders bool, baseImage string) error {
	log := zerolog.Ctx(ctx)

	envVarValues, err := manifest.LoadEnvVars(envFilePath)
	if err != nil {
		return err
	}

	agsb := manifest.NewAgentSpecBuilder()
	err = agsb.BuildAgentSpec(ctx, manifestPath, "", nil, envVarValues)
	if err != nil {
		return err
	}

	agDeploymentSpecs := make(map[string]internal.AgentDeploymentBuildSpec, len(agsb.AgentSpecs))

	for depName, agentSpec := range agsb.AgentSpecs {
		builder := builder.GetAgentBuilder(agentSpec.Manifest.Deployment.DeploymentOptions[agentSpec.SelectedDeploymentOption], deleteBuildFolders, baseImage)
		agdbSpec, err := builder.Build(ctx, agentSpec)
		if err != nil {
			return fmt.Errorf("failed to build agent: %v", err)
		}
		agDeploymentSpecs[depName] = agdbSpec
	}

	// run deployment of agent(s)
	hostStorageFolder, err := getHostStorage()
	if err != nil {
		return err
	}
	runner := docker.NewDockerComposeRunner(hostStorageFolder)

	afs, err := runner.Deploy(ctx, agsb.DeploymentName, agDeploymentSpecs, agsb.Dependencies, 0, dryRun)
	if err != nil {
		return fmt.Errorf("failed to deploy agent: %v", err)
	}
	log.Debug().Msg(string(afs))

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
