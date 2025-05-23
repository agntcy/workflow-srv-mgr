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

	"github.com/cisco-eti/wfsm/internal/platforms/docker"
	"github.com/cisco-eti/wfsm/internal/util"
)

var logsLongHelp = `
This command takes one required flag: --agentDeploymentName <agentDeploymentName>  
Agent deployment name is the name of the agent in the manifest file.
                                      
Optional flags:
	--platform specify the platform to deploy the agent(s) to. Currently only 'docker' is supported.
		
Examples:
- Shows latest logs of all running agent containers in 'emailreviewer' deployment:
	wfsm logs emailreviewer
`

const logsFail = "Logs Status: Failed - %s"
const logsError string = "logs failed"

// logsCmd show the logs of running deployment of the agent(s)
var logsCmd = &cobra.Command{
	Use:   "logs --agentDeploymentName agentDeploymentName",
	Short: "Show logs of an ACP agent deployment(s)",
	Long:  logsLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {

		agentDeploymentName, _ := cmd.Flags().GetString(agentDeploymentNameFlag)

		err := runLogs(getContextWithLogger(cmd), agentDeploymentName)
		if err != nil {
			util.OutputMessage(logsFail, err.Error())
			return fmt.Errorf(CmdErrorHelpText, logsError)
		}
		return nil
	},
}

func init() {
	logsCmd.Flags().StringP(agentDeploymentNameFlag, "n", "", "The name of the agent")
	logsCmd.Flags().StringP(platformsFlag, "p", "docker", "The deployment target platform")
	logsCmd.MarkFlagRequired(agentDeploymentNameFlag)
}

func runLogs(ctx context.Context, agentDeploymentName string) error {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &logger

	hostStorageFolder, err := getHostStorageFolder(agentDeploymentName)
	if err != nil {
		return err
	}
	runner := docker.NewDockerComposeRunner(hostStorageFolder)

	err = runner.Logs(ctx, agentDeploymentName, []string{})
	if err != nil {
		return fmt.Errorf("failed to stop agent deployment: %v", err)
	}

	return nil
}
