// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package k8s

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/cisco-eti/wfsm/assets"
	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

const ManifestCheckSum = "org.agntcy.wfsm.manifest"
const ServerPort = "8000/tcp"
const APIHost = "0.0.0.0"
const APIPort = "8000"

// Deploy if externalPort is 0, will try to find the port of already running container or find next available port
func (r *runner) Deploy(ctx context.Context,
	mainAgentName string,
	agentDeploymentSpecs map[string]internal.AgentDeploymentBuildSpec,
	dependencies map[string][]string,
	externalPort int,
	dryRun bool) (internal.DeploymentArtifact, error) {

	log := zerolog.Ctx(ctx)

	// insert api keys, agent IDs and service names as host into the deployment specs
	for agName, deps := range dependencies {
		agSpec := agentDeploymentSpecs[agName]
		for _, depName := range deps {
			depAgPrefix := util.CalculateEnvVarPrefix(depName)
			depSpec := agentDeploymentSpecs[depName]
			agSpec.EnvVars[depAgPrefix+"API_KEY"] = fmt.Sprintf("{\"x-api-key\": \"%s\"}", depSpec.ApiKey)
			agSpec.EnvVars[depAgPrefix+"ID"] = depSpec.AgentID
			agSpec.EnvVars[depAgPrefix+"ENDPOINT"] = fmt.Sprintf("http://%s:%s", depSpec.ServiceName, APIPort)
		}
	}

	agentValueConfigs := make([]AgentValues, 0, len(agentDeploymentSpecs))

	// only the main agent will be exposed to the outside world
	mainAgentSpec := agentDeploymentSpecs[mainAgentName]

	//port := externalPort
	//if port == 0 {
	//	port, err = r.getMainAgentPublicPort(ctx, cli, mainAgentName, mainAgentSpec)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	mainAgentID := mainAgentSpec.AgentID
	mainAgentAPiKey := mainAgentSpec.ApiKey
	sc, err := r.createAgentValuesConfig(mainAgentName, mainAgentSpec)
	sc.Service = Service{
		Type: "LoadBalancer",
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create service config: %v", err)
	}
	agentValueConfigs = append(agentValueConfigs, *sc)
	delete(agentDeploymentSpecs, mainAgentName)

	// Uncompress helm chart to r.hostStorageFolder
	if err := util.UntarGzFile(assets.AgentChart, r.hostStorageFolder); err != nil {
		return nil, fmt.Errorf("failed to uncompress tar.gz file: %v", err)
	}
	chartUrl := path.Join(r.hostStorageFolder, "agent")
	log.Info().Msgf("Agent helm chart available at: %s", chartUrl)

	// generate service configs for dependencies
	for _, deploymentSpec := range agentDeploymentSpecs {
		sc, err := r.createAgentValuesConfig(mainAgentName, deploymentSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to create service config: %v", err)
		}
		agentValueConfigs = append(agentValueConfigs, *sc)
	}

	chartValues := ChartValues{
		Agents: agentValueConfigs,
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(chartValues)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chart values: %v", err)
	}

	// TOD remove Print YAML
	fmt.Println(string(yamlData))

	valuesFilePath := path.Join(r.hostStorageFolder, fmt.Sprintf("values-%s.yaml", mainAgentName))
	err = os.WriteFile(valuesFilePath, yamlData, util.OwnerCanReadWrite)
	if err != nil {
		return nil, fmt.Errorf("failed to write values file: %v", err)
	}

	log.Info().Msgf("values file generated at: %s", valuesFilePath)

	if dryRun {
		return yamlData, nil
	}

	deployer := NewHelmDeployer()
	err = deployer.DeployChart(ctx, mainAgentName, chartUrl, "default", yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy chart: %v", err)
	}

	//TODO get load balancer IP

	log.Info().Msg("---------------------------------------------------------------------")
	log.Info().Msgf("ACP agent helm chart release name: %s", mainAgentName)
	log.Info().Msgf("ACP agent running in container: %s, listening for ACP requests on: http://<loadbalancerAddress>:%d", mainAgentName, 8000)
	log.Info().Msgf("Agent ID: %s", mainAgentID)
	log.Info().Msgf("API Key: %s", mainAgentAPiKey)
	log.Info().Msgf("API Docs: http://<loadbalancerAddress>:%d/agents/%s/docs", 8000, mainAgentID)
	log.Info().Msg("---------------------------------------------------------------------\n\n\n")

	return nil, nil
}

func (r *runner) createAgentValuesConfig(projectName string, deploymentSpec internal.AgentDeploymentBuildSpec) (*AgentValues, error) {
	envVars := deploymentSpec.EnvVars

	envVars["API_HOST"] = APIHost
	envVars["API_PORT"] = APIPort
	envVars["AGENT_ID"] = deploymentSpec.AgentID

	secretEnvVars := make(map[string]string, 10)
	secretEnvVars["API_KEY"] = deploymentSpec.ApiKey

	imageRepo, tag := util.SplitImageName(deploymentSpec.Image)

	agentValues := &AgentValues{
		Name: util.NormalizeAgentName(deploymentSpec.ServiceName),
		Image: Image{
			Repository: imageRepo,
			Tag:        tag,
		},
		//Labels:             deploymentSpec.Labels,
		Env:        convertEnvVars(envVars),
		SecretEnvs: convertEnvVars(secretEnvVars),
		VolumePath: "/opt/storage",
		//TODO setup ports
		ExternalPort: 8000, //ServerPort,
		InternalPort: 8000, //APIPort,
		//Service: Service{
		//	Type:        deploymentSpec.ServiceType,
		//	Labels:      deploymentSpec.ServiceLabels,
		//	Annotations: deploymentSpec.ServiceAnnotations,
		//},
		//StatefulSet: StatefulSet{
		//	Replicas:     deploymentSpec.Replicas,
		//	Labels:       deploymentSpec.StatefulSetLabels,
		//	Annotations:  deploymentSpec.StatefulSetAnnotations,
		//	Resources:    deploymentSpec.Resources,
		//	NodeSelector: deploymentSpec.NodeSelector,
		//	Affinity:     deploymentSpec.Affinity,
		//	Tolerations:  deploymentSpec.Tolerations,
		//},
	}

	return agentValues, nil
}

func convertEnvVars(envVars map[string]string) []EnvVar {
	var result []EnvVar
	for key, value := range envVars {
		result = append(result, EnvVar{
			Name:  key,
			Value: value,
		})
	}
	return result
}
