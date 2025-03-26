// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package python

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/cisco-eti/wfsm/assets"
	"github.com/cisco-eti/wfsm/internal/builder/python/source"
	containerclient "github.com/cisco-eti/wfsm/internal/container_client"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/rs/zerolog"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	dockerclient "github.com/docker/docker/client"
)

// containerImageBuildLock is used to synchronize the build of container images
//
//nolint:mnd
var containerImageBuildLock = util.NewStripedLock(100)

// EnsureContainerImage - ensure container image is available. If the image exists, it returns the name of the
// existing  image, otherwise it builds a new image with the necessary packages installed
// and returns its name.
func EnsureContainerImage(ctx context.Context, img string, src source.AgentSource, deleteBuildFolders bool) (string, bool, error) {

	log := zerolog.Ctx(ctx).With().Str("platforms", "runtime").Logger()
	ctx = log.WithContext(ctx)

	containerImageBuildLock.Lock(img)
	defer containerImageBuildLock.Unlock(img)

	var err error
	workspacePath, err := os.MkdirTemp("", "wfsm_build_")
	if err != nil {
		return "", false, fmt.Errorf("creating temporary workspace dir failed: %v", err)
	}

	log.Debug().Str("workspace_path", workspacePath).Msg("created temporary workspace dir")

	agentSourceDir := "agent_src"
	agentSrcPath := path.Join(workspacePath, agentSourceDir)

	log.Debug().Str("agent_src_path", agentSrcPath).Msg("copying agent source to workspace")

	err = src.CopyToWorkspace(agentSrcPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to copy agent source to workspace: %v", err)
	}

	if deleteBuildFolders {
		defer func() {
			if err := os.RemoveAll(workspacePath); err != nil {
				log.Error().Err(err).Str("path", workspacePath).Msg("failed to remove temporary workspace dir")
			}
		}()
	}

	// calc. hash based on agent source files will be used as image tag
	hashCode := calculateHash(agentSrcPath)
	img = fmt.Sprintf("%s:%s", img, hashCode)

	client, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return "", false, fmt.Errorf("failed to create runtime client: %w", err)
	}
	defer containerclient.Close(ctx, client)

	imageList, err := client.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: img,
		}),
	})
	if err != nil {
		return "", false, fmt.Errorf("failed to retrieve the list of images from the conainer runtime host: %w", err)
	}

	if len(imageList) == 1 {
		log.Debug().Str("image", img).Msg("found image on runtime host")
		return img, false, nil
	}
	if len(imageList) > 1 {
		return "", false, fmt.Errorf("more than one image %q found on runtime host", img)
	}

	log.Debug().Str("image", img).Msg("image not found on runtime host")

	// build image
	err = buildImage(ctx, client, img, workspacePath, agentSourceDir, assets.AgentBuilderDockerfile, deleteBuildFolders)
	if err != nil {
		return "", false, fmt.Errorf("failed to build image %s: %w", img, err)
	}

	return img, true, nil
}

func buildImage(ctx context.Context, client *dockerclient.Client, img string, workspacePath string, agentSourceDir string, dockerFile []byte, deleteBuildFolders bool) error {
	log := zerolog.Ctx(ctx).With().Str("image", img).Logger()

	if err := os.WriteFile(path.Join(workspacePath, "Dockerfile"), dockerFile, util.OwnerCanReadWrite); err != nil {
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
		"AGENT_DIR": &agentSourceDir,
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
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer func() {
		if err := buildResp.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close image build response body")
		}
		log.Debug().Msg("closed image build response body")
	}()

	rd := bufio.NewReader(buildResp.Body)
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
				return fmt.Errorf("failed to unmarshal image build log line from image build response: %w", err)
			}

			if imageBuildLogLine.Error != nil {
				log.Error().Str("log_line", imageBuildLogLine.Error.Error()).Msg("image build")
				return errors.New("failed to build image")
			}
			if len(imageBuildLogLine.Stream) > 0 {
				log.Debug().Str("log_line", imageBuildLogLine.Stream).Msg("image build")
			}
			logLine = logLine[:0]
		}
	}

	log.Debug().Msg("successfully built image")
	return nil
}
