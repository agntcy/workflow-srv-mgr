package builder

import (
	"context"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/builder/container"
	"github.com/cisco-eti/wfsm/internal/builder/python"
	"github.com/cisco-eti/wfsm/manifests"
)

// AgentDeploymentBuilder interface with deploy method
type AgentDeploymentBuilder interface {
	Build(ctx context.Context, inputSpec internal.AgentSpec) (internal.AgentDeploymentBuildSpec, error)
	Validate(ctx context.Context, inputSpec internal.AgentSpec) error
}

func GetAgentBuilder(deploymentOption manifests.AgentDeploymentDeploymentOptionsInner, deleteBuildFolders bool) AgentDeploymentBuilder {
	if deploymentOption.DockerDeployment != nil {
		return container.NewContainerAgentBuilder()
	} else if deploymentOption.SourceCodeDeployment != nil {
		return python.NewPythonAgentBuilder(deleteBuildFolders)
	}
	return nil
}
