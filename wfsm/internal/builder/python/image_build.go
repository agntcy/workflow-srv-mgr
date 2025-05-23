// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package python

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/cisco-eti/wfsm/assets"
	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/builder/python/source"
	containerclient "github.com/cisco-eti/wfsm/internal/container_client"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/cisco-eti/wfsm/manifests"
	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/rs/zerolog"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	dockerclient "github.com/docker/docker/client"
)

// containerImageBuildLock is used to synchronize the build of container images
//
//nolint:mnd
var containerImageBuildLock = util.NewStripedLock(100)

// WorkflowServerRepo is the repo name for the workflow server docker image
const WorkflowServerRepo = "ghcr.io/agntcy/acp/wfsrv"

// EnsureContainerImage - ensure container image is available. If the image exists, it returns the name of the
// existing  image, otherwise it builds a new image with the necessary packages installed
// and returns its name.
func EnsureContainerImage(ctx context.Context, img string, src source.AgentSource, inputSpec internal.AgentSpec, deleteBuildFolders bool, forceBuild bool, baseImage string) (string, error) {

	log := zerolog.Ctx(ctx)
	ctx = log.WithContext(ctx)

	containerImageBuildLock.Lock(img)
	defer containerImageBuildLock.Unlock(img)

	var err error
	workspacePath, err := os.MkdirTemp("", "wfsm_build_")
	if err != nil {
		return "", fmt.Errorf("creating temporary workspace dir failed: %v", err)
	}

	log.Info().Str("workspace_path", workspacePath).Msg("created temporary workspace dir")

	agentSourceDir := "agent_src"
	agentSrcPath := path.Join(workspacePath, agentSourceDir)

	log.Info().Str("agent_src_path", agentSrcPath).Msg("copying agent source to workspace")

	err = src.CopyToWorkspace(agentSrcPath)
	if err != nil {
		return "", fmt.Errorf("failed to copy agent source to workspace: %v", err)
	}

	if deleteBuildFolders {
		defer func() {
			if err := os.RemoveAll(workspacePath); err != nil {
				log.Error().Err(err).Str("path", workspacePath).Msg("failed to remove temporary workspace dir")
			}
		}()
	}

	//copy manifest file to workspace
	manifestFileBuf, err := manifests.NewNullableAgentManifest(&inputSpec.Manifest).MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal agent manifest: %v", err)
	}
	manifestFile := path.Join(workspacePath, "manifest.json")
	err = os.WriteFile(manifestFile, manifestFileBuf, util.OwnerCanReadWrite)

	// calc. hash based on agent source files and manifest file and use as image tag
	hashCode := calculateHash(workspacePath, baseImage)
	img = fmt.Sprintf("%s:%s", img, hashCode)

	dockerCli, err := util.GetDockerCLI(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to initialize docker client: %v", err)
	}
	defer dockerCli.Client().Close()

	// check if image already exists unless forceBuild is set
	if !forceBuild {
		found, err := findImage(ctx, dockerCli.Client(), img)
		if err != nil {
			return "", err
		}
		if found {
			return img, nil
		}
		log.Info().Str("image", img).Msg("image not found on runtime host")
	}

	baseImage, err = getBaseImage(baseImage, WorkflowServerRepo)
	if err != nil {
		return "", fmt.Errorf("failed to get base image: %v", err)
	}
	log.Info().Str("image", baseImage).Msg("base image to be used for building agent image")

	// find base image and pull it if not found
	found, err := findImage(ctx, dockerCli.Client(), baseImage)
	if err != nil {
		return "", err
	}
	if !found {
		log.Info().Str("image", baseImage).Msg("base image not found on container runtime host")
		// image not available locally, see if it can be pulled from registry
		err = pullImage(ctx, dockerCli.Client(), baseImage)
		if err != nil {
			if !errdefs.IsNotFound(err) {
				return "", fmt.Errorf("base image not found %s: %w", baseImage, err)
			}
			return "", fmt.Errorf("failed to pull base image %s: %w", baseImage, err)
		}
	}

	// build image
	err = buildImage(ctx, dockerCli.Client(), img, workspacePath, inputSpec, agentSourceDir, assets.AgentBuilderDockerfile, baseImage)
	if err != nil {
		return "", fmt.Errorf("failed to build image %s: %w", img, err)
	}

	return img, nil
}

func getBaseImage(baseImage string, workflowServerRepo string) (string, error) {
	if baseImage != "" {
		return baseImage, nil
	}

	repo, err := name.NewRepository(workflowServerRepo)
	if err != nil {
		return "", fmt.Errorf("Failed to create repo: %v", err)
	}

	// default page size is 1000
	tags, err := remote.List(repo)
	if err != nil {
		return "", fmt.Errorf("Failed to list tags: %v", err)
	}
	tag, err := util.GetLatestTag(tags)
	if err != nil {
		return "", fmt.Errorf("Failed to get the most recent tag from the following tags: %v, %v", tags, err)
	}
	return fmt.Sprintf("%s:%s", workflowServerRepo, tag), nil
}

func findImage(ctx context.Context, client dockerclient.ImageAPIClient, img string) (bool, error) {
	log := zerolog.Ctx(ctx)
	imageList, err := client.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: img,
		}),
	})
	if err != nil {
		return false, fmt.Errorf("failed to retrieve the list of images from the conainer runtime host: %w", err)
	}

	if len(imageList) == 1 {
		log.Info().Str("image", img).Msg("image found on runtime host")
		return true, nil
	}
	if len(imageList) > 1 {
		return false, fmt.Errorf("more than one image %q found on runtime host", img)
	}

	return false, nil
}

func buildImage(ctx context.Context, client dockerclient.ImageAPIClient, img string, workspacePath string, inputSpec internal.AgentSpec, agentSourceDir string, dockerFile []byte, baseImage string) error {
	log := zerolog.Ctx(ctx)
	log.Info().Str("image", img).Msg("building image")

	if err := os.WriteFile(path.Join(workspacePath, "Dockerfile"), dockerFile, util.OwnerCanReadWrite); err != nil {
		return fmt.Errorf("failed to write dockerfile to temporary workspace dir for building image: %w", err)
	}
	if err := os.WriteFile(path.Join(workspacePath, "start_agws.sh"), assets.StartAGWSScript, util.OwnerCanReadWrite); err != nil {
		return fmt.Errorf("failed to write dockerfile to temporary workspace dir for building image: %w", err)
	}

	imageBuildContext, err := containerclient.CreateBuildContext(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to create build context for image building: %w", err)
	}
	defer func() {
		if err := imageBuildContext.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close archived build context reader")
		}
		log.Debug().Msg("closed image build context")
	}()

	buildArgs := map[string]*string{
		"AGENT_DIR":  &agentSourceDir,
		"BASE_IMAGE": &baseImage,
	}

	deployment := inputSpec.Manifest.Extensions[0].Data.Deployment
	srcDeployment := deployment.DeploymentOptions[inputSpec.SelectedDeploymentOption].SourceCodeDeployment
	if srcDeployment.FrameworkConfig.LangGraphConfig != nil {

		buildArgs["AGENT_FRAMEWORK"] = &srcDeployment.FrameworkConfig.LangGraphConfig.FrameworkType
		graph := srcDeployment.FrameworkConfig.LangGraphConfig.Graph
		buildArgs["AGENT_OBJECT"] = &graph

	} else if srcDeployment.FrameworkConfig.LlamaIndexConfig != nil {

		buildArgs["AGENT_FRAMEWORK"] = &srcDeployment.FrameworkConfig.LlamaIndexConfig.FrameworkType
		path := srcDeployment.FrameworkConfig.LlamaIndexConfig.Path
		buildArgs["AGENT_OBJECT"] = &path

	} else {
		return fmt.Errorf("unsupported framework config")
	}

	buildResp, err := client.ImageBuild(ctx, imageBuildContext, types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{img},
		BuildArgs:  buildArgs,
		NoCache:    true,
		Remove:     true,
		PullParent: false,
		Platform:   util.CurrentArchToDockerPlatform(),
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := buildResp.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close image build response body")
		}
		log.Debug().Msg("closed image build response body")
	}()

	err = displayDockerLogs(buildResp.Body)
	if err != nil {
		return err
	}

	log.Info().Msg("successfully built image")
	return nil
}

func pullImage(ctx context.Context, client dockerclient.ImageAPIClient, img string) error {
	log := zerolog.Ctx(ctx)
	log.Info().Msgf("pulling image: %s", img)

	reader, err := client.ImagePull(ctx, img, image.PullOptions{Platform: util.CurrentArchToDockerPlatform()})
	if err != nil {
		return err
	}

	defer func() {
		if err := reader.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close docker log reader")
		}
	}()

	err = displayDockerLogs(reader)
	if err != nil {
		return err
	}

	return nil
}

func displayDockerLogs(reader io.ReadCloser) error {
	rd := bufio.NewReader(reader)
	var logLine []byte
	var imageBuildLogLine jsonmessage.JSONMessage
	for {
		line, isPrefix, err := rd.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}

			return fmt.Errorf("failed to read from image build response: %w", err)
		}
		logLine = append(logLine, line...)
		if !isPrefix {
			if err = json.Unmarshal(logLine, &imageBuildLogLine); err != nil {
				return fmt.Errorf("failed to unmarshal image build log line: %w", err)
			}
			err = imageBuildLogLine.Display(os.Stdout, true)
			if err != nil {
				return err
			}
			logLine = logLine[:0]
		}
	}
	return nil
}
