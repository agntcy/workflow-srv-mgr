// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package assets

import _ "embed"

//go:embed agent.Dockerfile
var AgentBuilderDockerfile []byte

//go:embed start_agws.sh
var StartAGWSScript []byte
