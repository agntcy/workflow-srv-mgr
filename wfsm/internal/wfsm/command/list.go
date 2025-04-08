// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	docker "github.com/cisco-eti/wfsm/internal/platforms/docker_compose"
	"github.com/cisco-eti/wfsm/internal/util"
)

var listLongHelp = `
This command takes one required flag: --agentDeploymentName <agentDeploymentName>  
Agent deployment name is the name of the agent in the manifest file.
                                      
Optional flags:
	--platform specify the platform to deploy the agent(s) to. Currently only 'docker' is supported.
		
Examples:
- List all running agent containers in 'emailreviewer' deployment:
	wfsm list emailreviewer
`

const listFail = "List Status: Failed - %s"
const listError string = "list failed"

// listCmd lists running agent container(s) in a deployment
var listCmd = &cobra.Command{
	Use:   "list --agentDeploymentName agentDeploymentName",
	Short: "List an ACP agents running in the deployment",
	Long:  listLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {

		agentDeploymentName, _ := cmd.Flags().GetString(agentDeploymentNameFlag)

		err := runList(getContextWithLogger(cmd), agentDeploymentName)
		if err != nil {
			util.OutputMessage(listFail, err.Error())
			return fmt.Errorf(CmdErrorHelpText, listError)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().StringP(agentDeploymentNameFlag, "n", "", "The name of the agent")
	listCmd.Flags().StringP(platformsFlag, "p", "docker", "The deployment target platform")
	listCmd.MarkFlagRequired(agentDeploymentNameFlag)
}

func runList(ctx context.Context, agentDeploymentName string) error {

	hostStorageFolder, err := getHostStorage()
	if err != nil {
		return err
	}
	runner := docker.NewDockerComposeRunner(hostStorageFolder)

	err = runner.List(ctx, agentDeploymentName)
	if err != nil {
		return fmt.Errorf("failed to stop agent deployment: %v", err)
	}

	return nil
}
