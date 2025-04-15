// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package docker

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/rs/zerolog"
)

// DockerComposeRunner implementation of AgentDeploymentRunner
type runner struct {
	hostStorageFolder string
}

func NewDockerComposeRunner(hostStorageFolder string) internal.AgentDeploymentRunner {
	return &runner{
		hostStorageFolder: hostStorageFolder,
	}
}

func (r *runner) Remove(ctx context.Context, deploymentName string) error {
	dockerCli, err := util.GetDockerCLI(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize docker client: %v", err)
	}
	defer dockerCli.Client().Close()

	deploymentName = GetProjectName(deploymentName)
	backend := compose.NewComposeService(dockerCli)
	err = backend.Down(ctx, deploymentName, api.DownOptions{
		//Project: project,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *runner) Logs(ctx context.Context, deploymentName string, agentNames []string) error {
	dockerCli, err := util.GetDockerCLI(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize docker client: %v", err)
	}
	defer dockerCli.Client().Close()

	deploymentName = GetProjectName(deploymentName)
	backend := compose.NewComposeService(dockerCli)
	logConsumer := formatter.NewLogConsumer(ctx, os.Stdout, os.Stderr, true, true, true)
	err = backend.Logs(ctx, deploymentName, logConsumer, api.LogOptions{
		Services: agentNames,
		Tail:     "100",
		Follow:   true,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *runner) List(ctx context.Context, deploymentName string) error {
	log := zerolog.Ctx(ctx)

	dockerCli, err := util.GetDockerCLI(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize docker client: %v", err)
	}
	defer dockerCli.Client().Close()

	deploymentName = GetProjectName(deploymentName)
	backend := compose.NewComposeService(dockerCli)
	list, err := backend.Ps(ctx, deploymentName, api.PsOptions{All: true})
	if err != nil {
		return err
	}
	for _, c := range list {
		log.Info().Msg(fmt.Sprintf("agent running in container: %s, status: '%s'", c.Name, c.Status))
	}

	return nil
}

func GetProjectName(name string) string {
	// replace all non-alphanumeric characters with _
	re := regexp.MustCompile(`[^a-z0-9-_]+`)
	return re.ReplaceAllString(strings.ToLower(name), "")
}
