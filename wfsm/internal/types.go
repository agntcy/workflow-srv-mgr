package internal

import (
	"context"

	"github.com/cisco-eti/wfsm/manifests"
)

type AgentSpec struct {
	DeploymentName           string
	Manifest                 manifests.AgentManifest
	SelectedDeploymentOption int
	EnvVars                  map[string]string
}

type AgentDeploymentBuildSpec struct {
	AgentSpec
	Image       string
	ServiceName string
}

type DeploymentArtifact []byte

// AgentDeploymentBuilder interface with deploy method
type AgentDeploymentBuilder interface {
	Build(ctx context.Context, inputSpec AgentSpec) (AgentDeploymentBuildSpec, error)
	Validate(ctx context.Context, inputSpec AgentSpec) error
}

type AgentDeploymentRunner interface {
	Deploy(ctx context.Context, deploymentName string, agentDeploymentSpecs map[string]AgentDeploymentBuildSpec, dependencies map[string][]string, dryRun bool) (DeploymentArtifact, error)

	Remove(ctx context.Context, deploymentName string) error
	Logs(ctx context.Context, deploymentName string) error
	List(ctx context.Context, deploymentName string) error
}
