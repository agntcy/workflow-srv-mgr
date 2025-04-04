// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cisco-eti/wfsm/manifests"
)

type ManifestService interface {
	Validate() error
	GetDeploymentOptionIdx(option *string) (int, error)
	GetManifest() manifests.AgentManifest
}

type manifestService struct {
	manifest manifests.AgentManifest
}

func NewManifestService(filePath string) (ManifestService, error) {
	manifest, err := loadManifest(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %s", err)
	}
	return &manifestService{
		manifest: manifest,
	}, nil
}

func (m manifestService) Validate() error {
	// validate ref name and version
	if m.manifest.Metadata.Ref.Name == "" {
		return errors.New("invalid agent manifest: no name found in manifest")
	}
	if m.manifest.Metadata.Ref.Version == "" {
		return errors.New("invalid agent manifest: no version found in manifest")
	}
	return m.ValidateDeploymentOptions()
}

func (m manifestService) ValidateDeploymentOptions() error {
	deployment := m.manifest.Deployment
	if deployment == nil {
		return errors.New("invalid agent manifest: no deployment found in manifest")
	}
	if len(deployment.DeploymentOptions) == 0 {
		return errors.New("invalid agent manifest: no deployment option found in manifest")
	}
	return nil
}

func (m manifestService) GetDeploymentOptionIdx(option *string) (int, error) {
	if len(*option) == 0 {
		return 0, nil
	}
	for i, opt := range m.manifest.Deployment.DeploymentOptions {
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

func loadManifest(filePath string) (manifests.AgentManifest, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}
	defer file.Close()

	// Read the file into a byte slice
	byteSlice, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}

	manifest := manifests.AgentManifest{}

	if err := json.Unmarshal(byteSlice, &manifest); err != nil {
		return manifests.AgentManifest{}, err
	}

	return manifest, nil
}

func (m manifestService) GetManifest() manifests.AgentManifest {
	return m.manifest
}
