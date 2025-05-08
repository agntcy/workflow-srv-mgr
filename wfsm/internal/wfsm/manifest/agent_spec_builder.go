// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/wfsm/config"

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
	manifestLoader, err := LoaderFactory(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to create manifest loader: %s", err)
	}
	manifestSvc, err := NewManifestService(ctx, manifestLoader)
	if err != nil {
		return err
	}

	err = manifestSvc.Validate()
	if err != nil {
		return fmt.Errorf("manifest validation failed: %s", err)
	}

	selectedDeploymentOptionIdx, err := manifestSvc.GetDeploymentOptionIdx(selectedDeploymentOption)
	if err != nil {
		return err
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

	agentSpec := internal.AgentSpec{
		DeploymentName:           deploymentName,
		Manifest:                 manifest,
		SelectedDeploymentOption: selectedDeploymentOptionIdx,
		EnvVars:                  envVarValues.Values,
		ManifestPath:             manifestPath,
	}
	a.AgentSpecs[deploymentName] = agentSpec

	if len(manifest.Deployment.AgentDeps) > 0 {
		depNames := make([]string, 0, len(manifest.Deployment.AgentDeps))
		for _, dependency := range manifest.Deployment.AgentDeps {
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

			normalizedManifestPath, nErr := a.NormalizeDependencyRef(manifestPath, *dependency.Ref.Url)
			if nErr != nil {
				return fmt.Errorf("failed to normalize manifest path for dependent agent: %s", nErr)
			}

			if err = a.BuildAgentSpec(ctx, normalizedManifestPath, dependency.Name, dependency.DeploymentOption, *dependency.EnvVarValues); err != nil {
				return fmt.Errorf("failed building spec for dependent agent: %s", err)
			}
		}
		a.Dependencies[deploymentName] = depNames
	}
	return nil
}

func (a *AgentSpecBuilder) LoadFromConfig(agentConfig config.ConfigFile) {
	for agentName, config := range agentConfig.Config {
		agentSpec := a.AgentSpecs[agentName]
		agentSpec.AgentID = config.ID
		agentSpec.ApiKey = config.APIKey
		agentSpec.Port = config.Port
		agentSpec.K8sConfig = config.K8sConfig
		a.AgentSpecs[agentName] = agentSpec
	}
}

// NormalizeDependencyRef normalizes the manifest path for the agent spec builder
func (a *AgentSpecBuilder) NormalizeDependencyRef(manifestPath string, dependencyRefPath string) (string, error) {
	parsedRef, err := url.Parse(dependencyRefPath)
	if err != nil {
		return "", err
	}

	if parsedRef.Scheme != "" && parsedRef.Scheme != "file" {
		// the reference is not a local file path
		return dependencyRefPath, nil
	}

	// the reference is a local file path, normalize it
	rawDependencyPath := strings.TrimPrefix(dependencyRefPath, "file://")

	if filepath.IsAbs(rawDependencyPath) {
		// the reference is an absolute path
		return rawDependencyPath, nil
	}

	// the reference is a relative path, resolve it relative to the manifest path
	normalizedPath := filepath.Join(filepath.Dir(manifestPath), rawDependencyPath)

	return filepath.Clean(normalizedPath), nil

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

	for _, depEnv := range src.EnvDeps {
		if depEnv.GetName() == dependencyName {
			// merge env var values for dependencyName
			dest.Values = mergeMaps(dest.Values, depEnv.Values)
			// merge env vars of dependencies of dependencyName
			dest.EnvDeps = mergeDepEnvVarValues(dest.EnvDeps, depEnv.EnvDeps)
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
