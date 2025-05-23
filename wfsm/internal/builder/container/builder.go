// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package container

import (
	"context"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/wfsm/manifest"
)

// builder implementation of AgentDeployer
type cbuilder struct {
}

func NewContainerAgentBuilder() internal.AgentDeploymentBuilder {
	return &cbuilder{}
}

func (b *cbuilder) Build(ctx context.Context, inputSpec internal.AgentSpec) (internal.AgentDeploymentBuildSpec, error) {
	deployment := manifest.GetDeployment(inputSpec.Manifest)
	dockerDeployment := deployment.DeploymentOptions[inputSpec.SelectedDeploymentOption].DockerDeployment
	return internal.AgentDeploymentBuildSpec{
		AgentSpec:   inputSpec,
		Image:       dockerDeployment.Image,
		ServiceName: inputSpec.DeploymentName,
	}, nil
}
