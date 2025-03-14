package deployer

import (
	"context"

	"github.com/cisco-eti/wfsm/internal/deployer/container"
	"github.com/cisco-eti/wfsm/internal/deployer/python"
	"github.com/cisco-eti/wfsm/manifests"
)

// AgentDeployer interface with deploy method
type AgentDeployer interface {
	Deploy(ctx context.Context) error
}

func GetAgentDeployer(agManifest manifests.AgentManifest,
	deploymentOption manifests.AgentDeploymentDeploymentOptionsInner,
	envFilePath string,
	hostStorageFolder string,
	deleteBuildFolders bool) AgentDeployer {
	if deploymentOption.DockerDeployment != nil {
		return container.NewContainerAgentDeployer(
			agManifest,
			envFilePath)
	} else if deploymentOption.SourceCodeDeployment != nil {
		return python.NewPythonAgentDeployer(
			agManifest,
			deploymentOption.SourceCodeDeployment,
			envFilePath,
			hostStorageFolder,
			deleteBuildFolders)
	}
	return nil
}
