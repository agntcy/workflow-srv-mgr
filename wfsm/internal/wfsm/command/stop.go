// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package command

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	docker "github.com/cisco-eti/wfsm/internal/platforms/docker_compose"
	"github.com/cisco-eti/wfsm/internal/util"
)

var stopLongHelp = `
This command takes one required flag: --agentDeploymentName <agentDeploymentName>  
Agent deployment name is the name of the agent in the manifest file.
                                   
Optional flags:
	--platform specify the platform to deploy the agent(s) to. Currently only 'docker' is supported.
		
Examples:
- Stops all running agents in 'emailreviewer' agent deployment:
	wfsm stop emailreviewer
`

const agentDeploymentNameFlag string = "agentDeploymentName"

const stopFail = "Stop Status: Failed - %s"
const stopError string = "stop failed"

// stopCmd stops the deployment of the agent(s)
var stopCmd = &cobra.Command{
	Use:   "stop --agentDeploymentName agentDeploymentName",
	Short: "Stop an ACP agent deployment",
	Long:  stopLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {

		agentDeploymentName, _ := cmd.Flags().GetString(agentDeploymentNameFlag)

		err := runStop(getContextWithLogger(cmd), agentDeploymentName)
		if err != nil {
			util.OutputMessage(stopFail, err.Error())
			return fmt.Errorf(CmdErrorHelpText, stopError)
		}
		return nil
	},
}

func init() {
	stopCmd.Flags().StringP(agentDeploymentNameFlag, "n", "", "The name of the agent")
	stopCmd.Flags().StringP(platformsFlag, "p", "docker", "The deployment target platform")
	stopCmd.MarkFlagRequired(agentDeploymentNameFlag)
}

func runStop(ctx context.Context, agentDeploymentName string) error {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &logger

	// stop deployment of agent(s)

	hostStorageFolder, err := getHostStorage()
	if err != nil {
		return err
	}
	runner := docker.NewDockerComposeRunner(hostStorageFolder)

	err = runner.Remove(ctx, agentDeploymentName)
	if err != nil {
		return fmt.Errorf("failed to stop agent deployment: %v", err)
	}

	return nil
}
