// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package util

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/hashicorp/go-version"
)

const OwnerCanReadWrite = 0777
const DevTagSuffix = "-dev"

func CurrentArchToDockerPlatform() string {
	if runtime.GOARCH == "amd64" {
		return "linux/amd64"
	}
	if runtime.GOARCH == "arm64" {
		return "linux/arm64"
	}
	// TODO: we need to see if other arch and OS needs to be supported
	return ""
}

func GetNextAvailablePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	addr := ln.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

func GetHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return usr.HomeDir, nil
}

func GetDockerCLI(ctx context.Context) (*command.DockerCli, error) {
	dockerCli, err := command.NewDockerCli(command.WithBaseContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create docker cli: %v", err)
	}
	clientOptions := flags.ClientOptions{
		LogLevel:  "debug",
		TLS:       false,
		TLSVerify: false,
	}
	err = dockerCli.Initialize(&clientOptions)
	return dockerCli, err
}

// UntarGzFile extracts a .tar.gz file to the specified destination folder.
func UntarGzFile(src []byte, dest string) error {
	reader := bytes.NewReader(src)

	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

func NormalizeAgentName(name string) string {
	// replace all non-alphanumeric characters with space
	re := regexp.MustCompile(`[^a-z0-9-]+`)
	return re.ReplaceAllString(strings.ToLower(name), "-")
}

func GetDockerComposeProjectName(name string) string {
	// replace all non-alphanumeric characters with space
	re := regexp.MustCompile(`[^a-z0-9-_]+`)
	return re.ReplaceAllString(strings.ToLower(name), "")
}

func CalculateEnvVarPrefix(agName string) string {
	prefix := strings.ToUpper(agName)
	// replace all non-alphanumeric characters with _
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	prefix = re.ReplaceAllString(prefix, "_")
	return prefix + "_"
}

func SplitImageName(fullImageName string) (string, string) {
	parts := strings.Split(fullImageName, ":")
	if len(parts) == 1 {
		return parts[0], "latest"
	}
	return parts[0], parts[1]
}

func GetLatestTag(tags []string) (string, error) {
	var versions []*version.Version
	for _, tag := range tags {
		v, err := version.NewVersion(tag)
		if v != nil && !IsDevTag(tag) && err == nil {
			versions = append(versions, v)
		}
	}
	if versions == nil {
		return "", fmt.Errorf("no valid tags found")
	}
	sort.Sort(version.Collection(versions))
	return versions[len(versions)-1].Original(), nil
}

func IsDevTag(tag string) bool {
	if strings.Contains(tag, DevTagSuffix) {
		return true
	}
	return false
}

func MergeMaps(dest map[string]string, src map[string]string) map[string]string {
	if dest == nil {
		dest = make(map[string]string)
	}
	for key, value := range src {
		dest[key] = value
	}
	return dest
}
