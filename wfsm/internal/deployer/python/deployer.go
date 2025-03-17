package python

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	containerClient "github.com/cisco-eti/wfsm/internal/container_client"
	"github.com/cisco-eti/wfsm/internal/deployer/python/source"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/cisco-eti/wfsm/manifests"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	imagespecv1 "github.com/opencontainers/image-spec/specs-go/v1"

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

	envVars, err := godotenv.Read(d.envFilePath)
	if err != nil {
		return fmt.Errorf("failed to get envVars: %v", err)
	}

	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}
	defer containerClient.Close(ctx, cli)

	containerName := strings.Join([]string{"wfsm", d.agManifest.Metadata.Ref.Name}, "-")
	containerID, err := containerClient.IsContainerRunning(ctx, cli, imgNameWithTag, containerName)
	if err != nil {
		return fmt.Errorf("failed to check if agent is running: %v", err)
	}

	if len(containerID) > 0 {
		log.Info().Msg("agent is already running")

		err := cli.ContainerStop(ctx, containerID, container.StopOptions{})
		if err != nil {
			return fmt.Errorf("failed to stop existing container_client: %v", err)
		}

		err = containerClient.RemoveContainer(ctx, cli, containerID)
		if err != nil {
			return fmt.Errorf("failed to remove existing container_client: %v", err)
		}
	}

	containerID, err = d.startContainer(ctx, envVars, imgNameWithTag, err, cli, containerName, log)
	if err != nil {
		return fmt.Errorf("failed to start container_client: %v", err)
	}

	rc, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})

	if err != nil {
		return fmt.Errorf("failed to get container_client logs: %v", err)
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, rc)

	return nil
}

func (d *deployer) startContainer(ctx context.Context, envVars map[string]string,
	imgName string, err error, cli *dockerClient.Client, containerName string, log *zerolog.Logger) (string, error) {

	manifestPath := "/opt/manifest.yaml"
	envVars["AGENT_MANIFEST_PATH"] = manifestPath
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
		return "", fmt.Errorf("unsupported framework config")
	}

	port, err := util.GetNextAvailablePort()
	if err != nil {
		return "", fmt.Errorf("failed to get next available port: %v", err)
	}
	log.Debug().Msgf("Next available port: %d", port)

	platform := strings.Split(util.CurrentArchToDockerPlatform(), "/")
	dockerPlatform := &imagespecv1.Platform{OS: platform[0], Architecture: platform[1]}

	hostConfig := getHostConfig(SERVER_PORT, HOST_IP, strconv.Itoa(port))
	if d.hostStorageFolder != "" {
		hostConfig.Mounts = []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   d.hostStorageFolder,
				Target:   "/opt/storage",
				ReadOnly: false,
			},
		}
		envVars["AGWS_STORAGE_FILE"] = path.Join("/opt/storage", fmt.Sprintf("agws_storage_%s.pkl", agentName))
	}
	containerConfig := getContainerConfig(imgName, envVars, SERVER_PORT)

	containerID, err := containerClient.CreateContainer(ctx, cli, containerConfig, hostConfig, dockerPlatform, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create docker container_client: %v", err)
	}

	ctx = log.With().Str("container_id", containerID).
		Logger().WithContext(ctx)
	log = zerolog.Ctx(ctx)

	manifestFileBuf, err := manifests.NewNullableAgentManifest(&d.agManifest).MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal agent manifest: %v", err)
	}
	workspacePath, err := os.MkdirTemp("", "agent_run_")
	if err != nil {
		return "", fmt.Errorf("creating temporary workspace dir failed: %v", err)
	}
	err = os.WriteFile(path.Join(workspacePath, "manifest.yaml"), manifestFileBuf, util.OwnerCanReadWrite)
	if err != nil {
		return "", fmt.Errorf("failed to write manifest to temporary workspace dir: %v", err)
	}

	err = containerClient.CopyToContainer(ctx, cli, containerID, workspacePath, "/opt")
	if err != nil {
		return "", fmt.Errorf("failed to copy manifest to container_client: %w", err)
	}

	if err := os.RemoveAll(workspacePath); err != nil {
		log.Error().Err(err).Str("path", workspacePath).Msg("failed to remove temporary dir")
	}

	err = containerClient.StartContainer(ctx, cli, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to start docker container_client: %v", err)
	}

	log.Info().Msg(fmt.Sprintf("agent running in container: %s, listening on port: %d", containerName, port))

	return containerID, nil
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
