// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"context"
	"errors"
	"fmt"

	"github.com/cisco-eti/wfsm/manifests"
)

type ManifestService interface {
	Validate() error
	GetDeploymentOptionIdx(option *string) (int, error)
	GetManifest() manifests.AgentManifest
}

type ManifestLoader interface {
	loadManifest(context.Context) (manifests.AgentManifest, error)
}

type manifestService struct {
	manifestLoader ManifestLoader
	manifest       manifests.AgentManifest
}

func NewManifestService(ctx context.Context, manifestLoader ManifestLoader) (ManifestService, error) {
	manifest, err := manifestLoader.loadManifest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %s", err)
	}
	return &manifestService{
		manifest: manifest,
	}, nil
}

func (m manifestService) GetManifest() manifests.AgentManifest {
	return m.manifest
}

func (m manifestService) Validate() error {
	// validate ref name and version
	if m.manifest.Name == "" {
		return errors.New("invalid agent manifest: no name found in manifest")
	}
	if m.manifest.Version == "" {
		return errors.New("invalid agent manifest: no version found in manifest")
	}
	if len(m.manifest.Extensions) == 0 {
		return errors.New("invalid agent manifest: no deployment extensions found in manifest")
	}
	if !depoloymentExtensionIsPresent(m.manifest) {
		return errors.New("invalid agent manifest: no deployment extension 'oasf.agntcy.org/feature/runtime/manifest' found in manifest")
	}
	return m.ValidateDeploymentOptions()
}

func (m manifestService) ValidateDeploymentOptions() error {
	deployment := GetDeployment(m.manifest)
	if len(deployment.DeploymentOptions) == 0 {
		return errors.New("invalid agent manifest: no deployment option found in manifest")
	}
	return nil
}

func (m manifestService) GetDeploymentOptionIdx(option *string) (int, error) {
	if option == nil || len(*option) == 0 {
		return 0, nil
	}
	deployment := GetDeployment(m.manifest)
	for i, opt := range deployment.DeploymentOptions {
		if opt.SourceCodeDeployment != nil &&
			opt.SourceCodeDeployment.Name != nil &&
			*opt.SourceCodeDeployment.Name == *option {
			return i, nil
		}
		if opt.DockerDeployment != nil &&
			opt.DockerDeployment.Name != nil &&
			*opt.DockerDeployment.Name == *option {
			return i, nil
		}
	}
	return 0, fmt.Errorf("invalid agent manifest: deployment option %s not found", *option)
}

func depoloymentExtensionIsPresent(manifest manifests.AgentManifest) bool {
	for _, ext := range manifest.Extensions {
		if ext.Name == "oasf.agntcy.org/feature/runtime/manifest" {
			return true
		}
	}
	return false
}

func GetDeployment(manifest manifests.AgentManifest) manifests.AgentDeployment {
	for _, ext := range manifest.Extensions {
		if ext.Name == "oasf.agntcy.org/feature/runtime/manifest" {
			return ext.Data.Deployment
		}
	}
	return manifests.AgentDeployment{}
}
