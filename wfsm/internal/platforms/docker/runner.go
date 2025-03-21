package docker

import (
	"context"

	"github.com/cisco-eti/wfsm/internal"
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
	return nil
}

func (r *runner) Logs(ctx context.Context, deploymentName string) error {
	return nil
}

func (r *runner) List(ctx context.Context, deploymentName string) error {
	return nil
}
