package source

import (
	"os"
	"path/filepath"
)

// LocalSource struct implementing AgentSource interface
type LocalSource struct {
	LocalPath string
}

// CopyToWorkspace copies all files from sourcePath to workspacePath
func (ls *LocalSource) CopyToWorkspace(workspacePath string) error {
	// Copy all files from sourcePath to workspacePath
	return copyDir(ls.LocalPath, workspacePath)
}

func copyDir(src string, dest string) error {
	// Read all files and directories from src
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Ensure dest exists
	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return err
	}

	// Copy each file and directory to dest
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			// Recursively copy directory
			err = copyDir(srcPath, destPath)
			if err != nil {
				return err
			}
		} else {
			// Copy file
			err = copyFile(srcPath, destPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src string, dest string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dest, input, fileInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}
