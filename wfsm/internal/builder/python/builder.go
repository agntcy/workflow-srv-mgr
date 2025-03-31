// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package python

import (
	"context"
	"fmt"
	"strings"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/builder/python/source"

	"github.com/rs/zerolog"
)

const AgentImage = "agntcy/wfsm"

// builder implementation of AgentDeployer
type pyBuilder struct {
	baseImage          string
	deleteBuildFolders bool
	forceBuild         bool
}

func NewPythonAgentBuilder(baseImage string, deleteBuildFolders bool, forceBuild bool) internal.AgentDeploymentBuilder {
	return &pyBuilder{
		baseImage:          baseImage,
		deleteBuildFolders: deleteBuildFolders,
		forceBuild:         forceBuild,
	}
}

func (b *pyBuilder) Build(ctx context.Context, inputSpec internal.AgentSpec) (internal.AgentDeploymentBuildSpec, error) {
	log := zerolog.Ctx(ctx)

	deploymentSpec := internal.AgentDeploymentBuildSpec{AgentSpec: inputSpec}

	deployment := inputSpec.Manifest.Deployment.DeploymentOptions[inputSpec.SelectedDeploymentOption]
	agSrc, err := source.GetAgentSource(deployment.SourceCodeDeployment)
	if err != nil {
		return deploymentSpec, fmt.Errorf("failed to get agent source: %v", err)
	}

	imageName := strings.Join([]string{AgentImage, inputSpec.Manifest.Metadata.Ref.Name}, "-")
	imgNameWithTag, err := EnsureContainerImage(ctx, imageName, agSrc, b.deleteBuildFolders, b.forceBuild, b.baseImage)
	if err != nil {
		return deploymentSpec, fmt.Errorf("failed to get/build container image: %v", err)
	}

	log.Debug().Msg(fmt.Sprintf("agent image: %s", imgNameWithTag))
	deploymentSpec.Image = imgNameWithTag
	deploymentSpec.ServiceName = inputSpec.DeploymentName
	return deploymentSpec, nil
}
