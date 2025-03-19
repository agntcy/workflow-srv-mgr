package python

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	containerClient "github.com/cisco-eti/wfsm/internal/container_client"
	"github.com/cisco-eti/wfsm/internal/deployer/python/source"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/cisco-eti/wfsm/manifests"
	//imagespecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/docker/cli/cli/command"

	"github.com/docker/compose/v2/pkg/compose"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

const AgentImage = "agntcy/wfsm"
const SERVER_PORT = "8000/tcp"
const HOST_IP = "0.0.0.0"
const HOST_PORT_BASE = "15000"

// deployer implementation of AgentDeployer
type deployer struct {
	agManifest         manifests.AgentManifest
	deployment         *manifests.SourceCodeDeployment
	envFilePath        string
	hostStorageFolder  string
	deleteBuildFolders bool
}

func NewPythonAgentDeployer(
	agManifest manifests.AgentManifest,
	srcCodeDeployment *manifests.SourceCodeDeployment,
	envFilePath string, hostStorageFolder string, deleteBuildFolders bool) *deployer {
	return &deployer{
		agManifest:         agManifest,
		deployment:         srcCodeDeployment,
		envFilePath:        envFilePath,
		hostStorageFolder:  hostStorageFolder,
		deleteBuildFolders: deleteBuildFolders,
	}
}

func (d *deployer) Deploy(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	agSrc, err := source.GetAgentSource(d.deployment)
	if err != nil {
		return fmt.Errorf("failed to get agent source: %v", err)
	}

	imageName := strings.Join([]string{AgentImage, d.agManifest.Metadata.Ref.Name}, "-")
	imgNameWithTag, _, err := EnsureContainerImage(ctx, imageName, agSrc, d.deleteBuildFolders)
	if err != nil {
		return fmt.Errorf("failed to get/build container_client image: %v", err)
	}

	log.Debug().Msg(fmt.Sprintf("agent image: %s", imgNameWithTag))
	//TODO should come from outside as a map
	envVars, err := godotenv.Read(d.envFilePath)
	if err != nil {
		return fmt.Errorf("failed to get envVars: %v", err)
	}
	err = d.startContainer(ctx, envVars, imgNameWithTag, err, d.agManifest.Metadata.Ref.Name, log)
	if err != nil {
		return fmt.Errorf("failed to start agent: %v", err)
	}
	return nil
}

func (d *deployer) startContainer(ctx context.Context, envVars map[string]string,
	imgName string, err error, serviceName string, log *zerolog.Logger) error {

	ctx = log.With().Str("serviceName", serviceName).
		Logger().WithContext(ctx)
	log = zerolog.Ctx(ctx)

	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}
	defer containerClient.Close(ctx, cli)

	containerName := strings.Join([]string{"wfsm", d.agManifest.Metadata.Ref.Name}, "-")
	port, err := containerClient.IsContainerRunning(ctx, cli, imgName, containerName)
	if err != nil {
		return fmt.Errorf("failed to check if agent is running: %v", err)
	}
	if port > 0 {
		log.Info().Msg(fmt.Sprintf("agent is already using port: %d", port))
	} else {
		port, err = util.GetNextAvailablePort()
		if err != nil {
			return fmt.Errorf("failed to get next available port: %v", err)
		}
		log.Debug().Msgf("agent will listen on port: %d", port)
	}

	manifestPath := "/opt/storage/manifest.yaml"
	envVars["AGENT_MANIFEST_PATH"] = manifestPath
	envVars["API_HOST"] = "0.0.0.0"
	envVars["API_PORT"] = "8000"
	agentName := d.agManifest.Metadata.Ref.Name

	if d.deployment.FrameworkConfig.LangGraphConfig != nil {

		envVars["AGENT_FRAMEWORK"] = "langgraph"
		graph := d.deployment.FrameworkConfig.LangGraphConfig.Graph
		envVars["AGENTS_REF"] = fmt.Sprintf(`{"%s": "%s"}`, agentName, graph)

	} else if d.deployment.FrameworkConfig.LlamaIndexConfig != nil {

		envVars["AGENT_FRAMEWORK"] = "llamaindex"
		path := d.deployment.FrameworkConfig.LlamaIndexConfig.Path
		envVars["AGENTS_REF"] = fmt.Sprintf(`{"%s": "%s"}`, agentName, path)

	} else {
		return fmt.Errorf("unsupported framework config")
	}

	//platform := strings.Split(util.CurrentArchToDockerPlatform(), "/")
	//dockerPlatform := &imagespecv1.Platform{OS: platform[0], Architecture: platform[1]}

	if d.hostStorageFolder != "" {
		envVars["AGWS_STORAGE_FILE"] = path.Join("/opt/storage", fmt.Sprintf("agws_storage_%s.pkl", agentName))
	}

	manifestFileBuf, err := manifests.NewNullableAgentManifest(&d.agManifest).MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal agent manifest: %v", err)
	}

	err = os.WriteFile(path.Join(d.hostStorageFolder, "manifest.yaml"), manifestFileBuf, util.OwnerCanReadWrite)
	if err != nil {
		return fmt.Errorf("failed to write manifest to temporary workspace dir: %v", err)
	}

	project := &types.Project{
		Name: "wfsm",
	}

	pc, err := types.ParsePortConfig(fmt.Sprintf("0.0.0.0:%v:%v", port, SERVER_PORT))
	if err != nil {
		return fmt.Errorf("failed to parse port config: %v", err)
	}
	project.Services = make(map[string]types.ServiceConfig)
	project.Services[serviceName] = types.ServiceConfig{
		Name: serviceName,
		Labels: map[string]string{
			api.ProjectLabel: project.Name,
			api.OneoffLabel:  "False",
			api.ServiceLabel: serviceName,
		},
		//ContainerName: serviceName,
		Image:       imgName,
		Ports:       pc,
		Environment: getEnvVars(envVars),
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   "bind",
				Source: d.hostStorageFolder,
				Target: "/opt/storage",
			},
		},
	}

	// Use the MarshalYAML method to get YAML representation
	projectYAML, err := project.MarshalYAML()
	if err != nil {
		return err
	}
	log.Debug().Msg(string(projectYAML))

	dockerCli, err := command.NewDockerCli(command.WithBaseContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to create docker cli: %v", err)
	}
	clientOptions := flags.ClientOptions{
		Hosts: []string{"unix:///var/run/docker.sock"},
		//opts := &flags.ClientOptions{Hosts: []string{fmt.Sprintf("unix://%s", socket)}}
		LogLevel:  "debug",
		TLS:       false,
		TLSVerify: false,
	}
	err = dockerCli.Initialize(&clientOptions)
	if err != nil {
		return fmt.Errorf("failed to initialize docker client: %v", err)
	}

	backend := compose.NewComposeService(dockerCli) //.(commands.Backend)
	err = backend.Up(ctx, project, api.UpOptions{
		//Create: api.CreateOptions{
		//	Services: project.ServiceNames(),
		//},
		//Start: api.StartOptions{
		//	Services: project.ServiceNames(),
		//	Project:  project,
		//},
	})
	if err != nil {
		return err
	}

	list, err := backend.Ps(ctx, project.Name, api.PsOptions{All: true})
	if err != nil {
		return err
	}
	for _, c := range list {
		//log.Debug().Msg(fmt.Sprintf("%s %s %s %v", c.Name, c.ID, c.Status))
		log.Info().Msg(fmt.Sprintf("agent running in container: %s, listening on port: %d status: %s", c.Name, port, c.Status))
	}

	logConsumer := formatter.NewLogConsumer(ctx, os.Stdout, os.Stderr, true, true, true)
	err = backend.Logs(ctx, project.Name, logConsumer, api.LogOptions{
		Project:  project,
		Services: []string{serviceName},
		Tail:     "100",
		Follow:   true,
	})
	if err != nil {
		return err
	}

	log.Info().Msg(fmt.Sprintf("agent running in container: %s, listening on port: %d", serviceName, port))

	return nil
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

func getHostConfig(port nat.Port, hostIP string, hostPort string) *container.HostConfig {
	return &container.HostConfig{
		PortBindings: nat.PortMap{
			port: []nat.PortBinding{
				{
					HostIP:   hostIP,
					HostPort: hostPort,
				},
			},
		},
	}
}

func getContainerConfig(imgName string, envVars map[string]string, port nat.Port) *container.Config {
	containerConfig := &container.Config{
		Image: imgName,
		Env:   []string{},
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
	}
	for k, v := range envVars {
		containerConfig.Env = append(containerConfig.Env, fmt.Sprintf("%s=%s", k, v))
	}
	return containerConfig
}
