// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package k8s

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/cisco-eti/wfsm/assets"
	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ConfigCheckSum = "org.agntcy.wfsm.config.checksum"
const APIHost = "0.0.0.0"

// Deploy generates a Docker compose file from the agent deployment specs and deploys it if dryRun = false
func (r *runner) Deploy(ctx context.Context,
	mainAgentName string,
	agentDeploymentSpecs map[string]internal.AgentDeploymentBuildSpec,
	dependencies map[string][]string,
	dryRun bool) (internal.DeploymentArtifact, error) {

	log := zerolog.Ctx(ctx)
	namespace := getK8sNamespace()

	// insert api keys, agent IDs and service names as host into the deployment specs
	for agName, deps := range dependencies {
		agSpec := agentDeploymentSpecs[agName]
		for _, depName := range deps {
			depAgPrefix := util.CalculateEnvVarPrefix(depName)
			depSpec := agentDeploymentSpecs[depName]
			agSpec.EnvVars[depAgPrefix+"API_KEY"] = fmt.Sprintf("{\"x-api-key\": \"%s\"}", depSpec.ApiKey)
			agSpec.EnvVars[depAgPrefix+"ID"] = depSpec.AgentID
			// service name is the same as the deployment name but should be normalized to k8s standard
			agSpec.EnvVars[depAgPrefix+"ENDPOINT"] = fmt.Sprintf("http://%s:%d", util.NormalizeAgentName(depSpec.ServiceName), depSpec.Port)
		}
	}

	agentValueConfigs := make([]AgentValues, 0, len(agentDeploymentSpecs))

	// only the main agent will be exposed to the outside world
	mainAgentSpec := agentDeploymentSpecs[mainAgentName]

	mainAgentID := mainAgentSpec.AgentID
	mainAgentAPiKey := mainAgentSpec.ApiKey
	sc, err := r.createAgentValuesConfig(mainAgentSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to create service config: %v", err)
	}
	agentValueConfigs = append(agentValueConfigs, *sc)
	delete(agentDeploymentSpecs, mainAgentName)

	// Uncompress helm chart to r.hostStorageFolder
	if err := util.UntarGzFile(assets.AgentChart, r.hostStorageFolder); err != nil {
		return nil, fmt.Errorf("failed to uncompress tar.gz file: %v", err)
	}
	chartUrl := path.Join(r.hostStorageFolder, "charts", "agent")
	log.Info().Msgf("Agent helm chart available at: %s", chartUrl)

	// generate service configs for dependencies
	for _, deploymentSpec := range agentDeploymentSpecs {
		sc, err := r.createAgentValuesConfig(deploymentSpec)
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

	valuesFilePath := path.Join(r.hostStorageFolder, fmt.Sprintf("values-%s.yaml", mainAgentName))
	err = os.WriteFile(valuesFilePath, yamlData, util.OwnerCanReadWrite)
	if err != nil {
		return nil, fmt.Errorf("failed to write values file: %v", err)
	}

	releaseName := util.NormalizeAgentName(mainAgentName)

	log.Info().Msgf("values file generated at: %s", valuesFilePath)
	log.Info().Msgf("You can deploy the agent running `wfsm deploy` with --dryRun=false` option or `helm install -n %s %s %s --values %s`", namespace, releaseName, chartUrl, valuesFilePath)

	if dryRun {
		return yamlData, nil
	}

	deployer := NewHelmDeployer()
	err = deployer.DeployChart(ctx, releaseName, chartUrl, namespace, yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy chart: %v", err)
	}

	endpoint, err := getNodePortEndpoint(ctx, util.NormalizeAgentName(mainAgentName), namespace)
	if err != nil {
		log.Error().Msgf("failed to get load balancer address: %v", err)
	}

	log.Info().Msg("---------------------------------------------------------------------")
	log.Info().Msgf("ACP agent helm chart release name: %s", releaseName)
	log.Info().Msgf("ACP agent running in namespace: %s, listening for ACP requests on: http://%s", namespace, endpoint)
	log.Info().Msgf("Agent ID: %s", mainAgentID)
	log.Info().Msgf("API Key: %s", mainAgentAPiKey)
	log.Info().Msgf("API Docs: http://%s/agents/%s/docs", endpoint, mainAgentID)
	log.Info().Msgf("\nAllow some time for the agents to start, you can check the status with: kubectl get pods -n %s", namespace)
	log.Info().Msg("---------------------------------------------------------------------\n\n\n")

	return nil, nil
}

func getLoadBalancerEndpoint(ctx context.Context, serviceName string, namespace string, port int) (string, error) {
	log := zerolog.Ctx(ctx)
	ip := "n/a"
	client, err := getK8sClient()
	if err != nil {
		return ip, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	timeout := time.After(60 * time.Second)
	for {
		svc, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			return ip, fmt.Errorf("failed to get service: %v", err)
		}

		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			ingress := svc.Status.LoadBalancer.Ingress[0]
			if ingress.IP != "" {
				ip = ingress.IP
				break
			}
		}

		select {
		case <-ctx.Done():
			return ip, fmt.Errorf("context canceled while waiting for load balancer IP")
		case <-timeout:
			return ip, fmt.Errorf("timeout reached while waiting for load balancer IP")
		default:
			log.Info().Msgf("waiting for load balancer IP")
			time.Sleep(2 * time.Second)
		}
	}

	return fmt.Sprintf("%s:%d", ip, port), nil
}

func getNodePortEndpoint(ctx context.Context, serviceName string, namespace string) (string, error) {
	//log := zerolog.Ctx(ctx)

	ip := "n/a"
	client, err := getK8sClient()
	if err != nil {
		return ip, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	svc, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return ip, fmt.Errorf("failed to get service: %v", err)
	}

	var port int32
	if len(svc.Spec.Ports) > 0 {
		port = svc.Spec.Ports[0].NodePort
	}

	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return ip, fmt.Errorf("failed to get nodes: %v", err)
	}
	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeExternalIP {
				ip = addr.Address
				break
			}
		}
		if ip != "n/a" {
			break
		}
	}

	return fmt.Sprintf("%s:%d", ip, port), nil
}

func (r *runner) createAgentValuesConfig(deploymentSpec internal.AgentDeploymentBuildSpec) (*AgentValues, error) {
	envVars := deploymentSpec.EnvVars

	envVars["API_HOST"] = APIHost
	envVars["API_PORT"] = strconv.Itoa(internal.DEFAULT_API_PORT)
	envVars["AGENT_ID"] = deploymentSpec.AgentID

	secretEnvVars := make(map[string]string, 10)
	secretEnvVars["API_KEY"] = deploymentSpec.ApiKey

	configHash := calculateConfigHash(envVars, secretEnvVars)

	imageRepo, tag := util.SplitImageName(deploymentSpec.Image)

	serviceConfig := deploymentSpec.K8sConfig.Service
	stset := deploymentSpec.K8sConfig.StatefulSet
	podAnnotations := stset.PodAnnotations
	if podAnnotations == nil {
		podAnnotations = make(map[string]string)
	}
	podAnnotations[ConfigCheckSum] = configHash

	agentValues := &AgentValues{
		Name: util.NormalizeAgentName(deploymentSpec.ServiceName),
		Image: Image{
			Repository: imageRepo,
			Tag:        tag,
		},
		//Labels:             deploymentSpec.Labels,
		Env:          convertEnvVars(envVars),
		SecretEnvs:   convertEnvVars(secretEnvVars),
		VolumePath:   "/opt/storage",
		ExternalPort: deploymentSpec.Port,
		InternalPort: internal.DEFAULT_API_PORT,
		Service: internal.Service{
			Type:        serviceConfig.Type,
			Labels:      serviceConfig.Labels,
			Annotations: serviceConfig.Annotations,
		},
		StatefulSet: internal.StatefulSet{
			Replicas:       stset.Replicas,
			Labels:         stset.Labels,
			Annotations:    stset.Annotations,
			PodAnnotations: podAnnotations,
			NodeSelector:   stset.NodeSelector,
			Affinity:       stset.Affinity,
			Tolerations:    stset.Tolerations,
		},
	}

	return agentValues, nil
}

func calculateConfigHash(vars ...map[string]string) string {
	hasher := sha256.New()

	for _, m := range vars {
		// Extract and sort the keys
		keys := make([]string, 0, len(m))
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Concatenate sorted keys and their values
		for _, key := range keys {
			hasher.Write([]byte(fmt.Sprintf("%s=%s;", key, m[key])))
		}
	}

	// Return the final hash as a hexadecimal string
	return fmt.Sprintf("%x", hasher.Sum(nil))
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

func getK8sNamespace() string {
	namespace := "default"
	if ns := os.Getenv("WFSM_K8S_NAMESPACE"); ns != "" {
		namespace = ns
	}
	return namespace
}
