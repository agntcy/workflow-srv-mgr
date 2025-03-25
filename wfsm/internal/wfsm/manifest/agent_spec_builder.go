package manifest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"

	"github.com/cisco-eti/wfsm/manifests"
)

type AgentSpecBuilder struct {
	//Name of the main agent
	DeploymentName string
	AgentSpecs     map[string]internal.AgentSpec
	Dependencies   map[string][]string
}

func NewAgentSpecBuilder() *AgentSpecBuilder {
	return &AgentSpecBuilder{
		AgentSpecs:   make(map[string]internal.AgentSpec),
		Dependencies: make(map[string][]string),
	}
}

func (a *AgentSpecBuilder) BuildAgentSpec(ctx context.Context, manifestPath string, deploymentName string, selectedDeploymentOption *string, envVarValues manifests.EnvVarValues) error {

	manifestSvc, err := NewManifestService(manifestPath)
	if err != nil {
		return err
	}

	if manifestSvc.Validate() != nil {
		return fmt.Errorf("manifest validation failed: %s", err)
	}

	manifest := manifestSvc.GetManifest()
	if deploymentName == "" {
		deploymentName = manifest.Metadata.Ref.Name
		a.DeploymentName = deploymentName
	}

	// check deployment name is unique among dependencies
	if _, ok := a.AgentSpecs[deploymentName]; ok {
		return fmt.Errorf("agent deployment name must be unique: %s", deploymentName)
	}

	selectedDeploymentOptionIdx := 0
	if selectedDeploymentOption != nil {
		selectedDeploymentOptionIdx = getSelectedDeploymentOptionIdx(manifest.Deployment.DeploymentOptions, *selectedDeploymentOption)
	}

	agentSpec := internal.AgentSpec{
		DeploymentName:           deploymentName,
		Manifest:                 manifest,
		SelectedDeploymentOption: selectedDeploymentOptionIdx,
		EnvVars:                  envVarValues.Values,
		AgentID:                  uuid.NewString(),
		ApiKey:                   uuid.NewString(),
	}
	a.AgentSpecs[deploymentName] = agentSpec

	if len(manifest.Deployment.Dependencies) > 0 {
		depNames := make([]string, 0, len(manifest.Deployment.Dependencies))
		for _, dependency := range manifest.Deployment.Dependencies {
			depNames = append(depNames, dependency.Name)

			if dependency.Ref.Url == nil {
				return fmt.Errorf("ref url is required for dependency: %s", dependency.Name)
			}

			// merge env vars
			dependency.EnvVarValues = mergeEnvVarValues(dependency.EnvVarValues, envVarValues, dependency.Name)
			// validate required env vars
			err := validateEnvVarValues(ctx, agentSpec)
			if err != nil {
				return fmt.Errorf("failed validating env vars for %s agent: %s", dependency.Name, err)
			}

			err = a.BuildAgentSpec(ctx, *dependency.Ref.Url, dependency.Name, dependency.DeploymentOption, *dependency.EnvVarValues)
			if err != nil {
				return fmt.Errorf("failed building spec for dependent agent: %s", err)
			}
		}
		a.Dependencies[deploymentName] = depNames
	}
	return nil
}

func validateEnvVarValues(ctx context.Context, inputSpec internal.AgentSpec) error {
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

func mergeMaps(dest map[string]string, src map[string]string) map[string]string {
	if dest == nil {
		dest = make(map[string]string)
	}
	for key, value := range src {
		dest[key] = value
	}
	return dest
}

func mergeEnvVarValues(dest *manifests.EnvVarValues, src manifests.EnvVarValues, dependencyName string) *manifests.EnvVarValues {
	if dest == nil {
		dest = &manifests.EnvVarValues{}
	}

	for _, depEnv := range src.Dependencies {
		if depEnv.GetName() == dependencyName {
			// merge env var values for dependencyName
			dest.Values = mergeMaps(dest.Values, depEnv.Values)
			// merge env vars of dependencies of dependencyName
			dest.Dependencies = mergeDepEnvVarValues(dest.Dependencies, depEnv.Dependencies)
		}
	}
	return dest
}

func mergeDepEnvVarValues(dest []manifests.EnvVarValues, src []manifests.EnvVarValues) []manifests.EnvVarValues {
	if src == nil {
		return dest
	}
	if dest == nil {
		dest = make([]manifests.EnvVarValues, 0, len(src))
	}
	for _, depEnv := range src {
		dest = append(dest, *mergeEnvVarValues(&depEnv, depEnv, depEnv.GetName()))
	}
	return dest
}

func getSelectedDeploymentOptionIdx(options []manifests.AgentDeploymentDeploymentOptionsInner, option string) int {
	for i, opt := range options {
		if opt.SourceCodeDeployment != nil &&
			opt.SourceCodeDeployment.Name != nil &&
			*opt.SourceCodeDeployment.Name == option {
			return i
		}
		if opt.DockerDeployment != nil &&
			opt.DockerDeployment.Name != nil &&
			*opt.DockerDeployment.Name == option {
			return i
		}
	}
	return 0
}

func LoadEnvVars(envFilePath string) (manifests.EnvVarValues, error) {
	file, err := os.Open(envFilePath)
	if err != nil {
		return manifests.EnvVarValues{}, errors.New("failed to open env file")
	}
	defer file.Close()

	// Read the file into a byte slice
	byteSlice, err := io.ReadAll(file)
	if err != nil {
		return manifests.EnvVarValues{}, errors.New("failed to read env file")
	}

	envVarValues := manifests.EnvVarValues{}

	if err := yaml.Unmarshal(byteSlice, &envVarValues); err != nil {
		return manifests.EnvVarValues{}, err
	}

	return envVarValues, nil
}
