package main

import (
	"os"

	"github.com/cisco-eti/wfsm/internal/wfsm/command"
)

// version is injected during build
var version = "n/a"
var workflowServerManager = command.NewWorkflowServerManager(version)
var exitFunc = os.Exit

func main() {
	exitCode := 0
	err := workflowServerManager.Execute()

	if err != nil {
		// error code 1 is cross platform compatible.
		exitCode = 1
	}

	exitFunc(exitCode)
}
