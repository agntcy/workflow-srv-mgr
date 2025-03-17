package assets

import _ "embed"

//go:embed agent.Dockerfile
var AgentBuilderDockerfile []byte

//go:embed workflowserver.Dockerfile
var WorkflowServerDockerfile []byte
