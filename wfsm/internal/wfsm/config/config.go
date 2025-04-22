package config

import (
	"fmt"
	"os"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// loadConfig parses a config.yaml file and returns an AgentConfig.
func LoadConfig(path string) (ConfigFile, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return ConfigFile{}, fmt.Errorf("failed to read config file: %v", err)
	}

	var config ConfigFile
	if err := yaml.Unmarshal(file, &config); err != nil {
		return ConfigFile{}, fmt.Errorf("failed to unmarshal config file: %v", err)
	}

	return config, nil
}

// WriteConfig writes the given AgentConfig to a config.yaml file.
func WriteConfig(path string, config ConfigFile) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GenerateDefaultConfig generates a ConfigFile with default values for the given agent names.
func GenerateDefaultConfig(agentSpecs map[string]internal.AgentSpec, platform string, mainAgent string) (ConfigFile, error) {
	config := make(map[string]AgentConfig, len(agentSpecs))

	for name, _ := range agentSpecs {

		agentConfig := AgentConfig{
			APIKey: uuid.NewString(),
			ID:     uuid.NewString(),
		}

		switch platform {
		case internal.DOCKER:
			if name == mainAgent {
				port, err := util.GetNextAvailablePort()
				if err != nil {
					return ConfigFile{}, fmt.Errorf("failed to get next available port: %v", err)
				}
				agentConfig.Port = port
			}
		case internal.KUBERNETES:
			agentConfig.K8sConfig = internal.K8sConfig{
				StatefulSet: internal.StatefulSet{
					Replicas: 1,
				},
				Service: internal.Service{
					Labels: map[string]string{
						"app": name,
					},
				},
			}

			if name == mainAgent {
				agentConfig.K8sConfig.Service.Type = "NodePort"
			} else {
				agentConfig.K8sConfig.Service.Type = "ClusterIP"
			}

		}

		config[name] = agentConfig

	}

	return ConfigFile{
		Config: config,
	}, nil
}

func MergeConfigs(agentConfig, userConfig ConfigFile) ConfigFile {
	for key, userValue := range userConfig.Config {
		if agentValue, exists := agentConfig.Config[key]; exists {
			// Override existing fields in agentConfig with userConfig values
			if userValue.Port != 0 {
				agentValue.Port = userValue.Port
			}
			if userValue.APIKey != "" {
				agentValue.APIKey = userValue.APIKey
			}
			if len(userValue.EnvVars) > 0 {
				for envKey, envValue := range userValue.EnvVars {
					agentValue.EnvVars[envKey] = envValue
				}
			}
			// Merge K8sConfig
			for labelKey, labelValue := range userValue.K8sConfig.StatefulSet.Labels {
				agentValue.K8sConfig.StatefulSet.Labels[labelKey] = labelValue
			}
			for annotationKey, annotationValue := range userValue.K8sConfig.StatefulSet.Annotations {
				agentValue.K8sConfig.StatefulSet.Annotations[annotationKey] = annotationValue
			}
			for serviceLabelKey, serviceLabelValue := range userValue.K8sConfig.Service.Labels {
				agentValue.K8sConfig.Service.Labels[serviceLabelKey] = serviceLabelValue
			}
			for serviceAnnotationKey, serviceAnnotationValue := range userValue.K8sConfig.Service.Annotations {
				agentValue.K8sConfig.Service.Annotations[serviceAnnotationKey] = serviceAnnotationValue
			}
			if userValue.K8sConfig.Service.Type != "" {
				agentValue.K8sConfig.Service.Type = userValue.K8sConfig.Service.Type
			}
			agentConfig.Config[key] = agentValue
		} else {
			// Add new key-value pairs from userConfig to agentConfig
			agentConfig.Config[key] = userValue
		}
	}
	return agentConfig
}
