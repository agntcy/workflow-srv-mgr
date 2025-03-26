// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package command

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type WorkflowServerManager interface {
	Execute() error
}

type workflowServerManager struct {
	version string
}

var cliToolDescription = `
ACP Workflow Server Manager Tool

Wraps an agent into a web server and exposes the agent functionality through ACP.
It also provides commands for managing existing deployments and cleanup tasks
`

const (
	ReadTheDocsText   = "please take a look at the documentation."
	CmdErrorHelpText  = "%s.\n\nFor additional help, " + ReadTheDocsText
	verboseChecksFlag = "verbose"
)

// NewRootCmd constructs a base command object
func newRootCmd(version string) *cobra.Command {

	var rootCmd = &cobra.Command{
		Use:           "wfsm",
		Short:         "Workflow Server Manager",
		Long:          cliToolDescription,
		SilenceUsage:  true,
		SilenceErrors: false,
		Version:       version,
	}

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gbear.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolP(verboseChecksFlag, "v", false, "Output verbose logs for the checks")

	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(logsCmd)

	return rootCmd
}

func (c workflowServerManager) Execute() error {
	return newRootCmd(c.version).Execute()
}

func NewWorkflowServerManager(version string) WorkflowServerManager {
	return workflowServerManager{version: version}
}

func getContextWithLogger(cmd *cobra.Command) context.Context {
	verbose, _ := cmd.Flags().GetBool(verboseChecksFlag)
	logger := setDefaultContextLogger(verbose)
	return logger.WithContext(context.Background())
}

func setDefaultContextLogger(verbose bool) zerolog.Logger {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	if verbose {
		logger = logger.Level(zerolog.DebugLevel)
	} else {
		logger = logger.Level(zerolog.InfoLevel)
	}
	zerolog.DefaultContextLogger = &logger
	return logger
}
