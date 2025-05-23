// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package docker

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/cisco-eti/wfsm/internal"
	containerClient "github.com/cisco-eti/wfsm/internal/container_client"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/docker/compose/v2/pkg/api"
	dockerClient "github.com/docker/docker/client"

	cmdcmp "github.com/docker/compose/v2/cmd/compose"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/rs/zerolog"
)

const APIHost = "0.0.0.0"

// Deploy if externalPort is 0, will try to find the port of already running container or find next available port
func (r *runner) Deploy(ctx context.Context,
	mainAgentName string,
	agentDeploymentSpecs map[string]internal.AgentDeploymentBuildSpec,
	dependencies map[string][]string,
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
			agSpec.EnvVars[depAgPrefix+"ENDPOINT"] = fmt.Sprintf("http://%s:%d", depSpec.ServiceName, internal.DEFAULT_API_PORT)
		}
	}

	dockerCli, err := util.GetDockerCLI(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %v", err)
	}
	defer dockerCli.Client().Close()

	project := &types.Project{
		Name: mainAgentName,
	}
	project.Services = make(map[string]types.ServiceConfig)

	// only the main agent will be exposed to the outside world
	mainAgentSpec := agentDeploymentSpecs[mainAgentName]

	port := mainAgentSpec.Port
	//if port == 0 {
	//	port, err = r.getMainAgentPublicPort(ctx, dockerCli.Client(), mainAgentName, mainAgentSpec)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	mainAgentID := mainAgentSpec.AgentID
	mainAgentAPiKey := mainAgentSpec.ApiKey

	// generate service configs for dependencies
	for _, deploymentSpec := range agentDeploymentSpecs {
		sc, err := r.createServiceConfig(mainAgentName, deploymentSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to create service config: %v", err)
		}
		project.Services[deploymentSpec.ServiceName] = *sc
	}

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
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %v", err)
	}
	log.Info().Msgf("Compose file generated at: %s", composeFilePath)
	log.Info().Msgf("You can deploy the agent running `wfsm deploy` with `--dryRun=false` option or `docker compose -f %v up`", composeFilePath)
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

func (r *runner) getMainAgentPublicPort(ctx context.Context, cntClient dockerClient.ContainerAPIClient, mainAgentName string, mainAgentSpec internal.AgentDeploymentBuildSpec) (int, error) {
	log := zerolog.Ctx(ctx)

	containerName := strings.Join([]string{mainAgentName, mainAgentSpec.DeploymentName}, "-")
	port, err := containerClient.IsContainerRunning(ctx, cntClient, mainAgentSpec.Image, containerName)
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

	envVars := deploymentSpec.EnvVars

	envVars["API_HOST"] = APIHost
	envVars["API_PORT"] = strconv.Itoa(internal.DEFAULT_API_PORT)

	envVars["API_KEY"] = deploymentSpec.ApiKey
	envVars["AGENT_ID"] = deploymentSpec.AgentID

	agDeploymentFolder := path.Join(r.hostStorageFolder, deploymentSpec.DeploymentName)
	// make sure the folder exists
	if _, err := os.Stat(agDeploymentFolder); os.IsNotExist(err) {
		if err := os.Mkdir(agDeploymentFolder, 0755); err != nil {
			return nil, fmt.Errorf("failed to create deployment folder for agent: %v", err)
		}
	}

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

	if deploymentSpec.Port > 0 {
		pc, err := types.ParsePortConfig(fmt.Sprintf("0.0.0.0:%v:%v/tcp", deploymentSpec.Port, internal.DEFAULT_API_PORT))
		if err != nil {
			return nil, fmt.Errorf("failed to parse port config: %v", err)
		}
		sc.Ports = pc
	}

	return &sc, nil
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
