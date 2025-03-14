package manifest

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/cisco-eti/wfsm/manifests"
)

type ManifestService interface {
	Validate(ctx context.Context) error
	GetManifest(ctx context.Context) (manifests.AgentManifest, error)
}

type manifestService struct {
	filePath string
}

func NewManifestService(filePath string) ManifestService {
	return &manifestService{
		filePath: filePath,
	}
}

func (m manifestService) Validate(ctx context.Context) error {
	_, err := m.GetManifest(ctx)
	return err
}

func (m manifestService) GetManifest(ctx context.Context) (manifests.AgentManifest, error) {
	file, err := os.Open(m.filePath)
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}
	defer file.Close()

	// Read the file into a byte slice
	byteSlice, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}

	manifestJson := manifests.AgentManifest{}

	if err := json.Unmarshal(byteSlice, &manifestJson); err != nil {
		return manifests.AgentManifest{}, err
	}

	return manifestJson, nil
}
