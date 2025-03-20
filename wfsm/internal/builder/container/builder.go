package container

import (
	"context"

	"github.com/cisco-eti/wfsm/internal"
)

// builder implementation of AgentDeployer
type builder struct {
}

func NewContainerAgentBuilder() *builder {
	return &builder{}
}

func (b *builder) Build(ctx context.Context, inputSpec internal.AgentSpec) (internal.AgentDeploymentBuildSpec, error) {
	return internal.AgentDeploymentBuildSpec{
		AgentSpec:   inputSpec,
		Image:       getImageName(inputSpec),
		ServiceName: "wfsm",
	}, nil
}

func getImageName(spec internal.AgentSpec) string {
	dopts := spec.Manifest.Deployment.DeploymentOptions[spec.SelectedDeploymentOption]
	return dopts.DockerDeployment.Image
}

func (b *builder) Validate(ctx context.Context, inputSpec internal.AgentSpec) error {
	return nil
}
