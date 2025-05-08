package config

import "github.com/cisco-eti/wfsm/internal"

type ConfigFile struct {
	Config map[string]AgentConfig `yaml:"config"`
}

type AgentConfig struct {
	Port      int                 `yaml:"port"`
	APIKey    string              `yaml:"apiKey"`
	ID        string              `yaml:"id"`
	EnvVars   map[string]string   `yaml:"envVars"`
	K8sConfig *internal.K8sConfig `yaml:"k8s,omitempty"`
}
