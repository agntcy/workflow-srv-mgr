// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package container_client

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	imagespecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rs/zerolog"
)

// CreateBuildContext archive a dir and return an io.Reader
func CreateBuildContext(path string) (io.ReadCloser, error) {
	return archive.Tar(path, archive.Uncompressed)
}

func Close(ctx context.Context, client *dockerclient.Client) {
	log := zerolog.Ctx(ctx)

	if err := client.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close container_client runtime client")
	}
	log.Debug().Msg("closed container_client runtime client")
}

func CreateContainer(
	ctx context.Context,
	client *dockerclient.Client,
	containerConfig *container.Config,
	hostConfig *container.HostConfig,
	platform *imagespecv1.Platform,
	containerName string) (string, error) {
	log := zerolog.Ctx(ctx).With().Str("container_name", containerName).Logger()

	createResp, err := client.ContainerCreate(
		ctx, containerConfig,
		hostConfig,
		nil,
		platform,
		containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container_client: %w", err)
	}

	if len(createResp.Warnings) == 0 {
		log.Debug().Msg("created container_client")
	} else {
		log.Warn().Strs("warnings", createResp.Warnings).Msg("created container_client with warnings")
	}
	return createResp.ID, nil
}

// if the container is running, return the public port or 0 otherwise
func IsContainerRunning(
	ctx context.Context,
	client *dockerclient.Client,
	image string,
	containerName string) (int, error) {

	log := zerolog.Ctx(ctx).With().Str("container_name", containerName).Logger()

	cList, err := client.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "name",
			Value: containerName,
		}),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list containers: %w", err)
	}

	if len(cList) > 0 {
		for _, c := range cList {
			log.Debug().Msg("container_client found")
			publicPort := getPublicPort(c.Ports)
			return publicPort, nil
		}

		log.Info().Msg("container_client already running")
		return 0, nil
	}

	log.Debug().Msg("container_client not found")
	return 0, nil
}

func getPublicPort(ports []container.Port) int {
	if len(ports) == 0 {
		return 0
	}
	for _, port := range ports {
		if port.IP == "0.0.0.0" && port.PublicPort != 0 {
			return int(port.PublicPort)
		}
	}
	return 0
}

func RemoveContainer(ctx context.Context, client *dockerclient.Client, containerID string) error {
	log := zerolog.Ctx(ctx)

	err := client.ContainerRemove(ctx, containerID, container.RemoveOptions{RemoveVolumes: true})
	if err != nil {
		return err
	}
	log.Debug().Msg("removed container_client")
	return nil
}

func StartContainer(
	ctx context.Context,
	client *dockerclient.Client,
	containerID string) error {
	log := zerolog.Ctx(ctx)

	err := client.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container_client: %w", err)
	}
	log.Debug().Msg("started container_client")

	return nil
}

func CopyToContainer(ctx context.Context, client *dockerclient.Client, containerID, src, dst string) error {
	log := zerolog.Ctx(ctx)

	appSrc, err := archive.Tar(src, archive.Uncompressed)
	if err != nil {
		return fmt.Errorf("failed to tar source dir: %w", err)
	}
	defer func() {
		if err = appSrc.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close archived source reader")
		}
	}()

	err = client.CopyToContainer(ctx, containerID, dst, appSrc, container.CopyToContainerOptions{})
	if err != nil {
		return fmt.Errorf("failed to archive to container_client: %w", err)
	}

	return nil
}
