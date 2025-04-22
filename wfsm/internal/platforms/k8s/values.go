package k8s

import (
	"github.com/cisco-eti/wfsm/internal"
)

type ChartValues struct {
	Agents []AgentValues `yaml:"agents"`
}

type AgentValues struct {
	Name               string               `yaml:"name"`
	Image              Image                `yaml:"image"`
	Labels             map[string]string    `yaml:"labels,omitempty"`
	Env                []EnvVar             `yaml:"env"`
	SecretEnvs         []EnvVar             `yaml:"secretEnvs"`
	ExistingSecretName string               `yaml:"existingSecretName,omitempty"`
	VolumePath         string               `yaml:"volumePath,omitempty"`
	ExternalPort       int                  `yaml:"externalPort"`
	InternalPort       int                  `yaml:"internalPort"`
	Service            internal.Service     `yaml:"service"`
	StatefulSet        internal.StatefulSet `yaml:"statefulset"`
}

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
}

type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
