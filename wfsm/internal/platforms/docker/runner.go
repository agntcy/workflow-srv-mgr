package docker

import (
	"context"
	"fmt"
	"os"

	"github.com/cisco-eti/wfsm/internal"
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
	dockerCli, err := getDockerCLI(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize docker client: %v", err)
	}
	defer dockerCli.Client().Close()

	backend := compose.NewComposeService(dockerCli)
	err = backend.Remove(ctx, deploymentName, api.RemoveOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (r *runner) Logs(ctx context.Context, deploymentName string, agentNames []string) error {
	dockerCli, err := getDockerCLI(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize docker client: %v", err)
	}
	defer dockerCli.Client().Close()

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

	dockerCli, err := getDockerCLI(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize docker client: %v", err)
	}
	defer dockerCli.Client().Close()

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
