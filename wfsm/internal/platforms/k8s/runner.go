// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package k8s

import (
	"context"

	"github.com/cisco-eti/wfsm/internal"
)

// NewK8sRunner implementation of AgentDeploymentRunner
type runner struct {
	hostStorageFolder string
}

func NewK8sRunner(hostStorageFolder string) internal.AgentDeploymentRunner {
	return &runner{
		hostStorageFolder: hostStorageFolder,
	}
}

func (r *runner) Remove(ctx context.Context, deploymentName string) error {
	return nil
}

func (r *runner) Logs(ctx context.Context, deploymentName string, agentNames []string) error {
	return nil
}

func (r *runner) List(ctx context.Context, deploymentName string) error {
	return nil
}
