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

type AgentDeploymentSpec struct {
	AgentDeploymentBuildSpec
	Dependencies []AgentDeploymentBuildSpec
}

type DeploymentArtifact []byte

type AgentDeploymentRunner interface {
	Deploy(ctx context.Context, agentDeploymentSpecs AgentDeploymentSpec, dryRun bool) (map[string]DeploymentArtifact, error)
	Remove(ctx context.Context, deploymentName string) error
	Logs(ctx context.Context, deploymentName string) error
	List(ctx context.Context, deploymentName string) error
}
