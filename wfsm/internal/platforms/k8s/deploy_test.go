package k8s

import (
	"context"
	"os"
	"testing"

	"github.com/cisco-eti/wfsm/internal"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDeploy_DryRun_GeneratesExpectedOutput(t *testing.T) {
	// Arrange
	runner := NewK8sRunner("/tmp")
	mainAgentName := "mailcomposer"
	agentDeploymentSpecs := map[string]internal.AgentDeploymentBuildSpec{
		"mailcomposer": {
			AgentSpec: internal.AgentSpec{
				AgentID: "1141b40c-8278-495f-9d0a-680d64573bae",
				ApiKey:  "aa15dbbe-e9c7-4d05-a750-464e7c8bfed1",
				EnvVars: map[string]string{
					"AZURE_OPENAI_API_KEY":  "xxxxxxx",
					"AZURE_OPENAI_ENDPOINT": "https://smith-project-agents.openai.azure.com",
					"OPENAI_API_VERSION":    "2024-07-01-preview",
					"AZURE_OPENAI_MODEL":    "gpt-4o-mini",
				},
				K8sConfig: internal.K8sConfig{
					Service: internal.Service{
						Type: "NodePort",
						Labels: map[string]string{
							"app": "mailcomposer",
						},
					},
					StatefulSet: internal.StatefulSet{
						Replicas: 1,
					},
				},
			},
			ServiceName: "mailcomposer",
			Image:       "agntcy/wfsm-mailcomposer:latest",
		},
		"email-reviewer-1": {
			AgentSpec: internal.AgentSpec{
				AgentID: "7f1d1e05-64c1-4a13-ac78-f470a1fc2b5f",
				ApiKey:  "76653017-d5b1-4f8f-b752-6392ee93dc8f",
				EnvVars: map[string]string{
					"AZURE_OPENAI_API_KEY":  "xxxxxxx",
					"AZURE_OPENAI_ENDPOINT": "https://smith-project-agents.openai.azure.com",
					"OPENAI_API_VERSION":    "2024-07-01-preview",
					"AZURE_OPENAI_MODEL":    "gpt-4o-mini",
					"TEST_ENV_VAR":          "some test value",
				},
				K8sConfig: internal.K8sConfig{
					Service: internal.Service{
						Type: "ClusterIP",
						Labels: map[string]string{
							"app": "email_reviewer_1",
						},
					},
					StatefulSet: internal.StatefulSet{
						Replicas: 1,
					},
				},
			},
			ServiceName: "email-reviewer-1",
			Image:       "agntcy/wfsm-email-reviewer:latest",
		},
	}
	dependencies := map[string][]string{
		"mailcomposer": {"email-reviewer-1"},
	}
	dryRun := true

	// Act
	output, err := runner.Deploy(context.Background(), mainAgentName, agentDeploymentSpecs, dependencies, 0, dryRun)

	// Assert
	assert.NoError(t, err)

	// Load expected values from test/expected_values.yaml
	expectedData, err := os.ReadFile("test/expected_values.yaml")
	assert.NoError(t, err)

	var expectedValues map[string]interface{}
	err = yaml.Unmarshal(expectedData, &expectedValues)
	assert.NoError(t, err)

	var actualValues map[string]interface{}
	err = yaml.Unmarshal(output, &actualValues)
	assert.NoError(t, err)

	// Compare EnvVars separately to ensure order independence
	expectedAgents := expectedValues["agents"].([]interface{})
	actualAgents := actualValues["agents"].([]interface{})

	for i := range expectedAgents {
		expectedAgent := expectedAgents[i].(map[string]interface{})
		actualAgent := actualAgents[i].(map[string]interface{})

		// Compare EnvVars as maps
		expectedEnvVars := make(map[string]string)
		actualEnvVars := make(map[string]string)

		for _, env := range expectedAgent["env"].([]interface{}) {
			envMap := env.(map[string]interface{})
			expectedEnvVars[envMap["name"].(string)] = envMap["value"].(string)
		}

		for _, env := range actualAgent["env"].([]interface{}) {
			envMap := env.(map[string]interface{})
			actualEnvVars[envMap["name"].(string)] = envMap["value"].(string)
		}

		assert.Equal(t, expectedEnvVars, actualEnvVars, "EnvVars do not match for agent: %v", expectedAgent["name"])

		// Remove EnvVars from comparison to avoid duplicate checks
		delete(expectedAgent, "env")
		delete(actualAgent, "env")

		// Compare the rest of the agent fields
		assert.Equal(t, expectedAgent, actualAgent, "Agent fields do not match for agent: %v", expectedAgent["name"])
	}
}

// TestCalculateConfigHash tests the calculateConfigHash function.
func TestCalculateConfigHash(t *testing.T) {
	// Arrange
	input1 := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	input2 := map[string]string{
		"key3": "value3",
		"key4": "value4",
	}
	expectedHash := calculateConfigHash(input1, input2)

	// Act
	actualHash := calculateConfigHash(input1, input2)

	// Assert
	assert.Equal(t, expectedHash, actualHash, "Hashes should match for the same input maps")

	// Test with reordered keys to ensure order independence
	input1Reordered := map[string]string{
		"key2": "value2",
		"key1": "value1",
	}
	actualHashReordered := calculateConfigHash(input1Reordered, input2)
	assert.Equal(t, expectedHash, actualHashReordered, "Hashes should match regardless of key order")

	// Test with different input to ensure hash changes
	input3 := map[string]string{
		"key1": "value1",
		"key2": "differentValue",
	}
	differentHash := calculateConfigHash(input3, input2)
	assert.NotEqual(t, expectedHash, differentHash, "Hashes should differ for different input maps")
}
