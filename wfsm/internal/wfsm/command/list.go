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
const listError string = "get failed"

// listCmd lists running agent container(s) in a deployment
var listCmd = &cobra.Command{
	Use:   "list --agentDeploymentName agentDeploymentName",
	Short: "List an ACP agents running in the deployment",
	Long:  listLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {

		agentDeploymentName, _ := cmd.Flags().GetString(agentDeploymentNameFlag)

		err := runList(agentDeploymentName)
		if err != nil {
			util.OutputMessage(listFail, err.Error())
			return fmt.Errorf(CmdErrorHelpText, listError)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().StringP(agentDeploymentNameFlag, "n", "docker", "Environment file for the application")
	listCmd.Flags().StringP(platformsFlag, "p", "docker", "Environment file for the application")
}

func runList(agentDeploymentName string) error {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &logger

	ctx := logger.WithContext(context.Background())

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
