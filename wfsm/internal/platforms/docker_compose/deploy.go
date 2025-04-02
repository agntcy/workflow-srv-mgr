// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package docker

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/cisco-eti/wfsm/internal"
	containerClient "github.com/cisco-eti/wfsm/internal/container_client"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/docker/compose/v2/pkg/api"
	dockerClient "github.com/docker/docker/client"

	cmdcmp "github.com/docker/compose/v2/cmd/compose"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/rs/zerolog"
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
			depAgPrefix := calculateEnvVarPrefix(depName)
			depSpec := agentDeploymentSpecs[depName]
			agSpec.EnvVars[depAgPrefix+"API_KEY"] = fmt.Sprintf("{\"x-api-key\": \"%s\"}", depSpec.ApiKey)
			agSpec.EnvVars[depAgPrefix+"ID"] = depSpec.AgentID
			agSpec.EnvVars[depAgPrefix+"ENDPOINT"] = fmt.Sprintf("http://%s:%s", depSpec.ServiceName, APIPort)
		}
	}

	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %v", err)
	}
	defer containerClient.Close(ctx, cli)

	project := &types.Project{
		Name: mainAgentName,
	}
	project.Services = make(map[string]types.ServiceConfig)

	// only the main agent will be exposed to the outside world
	mainAgentSpec := agentDeploymentSpecs[mainAgentName]

	port := externalPort
	if port == 0 {
		port, err = r.getMainAgentPublicPort(ctx, cli, mainAgentName, mainAgentSpec)
		if err != nil {
			return nil, err
		}
	}

	mainAgentID := mainAgentSpec.AgentID
	mainAgentAPiKey := mainAgentSpec.ApiKey
	sc, err := r.createServiceConfig(mainAgentName, mainAgentSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to create service config: %v", err)
	}
	pc, err := types.ParsePortConfig(fmt.Sprintf("0.0.0.0:%v:%v", port, ServerPort))
	if err != nil {
		return nil, fmt.Errorf("failed to parse port config: %v", err)
	}
	sc.Ports = pc
	project.Services[mainAgentSpec.ServiceName] = *sc
	delete(agentDeploymentSpecs, mainAgentName)

	// generate service configs for dependencies
	for _, deploymentSpec := range agentDeploymentSpecs {
		sc, err := r.createServiceConfig(mainAgentName, deploymentSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to create service config: %v", err)
		}
		project.Services[deploymentSpec.ServiceName] = *sc
	}

	dockerCli, err := getDockerCLI(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %v", err)
	}
	defer dockerCli.Client().Close()

	composeFilePath := path.Join(r.hostStorageFolder, fmt.Sprintf("compose-%s.yaml", mainAgentName))
	prjOpts := cmdcmp.ProjectOptions{
		ConfigPaths: []string{
			composeFilePath,
		},
	}

	projectYaml, err := project.MarshalYAML()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal compose config: %v", err)
	}

	err = os.WriteFile(composeFilePath, projectYaml, util.OwnerCanReadWrite)
	if err != nil {
		return nil, fmt.Errorf("failed to write compose config: %v", err)
	}
	project, _, err = prjOpts.ToProject(ctx, dockerCli, []string{
		//deploymentSpec.ServiceName
	})
	log.Info().Msgf("compose file generated at: %s", composeFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %v", err)
	}

	if dryRun {
		return projectYaml, nil
	}

	backend := compose.NewComposeService(dockerCli) //.(commands.Backend)
	err = backend.Up(ctx, project, api.UpOptions{api.CreateOptions{RemoveOrphans: true}, api.StartOptions{}})
	if err != nil {
		return nil, err
	}

	log.Info().Msg("---------------------------------------------------------------------")
	log.Info().Msgf("ACP agent deployment name: %s", mainAgentName)
	log.Info().Msgf("ACP agent running in container: %s, listening for ACP requests on: http://127.0.0.1:%d", mainAgentName, port)
	log.Info().Msgf("Agent ID: %s", mainAgentID)
	log.Info().Msgf("API Key: %s", mainAgentAPiKey)
	log.Info().Msgf("API Docs: http://127.0.0.1:%d/agents/%s/docs", port, mainAgentID)
	log.Info().Msg("---------------------------------------------------------------------\n\n\n")

	logConsumer := formatter.NewLogConsumer(ctx, os.Stdout, os.Stderr, true, true, true)
	err = backend.Logs(ctx, project.Name, logConsumer, api.LogOptions{
		Project:  project,
		Services: []string{},
		Tail:     "100",
		Follow:   true,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func calculateEnvVarPrefix(agName string) string {
	prefix := strings.ToUpper(agName)
	// replace all non-alphanumeric characters with _
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	prefix = re.ReplaceAllString(prefix, "_")
	return prefix + "_"
}

func (r *runner) getMainAgentPublicPort(ctx context.Context, cli *dockerClient.Client, mainAgentName string, mainAgentSpec internal.AgentDeploymentBuildSpec) (int, error) {
	log := zerolog.Ctx(ctx)

	containerName := strings.Join([]string{mainAgentName, mainAgentSpec.DeploymentName}, "-")
	port, err := containerClient.IsContainerRunning(ctx, cli, mainAgentSpec.Image, containerName)
	if err != nil {
		return 0, fmt.Errorf("failed to check if agent is running: %v", err)
	}
	if port > 0 {
		log.Info().Msg(fmt.Sprintf("agent is already using port: %d", port))
	} else {
		port, err = util.GetNextAvailablePort()
		if err != nil {
			return 0, fmt.Errorf("failed to get next available port: %v", err)
		}
		log.Debug().Msgf("agent will listen on port: %d", port)
	}
	return port, nil
}

func (r *runner) createServiceConfig(projectName string, deploymentSpec internal.AgentDeploymentBuildSpec) (*types.ServiceConfig, error) {

	agentID := deploymentSpec.AgentID
	apiKey := deploymentSpec.ApiKey

	envVars := deploymentSpec.EnvVars
	envVars["API_HOST"] = APIHost
	envVars["API_PORT"] = APIPort
	envVars["API_KEY"] = apiKey

	srcDeployment := deploymentSpec.Manifest.Deployment.DeploymentOptions[deploymentSpec.SelectedDeploymentOption].SourceCodeDeployment
	if srcDeployment.FrameworkConfig.LangGraphConfig != nil {

		envVars["AGENT_FRAMEWORK"] = "langgraph"
		graph := srcDeployment.FrameworkConfig.LangGraphConfig.Graph
		envVars["AGENTS_REF"] = fmt.Sprintf(`{"%s": "%s"}`, agentID, graph)

	} else if srcDeployment.FrameworkConfig.LlamaIndexConfig != nil {

		envVars["AGENT_FRAMEWORK"] = "llamaindex"
		path := srcDeployment.FrameworkConfig.LlamaIndexConfig.Path
		envVars["AGENTS_REF"] = fmt.Sprintf(`{"%s": "%s"}`, agentID, path)

	} else {
		return nil, fmt.Errorf("unsupported framework config")
	}

	agDeploymentFolder := path.Join(r.hostStorageFolder, deploymentSpec.DeploymentName)
	// make sure the folder exists
	if _, err := os.Stat(agDeploymentFolder); os.IsNotExist(err) {
		if err := os.Mkdir(agDeploymentFolder, 0755); err != nil {
			return nil, fmt.Errorf("failed to create deployment folder for agent: %v", err)
		}
	}

	envVars["AGWS_STORAGE_FILE"] = path.Join("/opt/storage", fmt.Sprintf("agws_storage.pkl"))

	sc := types.ServiceConfig{
		Name: deploymentSpec.ServiceName,
		Labels: map[string]string{
			api.ProjectLabel: projectName,
			api.OneoffLabel:  "False",
			api.ServiceLabel: deploymentSpec.ServiceName,
		},
		//ContainerName: serviceName,
		Image:       deploymentSpec.Image,
		Environment: getEnvVars(envVars),
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   "bind",
				Source: agDeploymentFolder,
				Target: "/opt/storage",
			},
		},
	}
	return &sc, nil
}

func getDockerCLI(ctx context.Context) (*command.DockerCli, error) {
	dockerCli, err := command.NewDockerCli(command.WithBaseContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create docker cli: %v", err)
	}
	clientOptions := flags.ClientOptions{
		LogLevel:  "debug",
		TLS:       false,
		TLSVerify: false,
	}
	err = dockerCli.Initialize(&clientOptions)
	return dockerCli, err
}

func getStringPtr(s string) *string {
	return &s
}

func getEnvVars(envvars map[string]string) map[string]*string {
	ev := make(map[string]*string)
	for k, v := range envvars {
		// clone the v value to avoid reference issues
		ev[k] = getStringPtr(strings.Clone(v))
	}
	return ev
}
