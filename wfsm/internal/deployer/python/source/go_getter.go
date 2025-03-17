package source

import "github.com/hashicorp/go-getter"

// GoGetSource struct implementing AgentSource interface
type GoGetSource struct {
	URL string
}

// CopyToWorkspace copies all files from sourcePath to workspacePath
func (ls *GoGetSource) CopyToWorkspace(workspacePath string) error {
	err := getter.Get(workspacePath, ls.URL)
	return err
}
