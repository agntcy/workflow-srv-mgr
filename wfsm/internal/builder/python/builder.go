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

	log.Info().Msgf("building image for agent: %s", inputSpec.DeploymentName)
	deploymentSpec := internal.AgentDeploymentBuildSpec{AgentSpec: inputSpec}

	deploymentManifest := inputSpec.Manifest.Extensions[0].Data.Deployment
	deployment := deploymentManifest.DeploymentOptions[inputSpec.SelectedDeploymentOption]
	agSrc, err := source.GetAgentSource(deployment.SourceCodeDeployment, inputSpec.ManifestPath)
	if err != nil {
		return deploymentSpec, fmt.Errorf("failed to get agent source: %v", err)
	}

	imageName := strings.Join([]string{AgentImage, inputSpec.Manifest.Name}, "-")
	imgNameWithTag, err := EnsureContainerImage(ctx, imageName, agSrc, inputSpec, b.deleteBuildFolders, b.forceBuild, b.baseImage)
	if err != nil {
		return deploymentSpec, err
	}

	log.Debug().Msg(fmt.Sprintf("agent image: %s", imgNameWithTag))
	deploymentSpec.Image = imgNameWithTag
	deploymentSpec.ServiceName = inputSpec.DeploymentName
	return deploymentSpec, nil
}
