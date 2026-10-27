package commands

import (
	"fmt"

	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

// windowsToolExecutable returns the resolved path of a Windows toolchain
// executable (makensis, MakeAppx.exe, signtool.exe). When the tool cannot be
// found it returns the bare name so the exec error names the missing tool.
func windowsToolExecutable(name string) string {
	if path, ok := pipeline.WindowsToolPath(name); ok {
		return path
	}
	return name
}

// windowsToolMissing is the error for a required Windows toolchain executable
// that is neither on PATH nor in its standard install location.
func windowsToolMissing(name, installHint string) error {
	return fmt.Errorf("%s is not found on PATH or in its standard install location; %s", name, installHint)
}
