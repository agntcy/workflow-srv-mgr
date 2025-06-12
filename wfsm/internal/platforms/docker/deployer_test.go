// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package docker

import (
	"context"
	"os"
	"testing"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/cisco-eti/wfsm/manifests"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestRunner_Deploy_DryRun(t *testing.T) {
	ctx := context.Background()
	version := "1.0.0"

	// Mock data for agent deployment specs
	agentDeploymentSpecs := map[string]internal.AgentDeploymentBuildSpec{
		"test-agent-A": {
			AgentSpec: internal.AgentSpec{
				Port:           62173,
				AgentID:        "d8084dc6-52c4-4316-8460-8f43b64db17a",
				ApiKey:         "4a69e02d-b03a-47e4-99ab-f0782be35f62",
				DeploymentName: "test-agent-A",
				EnvVars: map[string]string{
					"ENV_VAR_AGENT_A": "valueA",
				},
				Manifest: manifests.AgentManifest{
					Extensions: []manifests.Manifest{
						{
							Name:    "schema.oasf.agntcy.org/features/runtime/manifest",
							Version: &version,
							Data: manifests.DeploymentManifest{
								Deployment: manifests.AgentDeployment{
									DeploymentOptions: []manifests.AgentDeploymentDeploymentOptionsInner{
										{
											SourceCodeDeployment: &manifests.SourceCodeDeployment{
												FrameworkConfig: manifests.SourceCodeDeploymentFrameworkConfig{
													LangGraphConfig: &manifests.LangGraphConfig{
														Graph: "agentA.graph",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Image:       "test-agent-a-image",
			ServiceName: "test-agent-a-service",
		},
		"test-agent-B": {
			AgentSpec: internal.AgentSpec{
				AgentID:        "39c8d1ab-d155-440c-aa4c-7b2d244d1c09",
				ApiKey:         "657425ba-fc18-4a6d-9144-14e6a79fdcf4",
				DeploymentName: "test-agent-B",
				EnvVars: map[string]string{
					"ENV_VAR_AGENT_B": "valueB",
				},
				Manifest: manifests.AgentManifest{
					Extensions: []manifests.Manifest{
						{
							Name:    "schema.oasf.agntcy.org/features/runtime/manifest",
							Version: &version,
							Data: manifests.DeploymentManifest{
								Deployment: manifests.AgentDeployment{
									DeploymentOptions: []manifests.AgentDeploymentDeploymentOptionsInner{
										{
											SourceCodeDeployment: &manifests.SourceCodeDeployment{
												FrameworkConfig: manifests.SourceCodeDeploymentFrameworkConfig{
													LangGraphConfig: &manifests.LangGraphConfig{
														Graph: "agentB.graph",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Image:       "test-agent-b-image",
			ServiceName: "test-agent-b-service",
		},
	}

	// Mock dependencies
	dependencies := map[string][]string{
		"test-agent-A": {"test-agent-B"},
	}

	hostStorageFolder := ".wfsm"
	if _, err := os.Stat(hostStorageFolder); os.IsNotExist(err) {
		if err := os.Mkdir(hostStorageFolder, 0755); err != nil {
			t.Errorf("failed to create host storage folder for agent: %v", err)
		}
	}
	defer os.RemoveAll(hostStorageFolder)

	// Create a runner instance
	r := &runner{
		hostStorageFolder: hostStorageFolder,
	}

	// Call the Deploy function with dryRun = true
	artifact, err := r.Deploy(ctx, "test-agent-A", agentDeploymentSpecs, dependencies, true)

	// Validate the results
	assert.NoError(t, err, "Deploy should not return an error")
	assert.NotNil(t, artifact, "DeploymentArtifact should not be nil")

	// Read the expected artifact from the test/expected_compose.yaml file
	expectedArtifact, err := os.ReadFile("test/expected_compose.yaml")
	assert.NoError(t, err, "Failed to read test/expected_compose.yaml")

	// Unmarshal the expected artifact
	var expectedArtifactData map[string]interface{}
	err = yaml.Unmarshal(expectedArtifact, &expectedArtifactData)
	assert.NoError(t, err, "Failed to unmarshal expected artifact")

	// Unmarshal the actual artifact
	var actualArtifactData map[string]interface{}
	err = yaml.Unmarshal(artifact, &actualArtifactData)
	assert.NoError(t, err, "Failed to unmarshal actual artifact")

	// Compare the actual artifact to the expected artifact
	assert.Equal(t, expectedArtifactData, actualArtifactData, "The actual artifact should match the expected artifact")
}
