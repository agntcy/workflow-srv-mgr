// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package platforms

import (
	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/platforms/docker"
	"github.com/cisco-eti/wfsm/internal/platforms/k8s"
)

func GetPlatformRunner(platform string, hostStorageFolder string) internal.AgentDeploymentRunner {
	switch platform {
	case "k8s":
		return k8s.NewK8sRunner(hostStorageFolder)
	case "docker":
		return docker.NewDockerComposeRunner(hostStorageFolder)
	}
	return nil
}
