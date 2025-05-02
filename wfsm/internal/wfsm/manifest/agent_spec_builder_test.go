// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cisco-eti/wfsm/internal/wfsm/config"
	"github.com/stretchr/testify/assert"
)

func TestAgentSpecBuilder_BuildAgentSpec_WithEnvVars(t *testing.T) {
	tests := []struct {
		name                           string
		manifestPath                   string
		localEnv                       map[string]string
		envFilePath                    string
		config                         config.ConfigFile
		expectedAgentDeploymentEnvVars map[string]map[string]string
	}{
		{
			name:         "Test with agent_A_manifest.json",
			manifestPath: "test/manifest_2/agent_A_manifest.json",
			localEnv: map[string]string{
				// it's not declared in the manifest --> will be ignored
				"ENV_VAR_AGENT_A_1": "env_var_value_from_local_env",
				// will be overridden by env file
				"AGENT_A_ENV_VAR_AGENT_A": "env_var_value_from_local_env",
				// this is expected to be set
				"AGENT_A_ENV_VAR_AGENT_A_2": "env_var_value_from_local_env",
			},
			envFilePath: "test/manifest_2/env-vars",
			config: config.ConfigFile{
				Config: map[string]config.AgentConfig{},
			},
			expectedAgentDeploymentEnvVars: map[string]map[string]string{
				"agent_A": {
					"ENV_VAR_AGENT_A":   "env_var_value_agent_a_override",
					"ENV_VAR_AGENT_A_2": "env_var_value_from_local_env",
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
		{
			name:         "Test with agent_A_manifest.json with env file",
			manifestPath: "test/manifest_2/agent_A_manifest.json",
			envFilePath:  "test/manifest_2/env-vars",
			config: config.ConfigFile{
				Config: map[string]config.AgentConfig{},
			},
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
		{
			name:         "Test with agent_A_manifest.json with config",
			manifestPath: "test/manifest_2/agent_A_manifest.json",
			envFilePath:  "test/manifest_2/env-vars",
			config: config.ConfigFile{
				Config: map[string]config.AgentConfig{
					"agent_A": {
						EnvVars: map[string]string{
							"ENV_VAR_AGENT_A": "env_var_value_agent_a_override_from_config",
						},
					},
					"agent_B_1": {
						EnvVars: map[string]string{
							"ENV_VAR_AGENT_B_2": "env_var_value_agent_b_override_from_config",
						},
					},
				},
			},
			expectedAgentDeploymentEnvVars: map[string]map[string]string{
				"agent_A": {
					"ENV_VAR_AGENT_A": "env_var_value_agent_a_override_from_config",
				},
				"agent_B_1": {
					"ENV_VAR_AGENT_B_1": "env_var_value_agent_b_1_override",
					"ENV_VAR_AGENT_B_2": "env_var_value_agent_b_override_from_config",
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
			envFile, err := LoadEnvVars(tt.envFilePath)
			assert.NoError(t, err, "LoadEnvVars should not return an error")

			builder := NewAgentSpecBuilder()
			err = builder.BuildAgentSpec(context.Background(), tt.manifestPath, "", nil, nil)
			assert.NoError(t, err, "BuildAgentSpec should not return an error")

			if tt.localEnv != nil {
				setLocalEnvVars(tt.localEnv)
			}
			builder.LoadFromConfig(context.Background(), tt.config, envFile)

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
	envFilePath := "test/manifest_3/env-vars"

	envFile, err := LoadEnvVars(envFilePath)
	assert.NoError(t, err, "LoadEnvVars should not return an error")

	builder := NewAgentSpecBuilder()
	err = builder.BuildAgentSpec(context.Background(), manifestPath, "", nil, nil)
	assert.NoError(t, err, "BuildAgentSpec should not return an error")

	builder.LoadFromConfig(context.Background(), config.ConfigFile{
		Config: map[string]config.AgentConfig{},
	}, envFile)
	errs := builder.ValidateEnvVars(context.Background())

	assert.Len(t, errs, 1, "ValidateEnvVars should return one error")
}

func TestAgentSpecBuilder_Deployment_Name_Non_Unique(t *testing.T) {

	manifestPath := "test/manifest_4/agent_A_manifest.json"

	builder := NewAgentSpecBuilder()
	err := builder.BuildAgentSpec(context.Background(), manifestPath, "", nil, nil)

	assert.Error(t, err, "BuildAgentSpec should return an error")
	assert.Contains(t, err.Error(), "agent deployment name must be unique: agent_C_1")
}

func setLocalEnvVars(envVars map[string]string) {
	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			panic(fmt.Sprintf("failed to set local env var %s: %v", key, err))
		}
	}
}

func TestAgentSpecBuilder_NormalizeManifestPath(t *testing.T) {
	type args struct {
		manifestPath      string
		dependencyRefPath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "manifest file reference is empty",
			args: args{
				manifestPath:      "",
				dependencyRefPath: "/hurricane.json",
			},
			want: "/hurricane.json",
		},
		{
			name: "dependency file reference is absolute",
			args: args{
				manifestPath:      "/etwc/agent/agent_A_manifest.json",
				dependencyRefPath: "/hurricane.json",
			},
			want: "/hurricane.json",
		},
		{
			name: "dependency file reference is relative to the manifest reference",
			args: args{
				manifestPath:      "/etwc/agent/agent_A_manifest.json",
				dependencyRefPath: "hurricane.json",
			},
			want: "/etwc/agent/hurricane.json",
		},
		{
			name: "dependency file reference is relative to the manifest reference",
			args: args{
				manifestPath:      "/etwc/agent/agent_A_manifest.json",
				dependencyRefPath: "./hurricane.json",
			},
			want: "/etwc/agent/hurricane.json",
		},
		{
			name: "dependency file reference is relative to the manifest reference, up one level (./../)",
			args: args{
				manifestPath:      "/etwc/agent/agent_A_manifest.json",
				dependencyRefPath: "./../hurricane.json",
			},
			want: "/etwc/hurricane.json",
		},
		{
			name: "dependency file reference is relative to the manifest reference, up one level( ../)",
			args: args{
				manifestPath:      "/etwc/agent/agent_A_manifest.json",
				dependencyRefPath: "../hurricane.json",
			},
			want: "/etwc/hurricane.json",
		},
		{
			name: "dependency file reference is not local",
			args: args{
				manifestPath:      "/etwc/agent/agent_A_manifest.json",
				dependencyRefPath: "http://example.com/hurricane.json",
			},
			want: "http://example.com/hurricane.json",
		},
		{
			name: "dependency file reference has file:// scheme",
			args: args{
				manifestPath:      "/etwc/agent/agent_A_manifest.json",
				dependencyRefPath: "file://./",
			},
			want: "/etwc/agent",
		},
		{
			name: "dependency file reference has file:// scheme, + relative path",
			args: args{
				manifestPath:      "/etwc/agent/agent_A_manifest.json",
				dependencyRefPath: "file://./hurricane.json",
			},
			want: "/etwc/agent/hurricane.json",
		},
		{
			name: "dependency file reference has file:// scheme, + relative path, + up one level",
			args: args{
				manifestPath:      "/etwc/agent/agent_A_manifest.json",
				dependencyRefPath: "file://./../hurricane.json",
			},
			want: "/etwc/hurricane.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AgentSpecBuilder{}
			got, err := a.NormalizeDependencyRef(tt.args.manifestPath, tt.args.dependencyRefPath)
			assert.NoError(t, err, "NormalizeDependencyRef should not return an error")
			assert.Equalf(t, tt.want, got, "NormalizeDependencyRef(%v, %v)", tt.args.manifestPath, tt.args.dependencyRefPath)
		})
	}
}
