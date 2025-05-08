package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/internal/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
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
func GenerateDefaultConfig(agentSpecs map[string]internal.AgentSpec, platform string, mainAgent string, envFile map[string]string) (ConfigFile, error) {
	config := make(map[string]AgentConfig, len(agentSpecs))

	for name, _ := range agentSpecs {

		agentConfig := AgentConfig{
			EnvVars: map[string]string{},
		}

		id := getEnvVarValue(name, "ID", envFile)
		if id == "" {
			id = uuid.NewString()
		}
		agentConfig.ID = id

		apiKey := getEnvVarValue(name, "API_KEY", envFile)
		if apiKey == "" {
			apiKey = uuid.NewString()
		}
		agentConfig.APIKey = apiKey

		// Get default port from env var if any
		portStr := getEnvVarValue(name, "PORT", envFile)
		if portStr != "" {
			portI, err := strconv.Atoi(portStr)
			if err != nil {
				return ConfigFile{}, errors.New("invalid port specified in environment variable: PORT")
			}
			agentConfig.Port = portI
		}

		switch platform {
		case internal.DOCKER:
			if name == mainAgent {
				// if no port is specified get next available
				if agentConfig.Port == 0 {
					freePort, err := util.GetNextAvailablePort()
					if err != nil {
						return ConfigFile{}, fmt.Errorf("failed to get next available port: %v", err)
					}
					agentConfig.Port = freePort
				}
			}
		case internal.KUBERNETES:
			if agentConfig.Port == 0 {
				agentConfig.Port = internal.DEFAULT_API_PORT
			}
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

func getEnvVarValue(agentName string, envVarName string, envFile map[string]string) string {
	agentPrefix := util.CalculateEnvVarPrefix(agentName)
	if value, ok := envFile[agentPrefix+envVarName]; ok {
		return value
	}
	if value := os.Getenv(agentPrefix + envVarName); value != "" {
		return value
	}
	return ""
}

func PrintConfig(ctx context.Context, file ConfigFile) error {
	data, err := yaml.Marshal(file)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	log := zerolog.Ctx(ctx)
	log.Info().Msgf("CONFIG: \n%s", (string(data)))
	return nil
}

func MergeConfigs(agentConfig, userConfig ConfigFile, platform string) ConfigFile {
	for key, userValue := range userConfig.Config {
		if agentValue, exists := agentConfig.Config[key]; exists {
			// Override existing fields in agentConfig with userConfig values
			if userValue.Port != 0 {
				agentValue.Port = userValue.Port
			}
			if userValue.APIKey != "" {
				agentValue.APIKey = userValue.APIKey
			}
			if userValue.ID != "" {
				agentValue.ID = userValue.ID
			}
			if len(userValue.EnvVars) > 0 {
				for envKey, envValue := range userValue.EnvVars {
					agentValue.EnvVars[envKey] = envValue
				}
			}
			if platform == internal.KUBERNETES {
				agentValue = mergeK8sConfigs(agentValue, userValue)
			}
			agentConfig.Config[key] = agentValue
		} else {
			// Add new key-value pairs from userConfig to agentConfig
			agentConfig.Config[key] = userValue
		}
	}
	return agentConfig
}

func mergeK8sConfigs(agentValue AgentConfig, userValue AgentConfig) AgentConfig {
	if userValue.K8sConfig.EnvVarsFromSecret != "" {
		agentValue.K8sConfig.EnvVarsFromSecret = userValue.K8sConfig.EnvVarsFromSecret
	}
	// Merge K8sConfig.StatefulSet
	for labelKey, labelValue := range userValue.K8sConfig.StatefulSet.Labels {
		agentValue.K8sConfig.StatefulSet.Labels[labelKey] = labelValue
	}
	for annotationKey, annotationValue := range userValue.K8sConfig.StatefulSet.Annotations {
		agentValue.K8sConfig.StatefulSet.Annotations[annotationKey] = annotationValue
	}
	for podAnnotationKey, podAnnotationValue := range userValue.K8sConfig.StatefulSet.PodAnnotations {
		agentValue.K8sConfig.StatefulSet.PodAnnotations[podAnnotationKey] = podAnnotationValue
	}
	if userValue.K8sConfig.StatefulSet.Replicas != 0 {
		agentValue.K8sConfig.StatefulSet.Replicas = userValue.K8sConfig.StatefulSet.Replicas
	}
	if userValue.K8sConfig.StatefulSet.NodeSelector != nil {
		agentValue.K8sConfig.StatefulSet.NodeSelector = userValue.K8sConfig.StatefulSet.NodeSelector
	}
	if userValue.K8sConfig.StatefulSet.Tolerations != nil {
		agentValue.K8sConfig.StatefulSet.Tolerations = userValue.K8sConfig.StatefulSet.Tolerations
	}
	agentValue.K8sConfig.StatefulSet.Affinity = userValue.K8sConfig.StatefulSet.Affinity
	agentValue.K8sConfig.StatefulSet.Resources = userValue.K8sConfig.StatefulSet.Resources

	// Merge K8sConfig.Service
	for serviceLabelKey, serviceLabelValue := range userValue.K8sConfig.Service.Labels {
		agentValue.K8sConfig.Service.Labels[serviceLabelKey] = serviceLabelValue
	}
	for serviceAnnotationKey, serviceAnnotationValue := range userValue.K8sConfig.Service.Annotations {
		agentValue.K8sConfig.Service.Annotations[serviceAnnotationKey] = serviceAnnotationValue
	}
	if userValue.K8sConfig.Service.Type != "" {
		agentValue.K8sConfig.Service.Type = userValue.K8sConfig.Service.Type
	}
	return agentValue
}
