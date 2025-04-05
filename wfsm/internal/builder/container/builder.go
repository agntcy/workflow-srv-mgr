// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
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
	dockerDeployment := inputSpec.Manifest.Deployment.DeploymentOptions[inputSpec.SelectedDeploymentOption].DockerDeployment
	return internal.AgentDeploymentBuildSpec{
		AgentSpec:   inputSpec,
		Image:       dockerDeployment.Image,
		ServiceName: inputSpec.DeploymentName,
	}, nil
}
