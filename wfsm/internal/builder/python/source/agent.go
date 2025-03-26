// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package source

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/cisco-eti/wfsm/manifests"
)

// AgentSource interface with CopyToWorkspace method
type AgentSource interface {
	CopyToWorkspace(workspacePath string) error
}

func GetAgentSource(deployment *manifests.SourceCodeDeployment) (AgentSource, error) {
	parsedURL, err := url.Parse(deployment.Url)
	if err != nil {
		return nil, err
	}
	switch parsedURL.Scheme {
	case "http", "https", "":
		if parsedURL.Host == "github.com" {
			return &GoGetSource{
				URL: "git::" + deployment.Url,
			}, nil
		}
		if isLocalPath(deployment.Url) && !isZipOrTarball(deployment.Url) {
			return &LocalSource{
				LocalPath: deployment.Url,
			}, nil
		}
		return &GoGetSource{
			URL: deployment.Url,
		}, nil
	case "file":
		// remove file:// prefix from deployment URL
		deploymentUrl := strings.TrimPrefix(deployment.Url, "file://")
		if isZipOrTarball(deployment.Url) {
			return &GoGetSource{
				URL: deployment.Url,
			}, nil
		}
		if isLocalPath(deploymentUrl) {
			return &LocalSource{
				LocalPath: deploymentUrl,
			}, nil
		}

	}
	return nil, fmt.Errorf("unsupported source code deployment URL: %s", deployment.Url)
}

func isZipOrTarball(url string) bool {
	// Implement logic to check if the URL is a zip or tarball file
	return strings.HasSuffix(url, ".zip") || strings.HasSuffix(url, ".tar") || strings.HasSuffix(url, ".tar.gz")
}

func isLocalPath(url string) bool {
	// Implement logic to check if the URL is a local path
	_, err := os.Stat(url)
	return err == nil
}
