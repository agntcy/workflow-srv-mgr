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
	"github.com/docker/docker/pkg/idtools"
	"github.com/rs/zerolog"
)

// CreateBuildContext archive a dir and return an io.Reader
func CreateBuildContext(path string) (io.ReadCloser, error) {
	return archive.TarWithOptions(path, &archive.TarOptions{
		ExcludePatterns: []string{"**/.env", "**/.venv", "**/.git", "**/.github", "**/.idea", "**/.vscode"},
		ChownOpts:       &idtools.Identity{UID: 0, GID: 0},
	})
}

// if the container is running, return the public port or 0 otherwise
func IsContainerRunning(
	ctx context.Context,
	cntClient dockerclient.ContainerAPIClient,
	image string,
	containerName string) (int, error) {

	log := zerolog.Ctx(ctx).With().Str("container_name", containerName).Logger()

	cList, err := cntClient.ContainerList(ctx, container.ListOptions{
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
