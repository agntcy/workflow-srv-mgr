package container

import (
	"context"

	"github.com/cisco-eti/wfsm/internal"
)

// builder implementation of AgentDeployer
type cbuilder struct {
}

func NewContainerAgentBuilder() internal.AgentDeploymentBuilder {
	return &cbuilder{}
}

func (b *cbuilder) Build(ctx context.Context, inputSpec internal.AgentSpec) (internal.AgentDeploymentBuildSpec, error) {
	return internal.AgentDeploymentBuildSpec{
		AgentSpec:   inputSpec,
		Image:       getImageName(inputSpec),
		ServiceName: inputSpec.DeploymentName,
	}, nil
}

func getImageName(spec internal.AgentSpec) string {
	dopts := spec.Manifest.Deployment.DeploymentOptions[spec.SelectedDeploymentOption]
	return dopts.DockerDeployment.Image
}
