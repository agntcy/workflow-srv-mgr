// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package util

import (
	"context"
	"fmt"
	"net"
	"os/user"
	"runtime"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
)

const OwnerCanReadWrite = 0777

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
