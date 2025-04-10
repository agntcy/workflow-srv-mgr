// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	hubClient "github.com/agntcy/dir/cli/hub/client"

	"github.com/agntcy/dir/api/hub/v1alpha1"
	"github.com/cisco-eti/wfsm/manifests"
	"github.com/cisco-eti/wfsm/oasf"
	"google.golang.org/grpc/metadata"
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

type fileManifestLoader struct {
	filePath string
}
type hubManifestLoader struct {
	accessToken string
	digest      string
	host        string
}
type directoryManifestLoader struct{}

func loaderFactory(path string) (ManifestLoader, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid manifest path: %s", err)
	}
	if u.Scheme == "file" || u.Scheme == "" {
		return &fileManifestLoader{
			filePath: strings.TrimPrefix(path, "file://"),
		}, nil
	}
	if u.Scheme == "hub" {
		fmt.Printf("HUB_LOADER\n\n\n")
		accessToken := os.Getenv("ACCESS_TOKEN")
		if accessToken == "" {
			return nil, fmt.Errorf("access token is not set")
		}
		return &hubManifestLoader{
			accessToken: accessToken,
			digest:      strings.TrimLeft(u.Path, "/"),
			host:        u.Host,
		}, nil
	}
	if u.Scheme == "sha256" {
		return &directoryManifestLoader{}, nil
	}

	return nil, fmt.Errorf("unsupported manifest path: %s", u.Scheme)
}

func NewManifestService(ctx context.Context, path string) (ManifestService, error) {
	manifestLoader, err := loaderFactory(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest loader: %s", err)
	}
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
	if option == nil || len(*option) == 0 {
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

func (f *fileManifestLoader) loadManifest(ctx context.Context) (manifests.AgentManifest, error) {
	file, err := os.Open(f.filePath)
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

func (f *hubManifestLoader) loadManifest(ctx context.Context) (manifests.AgentManifest, error) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+f.accessToken))
	hc, err := hubClient.New(f.host)
	agentID := &v1alpha1.AgentIdentifier{
		Id: &v1alpha1.AgentIdentifier_Digest{
			Digest: f.digest,
		},
	}

	dirManifest, err := hc.PullAgent(ctx, &v1alpha1.PullAgentRequest{
		Id: agentID,
	})
	if err != nil {
		return manifests.AgentManifest{}, fmt.Errorf("failed to pull agent: %v", err)
	}

	var hubManifest oasf.OasfJson
	err = hubManifest.UnmarshalJSON(dirManifest)
	if err != nil {
		return manifests.AgentManifest{}, fmt.Errorf("failed to unmarshal file: %s", err)
	}

	agentSpec := hubManifest.Extensions[0]["specs"]

	agentByte, err := json.Marshal(agentSpec)
	if err != nil {
		return manifests.AgentManifest{}, fmt.Errorf("failed to marshal agentSpec file: %s", err)
	}
	var agentManifest manifests.AgentManifest
	if err := json.Unmarshal(agentByte, &agentManifest); err != nil {
		return manifests.AgentManifest{}, fmt.Errorf("failed to unmarshal agentManifest file: %s", err)
	}

	// ref, ok := hubManifest.Extensions[0]["name"].(string)
	// if !ok {
	// 	return manifests.AgentManifest{}, fmt.Errorf("failed to get name from manifest: %s", err)
	// }
	version, ok := hubManifest.Extensions[0]["version"].(string)
	if !ok {
		return manifests.AgentManifest{}, fmt.Errorf("failed to get version from manifest: %s", err)
	}
	agentManifest.Metadata.Ref.Name = hubManifest.Name
	agentManifest.Metadata.Ref.Version = version

	return agentManifest, nil
}

func (f *directoryManifestLoader) loadManifest(ctx context.Context) (manifests.AgentManifest, error) {
	//TODO: implement directory manifest loader
	return manifests.AgentManifest{}, nil
}
