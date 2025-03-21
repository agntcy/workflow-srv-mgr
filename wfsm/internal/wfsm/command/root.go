package command

import (
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
	ReadTheDocsText  = "please take a look at the documentation."
	CmdErrorHelpText = "%s.\n\nFor additional help, " + ReadTheDocsText
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
