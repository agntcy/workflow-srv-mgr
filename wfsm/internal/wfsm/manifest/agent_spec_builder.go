package manifest

import (
	"fmt"

	"github.com/cisco-eti/wfsm/internal"
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

func (a *AgentSpecBuilder) BuildAgentSpec(manifestPath string, deploymentName string, selectedDeploymentOption *string, envVarValues manifests.EnvVarValues) error {

	manifestSvc, err := NewManifestService(manifestPath)
	if err != nil {
		return err
	}

	if manifestSvc.ValidateDeploymentOptions() != nil {
		return fmt.Errorf("manifest validation failed: %s", err)
	}

	manifest := manifestSvc.GetManifest()
	if deploymentName == "" {
		deploymentName = manifest.Metadata.Ref.Name
		a.DeploymentName = deploymentName
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
	}
	a.AgentSpecs[deploymentName] = agentSpec

	if len(manifest.Deployment.Dependencies) > 0 {
		depNames := make([]string, len(manifest.Deployment.Dependencies), 0)
		for _, dependency := range manifest.Deployment.Dependencies {
			depNames = append(depNames, dependency.Name)

			if dependency.Ref.Url == nil {
				return fmt.Errorf("ref url is required for dependency: %s", dependency.Name)
			}

			dependency.EnvVarValues = mergeEnvVarValues(dependency.EnvVarValues, envVarValues, dependency.Name)

			// merge env vars
			err = a.BuildAgentSpec(*dependency.Ref.Url, dependency.Name, dependency.DeploymentOption, *dependency.EnvVarValues)
			if err != nil {
				return fmt.Errorf("failed building spec for dependent agent: %s", err)
			}
		}
		a.Dependencies[deploymentName] = depNames
	}
	return nil
}

func mergeEnvVarValues(dest *manifests.EnvVarValues, src manifests.EnvVarValues, dependencyName string) *manifests.EnvVarValues {
	if dest == nil {
		dest = &manifests.EnvVarValues{}
	}
	//TODO merge dependencies values
	for _, depEnv := range src.Dependencies {
		if depEnv.GetName() == dependencyName {
			dest.Values = mergeMaps(dest.Values, depEnv.GetValues().Values)
		}
	}
	return dest
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
