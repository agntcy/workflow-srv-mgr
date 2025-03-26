// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentSpecBuilder_BuildAgentSpec_WithEnvVars(t *testing.T) {
	tests := []struct {
		name                           string
		manifestPath                   string
		envFilePath                    string
		expectedAgentDeploymentEnvVars map[string]map[string]string
	}{
		{
			name:         "Test with agent_A_manifest.json",
			manifestPath: "test/manifest_2/agent_A_manifest.json",
			envFilePath:  "test/manifest_2/env-vars.yaml",
			expectedAgentDeploymentEnvVars: map[string]map[string]string{
				"agent_A": {
					"ENV_VAR_AGENT_A": "env_var_value_agent_a_override",
				},
				"agent_B_1": {
					"ENV_VAR_AGENT_B_1": "env_var_value_agent_b_1_override",
					"ENV_VAR_AGENT_B_2": "env_var_value_agent_b_a2",
				},
				"agent_C_1": {
					"ENV_VAR_AGENT_C_1": "env_var_value_agent_c_1_override",
					"ENV_VAR_AGENT_C_2": "env_var_value_agent_c_a2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVarValues, err := LoadEnvVars(tt.envFilePath)
			assert.NoError(t, err, "LoadEnvVars should not return an error")

			builder := NewAgentSpecBuilder()
			err = builder.BuildAgentSpec(context.Background(), tt.manifestPath, "", nil, envVarValues)
			assert.NoError(t, err, "BuildAgentSpec should not return an error")

			for agentDepName, envVarValues := range tt.expectedAgentDeploymentEnvVars {
				for key, value := range envVarValues {
					assert.Equal(t, value, builder.AgentSpecs[agentDepName].EnvVars[key], fmt.Sprintf("EnvVar %s value should match", key))
				}

			}
		})
	}
}

func TestAgentSpecBuilder_Required_Env_Var_Missing(t *testing.T) {

	manifestPath := "test/manifest_3/agent_A_manifest.json"
	envFilePath := "test/manifest_3/env-vars.yaml"

	envVarValues, err := LoadEnvVars(envFilePath)
	assert.NoError(t, err, "LoadEnvVars should not return an error")

	builder := NewAgentSpecBuilder()
	err = builder.BuildAgentSpec(context.Background(), manifestPath, "", nil, envVarValues)

	assert.Error(t, err, "BuildAgentSpec should return an error")
	assert.Contains(t, err.Error(), "missing required env var")

}

func TestAgentSpecBuilder_Deployment_Name_Non_Unique(t *testing.T) {

	manifestPath := "test/manifest_4/agent_A_manifest.json"
	envFilePath := "test/manifest_4/env-vars.yaml"

	envVarValues, err := LoadEnvVars(envFilePath)
	assert.NoError(t, err, "LoadEnvVars should not return an error")

	builder := NewAgentSpecBuilder()
	err = builder.BuildAgentSpec(context.Background(), manifestPath, "", nil, envVarValues)

	assert.Error(t, err, "BuildAgentSpec should return an error")
	assert.Contains(t, err.Error(), "agent deployment name must be unique: agent_C_1")
}
