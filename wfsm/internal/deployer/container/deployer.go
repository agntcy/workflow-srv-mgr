package container

import (
	"context"

	"github.com/cisco-eti/wfsm/manifests"
)

// deployer implementation of AgentDeployer
type deployer struct {
	agManifest  manifests.AgentManifest
	envFilePath string
}

func NewContainerAgentDeployer(agManifest manifests.AgentManifest, envFilePath string) *deployer {
	return &deployer{
		agManifest:  agManifest,
		envFilePath: envFilePath,
	}
}

func (d *deployer) Deploy(ctx context.Context) error {
	// run container_client with agent
	return nil
}
