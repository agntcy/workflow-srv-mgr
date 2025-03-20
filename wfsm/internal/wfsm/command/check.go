package command

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/cisco-eti/wfsm/internal/executor"
)

const (
	checkLongHelp = `
This command checks the prerequisites on the host executing the wfsm command, 
particularly for the docker and docker-compose comands
		
Examples:
- Check the prerequisites for the host:
	wfsm check
`

	verboseChecksFlag string = "true"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks the prerequisites for the command",
	Long:  checkLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool(verboseChecksFlag)
		logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		if verbose {
			logger = logger.Level(zerolog.DebugLevel)
		} else {
			logger = logger.Level(zerolog.InfoLevel)
		}

		logger.Info().Msg("Checking prerequisites for the command...")
		err := runChecks(verbose, logger)
		if err != nil {
			logger.Error().Msg("Checking prerequisites failed")
			return fmt.Errorf(CmdErrorHelpText, err)
		}
		logger.Info().Msg("Checking prerequisites check passed")
		return nil
	},
}

func init() {
	checkCmd.Flags().BoolP(verboseChecksFlag, "v", false, "Output verbose logs for the checks")
}

func runChecks(verbose bool, logger zerolog.Logger) error {
	executorService := executor.NewExecutorService(logger)
	err := checkDocker(executorService)
	if err != nil {
		return err
	}
	return nil
}

func checkDocker(executorService executor.Executor) error {
	_, err := executorService.Execute(context.Background(), executor.Command{
		WorkDir: os.TempDir(),
		Command: "docker",
		Args:    []string{"info"},
	})

	if err != nil {
		return err
	}
	return nil
}
