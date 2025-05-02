// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package internal

import (
	"context"

	"github.com/cisco-eti/wfsm/manifests"
)

const (
	KUBERNETES       = "k8s"
	DOCKER           = "docker"
	DEFAULT_API_PORT = 8000
)

type AgentSpec struct {
	DeploymentName           string
	Manifest                 manifests.AgentManifest
	SelectedDeploymentOption int
	EnvVars                  map[string]string
	AgentID                  string
	ApiKey                   string
	Port                     int
	K8sConfig                K8sConfig
	ManifestPath             string
}
type K8sConfig struct {
	EnvVarsFromSecret string      `yaml:"envVarsFromSecret"`
	StatefulSet       StatefulSet `yaml:"statefulset"`
	Service           Service     `yaml:"service"`
}

type Service struct {
	Type        string            `yaml:"type,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type StatefulSet struct {
	Replicas       int               `yaml:"replicas,omitempty"`
	Labels         map[string]string `yaml:"labels,omitempty"`
	Annotations    map[string]string `yaml:"annotations,omitempty"`
	PodAnnotations map[string]string `yaml:"podAnnotations,omitempty"`
	Resources      Resources         `yaml:"resources,omitempty"`
	NodeSelector   map[string]string `yaml:"nodeSelector,omitempty"`
	Affinity       Affinity          `yaml:"affinity,omitempty"`
	Tolerations    []Toleration      `yaml:"tolerations,omitempty"`
}

type Resources struct {
	Requests map[string]string `yaml:"requests,omitempty"`
	Limits   map[string]string `yaml:"limits,omitempty"`
}

type Affinity struct {
	NodeAffinity NodeAffinity `yaml:"nodeAffinity,omitempty"`
}

type NodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution RequiredDuringSchedulingIgnoredDuringExecution `yaml:"requiredDuringSchedulingIgnoredDuringExecution"`
}

type RequiredDuringSchedulingIgnoredDuringExecution struct {
	NodeSelectorTerms []NodeSelectorTerm `yaml:"nodeSelectorTerms"`
}

type NodeSelectorTerm struct {
	MatchExpressions []MatchExpression `yaml:"matchExpressions"`
}

type MatchExpression struct {
	Key      string   `yaml:"key"`
	Operator string   `yaml:"operator"`
	Values   []string `yaml:"values"`
}

type Toleration struct {
	Key      string `yaml:"key"`
	Operator string `yaml:"operator"`
	Effect   string `yaml:"effect"`
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
}

type AgentDeploymentRunner interface {
	Deploy(ctx context.Context, deploymentName string, agentDeploymentSpecs map[string]AgentDeploymentBuildSpec, dependencies map[string][]string, dryRun bool) (DeploymentArtifact, error)
	Remove(ctx context.Context, deploymentName string) error
	Logs(ctx context.Context, deploymentName string, agentNames []string) error
	List(ctx context.Context, deploymentName string) error
}
