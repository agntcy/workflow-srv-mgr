package docker

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/docker/compose/v2/pkg/api"
	dockerClient "github.com/docker/docker/client"

	containerClient "github.com/cisco-eti/wfsm/internal/container_client"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/cisco-eti/wfsm/manifests"
	"github.com/docker/cli/cli/command"

	cmdcmp "github.com/docker/compose/v2/cmd/compose"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/rs/zerolog"
)

const SERVER_PORT = "8000/tcp"
const API_HOST = "0.0.0.0"
const API_PORT = "8000"

func (r *runner) Deploy(ctx context.Context, deploymentSpec internal.AgentDeploymentSpec, dryRun bool) (internal.DeploymentArtifact, error) {
	log := zerolog.Ctx(ctx)
	ctx = log.With().Str("serviceName", deploymentSpec.ServiceName).
		Logger().WithContext(ctx)
	log = zerolog.Ctx(ctx)

	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %v", err)
	}
	defer containerClient.Close(ctx, cli)

	projectName := deploymentSpec.DeploymentName
	containerName := strings.Join([]string{projectName, deploymentSpec.DeploymentName}, "-")
	port, err := containerClient.IsContainerRunning(ctx, cli, deploymentSpec.Image, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if agent is running: %v", err)
	}
	if port > 0 {
		log.Info().Msg(fmt.Sprintf("agent is already using port: %d", port))
	} else {
		port, err = util.GetNextAvailablePort()
		if err != nil {
			return nil, fmt.Errorf("failed to get next available port: %v", err)
		}
		log.Debug().Msgf("agent will listen on port: %d", port)
	}

	manifestPath := "/opt/storage/manifest.yaml"
	envVars := deploymentSpec.EnvVars
	envVars["AGENT_MANIFEST_PATH"] = manifestPath
	envVars["API_HOST"] = API_HOST
	envVars["API_PORT"] = API_PORT
	//TODO add AGENT_ID and API_KEY

	agentName := deploymentSpec.Manifest.Metadata.Ref.Name

	srcDeployment := deploymentSpec.Manifest.Deployment.DeploymentOptions[deploymentSpec.SelectedDeploymentOption].SourceCodeDeployment
	if srcDeployment.FrameworkConfig.LangGraphConfig != nil {

		envVars["AGENT_FRAMEWORK"] = "langgraph"
		graph := srcDeployment.FrameworkConfig.LangGraphConfig.Graph
		envVars["AGENTS_REF"] = fmt.Sprintf(`{"%s": "%s"}`, agentName, graph)

	} else if srcDeployment.FrameworkConfig.LlamaIndexConfig != nil {

		envVars["AGENT_FRAMEWORK"] = "llamaindex"
		path := srcDeployment.FrameworkConfig.LlamaIndexConfig.Path
		envVars["AGENTS_REF"] = fmt.Sprintf(`{"%s": "%s"}`, agentName, path)

	} else {
		return nil, fmt.Errorf("unsupported framework config")
	}

	//platforms := strings.Split(util.CurrentArchToDockerPlatform(), "/")
	//dockerPlatform := &imagespecv1.Platform{OS: platforms[0], Architecture: platforms[1]}
	agDeploymentFolder := path.Join(r.hostStorageFolder, deploymentSpec.DeploymentName)
	// make sure the folder exists
	if _, err := os.Stat(agDeploymentFolder); os.IsNotExist(err) {
		if err := os.Mkdir(agDeploymentFolder, 0755); err != nil {
			return nil, fmt.Errorf("failed to create deployment folder for agent: %v", err)
		}
	}

	envVars["AGWS_STORAGE_FILE"] = path.Join("/opt/storage", fmt.Sprintf("agws_storage.pkl"))

	manifestFileBuf, err := manifests.NewNullableAgentManifest(&deploymentSpec.Manifest).MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent manifest: %v", err)
	}

	err = os.WriteFile(path.Join(agDeploymentFolder, "manifest.yaml"), manifestFileBuf, util.OwnerCanReadWrite)
	if err != nil {
		return nil, fmt.Errorf("failed to write manifest to temporary workspace dir: %v", err)
	}

	project := &types.Project{
		Name: projectName,
	}

	pc, err := types.ParsePortConfig(fmt.Sprintf("0.0.0.0:%v:%v", port, SERVER_PORT))
	if err != nil {
		return nil, fmt.Errorf("failed to parse port config: %v", err)
	}
	project.Services = make(map[string]types.ServiceConfig)
	project.Services[deploymentSpec.ServiceName] = types.ServiceConfig{
		Name: deploymentSpec.ServiceName,
		Labels: map[string]string{
			api.ProjectLabel: project.Name,
			api.OneoffLabel:  "False",
			api.ServiceLabel: deploymentSpec.ServiceName,
		},
		//ContainerName: serviceName,
		Image:       deploymentSpec.Image,
		Ports:       pc,
		Environment: getEnvVars(envVars),
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   "bind",
				Source: agDeploymentFolder,
				Target: "/opt/storage",
			},
		},
	}

	dockerCli, err := getDockerCLI(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %v", err)
	}
	defer dockerCli.Client().Close()

	prjOpts := cmdcmp.ProjectOptions{
		ConfigPaths: []string{
			path.Join(agDeploymentFolder, "compose.yaml"),
		},
	}

	projectYaml, err := project.MarshalYAML()
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(path.Join(agDeploymentFolder, "compose.yaml"), projectYaml, util.OwnerCanReadWrite)
	project, _, err = prjOpts.ToProject(ctx, dockerCli, []string{deploymentSpec.ServiceName})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %v", err)
	}

	if dryRun {
		return projectYaml, nil
	}

	backend := compose.NewComposeService(dockerCli) //.(commands.Backend)
	err = backend.Up(ctx, project, api.UpOptions{})
	if err != nil {
		return nil, err
	}

	//list, err := backend.Ps(ctx, project.Name, api.PsOptions{All: true})
	//if err != nil {
	//	return nil, err
	//}
	//for _, c := range list {
	//	log.Info().Msg(fmt.Sprintf("agent running in container: %s, listening on port: %d status: %s", c.Name, port, c.Status))
	//}

	log.Info().Msg("---------------------------------------------------------------------")
	log.Info().Msg(fmt.Sprintf("agent running in container: %s, listening on: http://127.0.0.1:%d", deploymentSpec.ServiceName, port))
	log.Info().Msg("---------------------------------------------------------------------\n\n\n")

	logConsumer := formatter.NewLogConsumer(ctx, os.Stdout, os.Stderr, true, true, true)
	err = backend.Logs(ctx, project.Name, logConsumer, api.LogOptions{
		Project:  project,
		Services: []string{deploymentSpec.ServiceName},
		Tail:     "100",
		Follow:   true,
	})
	if err != nil {
		return nil, err
	}

	return project.MarshalYAML()
}

func getDockerCLI(ctx context.Context) (*command.DockerCli, error) {
	dockerCli, err := command.NewDockerCli(command.WithBaseContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create docker cli: %v", err)
	}
	clientOptions := flags.ClientOptions{
		Hosts: []string{"unix:///var/run/docker.sock"},
		//opts := &flags.ClientOptions{Hosts: []string{fmt.Sprintf("unix://%s", socket)}}
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
