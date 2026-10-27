package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/internal/flags"
	"github.com/wailsapp/wails/v3/internal/term"
	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

// crossDockerfileAsset is the embedded definition of the wails-cross image.
const crossDockerfileAsset = "build_assets/docker/Dockerfile.cross"

// SetupDocker obtains the wails-cross image used for cross-compiling Linux and
// Windows/macOS targets through Docker or Podman. It pulls the published image
// by default, or builds it locally from the embedded Dockerfile with -build.
func SetupDocker(options *flags.SetupDocker) error {
	DisableFooter = true
	return setupDocker(options.Build, exec.LookPath, runSetupCommand)
}

// setupDocker is SetupDocker with its host lookup and command runner injected.
func setupDocker(build bool, lookPath func(string) (string, error), run func(name string, arguments ...string) error) error {
	runtime, err := setupContainerRuntime(lookPath)
	if err != nil {
		return err
	}
	contextDir := ""
	if build {
		contextDir, err = writeCrossDockerfile()
		if err != nil {
			return err
		}
		defer os.RemoveAll(contextDir)
	}
	name, arguments := dockerSetupCommand(runtime, build, contextDir)
	term.Println("Running: " + name + " " + strings.Join(arguments, " "))
	if err := run(name, arguments...); err != nil {
		return fmt.Errorf("%s %s failed: %w", name, arguments[0], err)
	}
	return nil
}

// setupContainerRuntime picks the runtime the build planner uses: Docker when
// it is installed, otherwise Podman.
func setupContainerRuntime(lookPath func(string) (string, error)) (string, error) {
	for _, runtime := range []string{"docker", "podman"} {
		if _, err := lookPath(runtime); err == nil {
			return runtime, nil
		}
	}
	return "", fmt.Errorf("setup docker requires Docker or Podman on PATH")
}

// dockerSetupCommand returns the container CLI invocation that either pulls
// the published wails-cross image or builds it from the Dockerfile in
// contextDir under the local name the planner also accepts.
func dockerSetupCommand(runtime string, build bool, contextDir string) (string, []string) {
	if build {
		return runtime, []string{"build", "-t", "wails-cross", "-f", filepath.Join(contextDir, "Dockerfile.cross"), contextDir}
	}
	return runtime, []string{"pull", pipeline.CrossContainerImage}
}

// writeCrossDockerfile writes the embedded Dockerfile into a fresh temporary
// directory and returns that directory.
func writeCrossDockerfile() (string, error) {
	data, err := buildAssets.ReadFile(crossDockerfileAsset)
	if err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp("", "wails-cross-*")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile.cross"), data, 0o644); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

// runSetupCommand runs a container CLI command with its output attached to the
// terminal so pull and build progress stays visible.
func runSetupCommand(name string, arguments ...string) error {
	command := exec.Command(name, arguments...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
