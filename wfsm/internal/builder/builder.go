// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package builder

import (
	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/builder/container"
	"github.com/cisco-eti/wfsm/internal/builder/python"
	"github.com/cisco-eti/wfsm/manifests"
)

func GetAgentBuilder(deploymentOption manifests.AgentDeploymentDeploymentOptionsInner, deleteBuildFolders bool) internal.AgentDeploymentBuilder {
	if deploymentOption.DockerDeployment != nil {
		return container.NewContainerAgentBuilder()
	} else if deploymentOption.SourceCodeDeployment != nil {
		return python.NewPythonAgentBuilder(deleteBuildFolders)
	}
	return nil
}
