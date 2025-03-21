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
	deleteBuildFolders bool
}

func NewPythonAgentBuilder(deleteBuildFolders bool) internal.AgentDeploymentBuilder {
	return &pyBuilder{
		deleteBuildFolders: deleteBuildFolders,
	}
}

func (b *pyBuilder) Validate(ctx context.Context, inputSpec internal.AgentSpec) error {
	log := zerolog.Ctx(ctx)
	// validate that all required env vars are present in inputSpec.EnvVars
	// validate that SourceCodeDeploymentFrameworkConfig settings are correct
	for _, envVarDefs := range inputSpec.Manifest.Deployment.EnvVars {
		if envVarDefs.GetRequired() {
			if _, ok := inputSpec.EnvVars[envVarDefs.GetName()]; !ok {
				if envVarDefs.HasDefaultValue() {
					log.Warn().Msgf("agent %s config is missing required env var %s, using default value %s", inputSpec.DeploymentName, envVarDefs.GetName(), envVarDefs.GetDefaultValue())
					inputSpec.EnvVars[envVarDefs.GetName()] = envVarDefs.GetDefaultValue()
				} else {
					return fmt.Errorf("missing required env var %s", envVarDefs.GetName())
				}
			}
		}
	}
	return nil
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
	imgNameWithTag, _, err := EnsureContainerImage(ctx, imageName, agSrc, b.deleteBuildFolders)
	if err != nil {
		return deploymentSpec, fmt.Errorf("failed to get/build container image: %v", err)
	}

	log.Debug().Msg(fmt.Sprintf("agent image: %s", imgNameWithTag))
	deploymentSpec.Image = imgNameWithTag
	deploymentSpec.ServiceName = inputSpec.DeploymentName
	return deploymentSpec, nil
}
