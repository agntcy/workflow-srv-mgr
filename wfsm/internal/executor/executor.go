// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package executor

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/rs/zerolog"
)

// Command is a struct that contains the information of the command that you want to run.
type Command struct {
	WorkDir string
	Command string
	Args    []string
}

func (c Command) String() string {
	return fmt.Sprintf("%s %v", c.Command, c.Args)
}

// Executor is the interface that wraps the Execute method.
type Executor interface {
	Execute(ctx context.Context, command Command) (string, error)
}

type executorService struct {
	logger zerolog.Logger
}

func NewExecutorService(logger zerolog.Logger) Executor {
	return executorService{
		logger: logger,
	}
}

func (e executorService) Execute(ctx context.Context, command Command) (string, error) {
	// define the command that you want to run
	cmd := exec.Command(command.Command, command.Args...)
	// specify the working directory of the command
	cmd.Dir = command.WorkDir
	// define the process standard output
	e.logger.Info().Msgf("excuting `%s` ...", cmd.String())
	output, err := cmd.CombinedOutput() // Run the command
	if err != nil {
		e.logger.Error().Msgf("command `%s` failed. %s", cmd.String(), err)
		e.logger.Debug().Msgf("output: %s", output)
		return string(output), err
	}
	e.logger.Info().Msgf("command `%s` succeeded", cmd.String())

	return string(output), nil
}
