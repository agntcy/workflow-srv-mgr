// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/cisco-eti/wfsm/internal/wfsm/config"
	"github.com/cisco-eti/wfsm/manifests"
	"github.com/rs/zerolog"
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

// BuildAgentSpec builds the agent spec from the given manifest recursively for all dependencies of the agent.
// Environment values are merged from the manifest and the env var values passed in.
func (a *AgentSpecBuilder) BuildAgentSpec(ctx context.Context, manifestPath string, deploymentName string, selectedDeploymentOption *string, envVarValues *manifests.EnvVarValues) error {

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

	if envVarValues == nil {
		envVarValues = &manifests.EnvVarValues{
			Values: make(map[string]string),
		}
	}

	agentSpec := internal.AgentSpec{
		DeploymentName:           deploymentName,
		Manifest:                 manifest,
		SelectedDeploymentOption: selectedDeploymentOptionIdx,
		EnvVars:                  envVarValues.Values,
		ManifestPath:             manifestPath,
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
			dependency.EnvVarValues = mergeEnvVarValues(dependency.EnvVarValues, *envVarValues, dependency.Name)

			normalizedManifestPath, nErr := a.NormalizeDependencyRef(manifestPath, *dependency.Ref.Url)
			if nErr != nil {
				return fmt.Errorf("failed to normalize manifest path for dependent agent: %s", nErr)
			}

			if err = a.BuildAgentSpec(ctx, normalizedManifestPath, dependency.Name, dependency.DeploymentOption, dependency.EnvVarValues); err != nil {
				return fmt.Errorf("failed building spec for dependent agent: %s", err)
			}
		}
		a.Dependencies[deploymentName] = depNames
	}
	return nil
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

func (a *AgentSpecBuilder) LoadFromConfig(ctx context.Context, configFile config.ConfigFile, envFile map[string]string) {
	for agentName, agentSpec := range a.AgentSpecs {
		agentConfig := configFile.Config[agentName]
		agentSpec.AgentID = agentConfig.ID
		agentSpec.ApiKey = agentConfig.APIKey
		agentSpec.Port = agentConfig.Port
		agentSpec.K8sConfig = agentConfig.K8sConfig

		// configure env vars in order or precedence
		localEnvVars := getLocalEnvs()
		// set declared env vars from local env
		setDeclaredEnvVars(agentSpec, localEnvVars)
		// set prefixed env vars from local env
		setPrefixedEnvVars(agentSpec, localEnvVars)

		// set declared env vars from env file
		setDeclaredEnvVars(agentSpec, envFile)
		// set prefixed env vars from env file
		setPrefixedEnvVars(agentSpec, envFile)

		// set env vars from config
		agentSpec.EnvVars = mergeMaps(agentSpec.EnvVars, agentConfig.EnvVars)

		setDefaultsForEnvVars(ctx, agentSpec)

		a.AgentSpecs[agentName] = agentSpec
	}
}

func (a *AgentSpecBuilder) ValidateEnvVars(ctx context.Context) []error {
	errs := make([]error, 0)
	for _, spec := range a.AgentSpecs {
		if err := validateAgentEnvVars(ctx, spec); err != nil {
			errs = append(errs, validateAgentEnvVars(ctx, spec))
		}
	}
	return errs
}

func getLocalEnvs() map[string]string {
	// Get the environment variables
	envVars := os.Environ()
	envVarsMap := make(map[string]string, len(envVars))

	// Print the environment variables
	for _, env := range envVars {
		// Split the environment variable into key and value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envVarsMap[key] = value
		}
	}
	return envVarsMap
}

// set env vars for the agent spec which are prefixed with the agent name
func setPrefixedEnvVars(inputSpec internal.AgentSpec, envVars map[string]string) {
	agentPrefix := util.CalculateEnvVarPrefix(inputSpec.DeploymentName)
	for key, value := range envVars {
		if strings.HasPrefix(key, agentPrefix) {
			envVarName := strings.TrimPrefix(key, agentPrefix)
			inputSpec.EnvVars[envVarName] = value
		}
	}
}

// set env vars for the agent spec which are declared in the agent manifest
func setDeclaredEnvVars(inputSpec internal.AgentSpec, envFile map[string]string) {
	for _, envVarDefs := range inputSpec.Manifest.Deployment.EnvVars {
		if value := getEnvVarValue(envVarDefs.GetName(), envFile); value != "" {
			inputSpec.EnvVars[envVarDefs.GetName()] = value
		}
	}
}

func getEnvVarValue(envVarName string, envFile map[string]string) string {
	if value, ok := envFile[envVarName]; ok {
		return value
	}
	if value := os.Getenv(envVarName); value != "" {
		return value
	}
	return ""
}

func setDefaultsForEnvVars(ctx context.Context, inputSpec internal.AgentSpec) {
	//log := zerolog.Ctx(ctx)
	for _, envVarDefs := range inputSpec.Manifest.Deployment.EnvVars {
		if inputSpec.EnvVars[envVarDefs.GetName()] == "" && envVarDefs.HasDefaultValue() {
			inputSpec.EnvVars[envVarDefs.GetName()] = envVarDefs.GetDefaultValue()
		}
	}
}

func validateAgentEnvVars(ctx context.Context, inputSpec internal.AgentSpec) error {
	log := zerolog.Ctx(ctx)
	// validate that all required env vars are present in inputSpec.EnvVars
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

func LoadEnvVars(envFilePath string) (map[string]string, error) {
	envVars := make(map[string]string)
	if envFilePath == "" {
		return envVars, nil
	}

	file, err := os.Open(envFilePath)
	if err != nil {
		return nil, errors.New("failed to open env file")
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines or comments
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		// Split the line into key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line in env file: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envVars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("failed to read env file")
	}

	return envVars, nil
}
