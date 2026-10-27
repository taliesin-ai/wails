package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// runCLIHelp runs the command tree with args and returns what it printed to
// stdout. clir exposes no command introspection, so routing is asserted through
// the help text each command prints.
func runCLIHelp(t *testing.T, args ...string) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = writer
	runErr := newApp(nil).Run(args...)
	os.Stdout = stdout
	require.NoError(t, writer.Close())
	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	require.NoError(t, runErr)
	return string(output)
}

func TestSetupSubcommandsAreChildrenOfSetup(t *testing.T) {
	docker := runCLIHelp(t, "setup", "docker", "-help")
	require.Contains(t, docker, "wails setup docker - Pull or build the wails-cross image")
	require.Contains(t, docker, "-build")

	signing := runCLIHelp(t, "setup", "signing", "-help")
	require.Contains(t, signing, "wails setup signing - Configure code signing")

	entitlements := runCLIHelp(t, "setup", "entitlements", "-help")
	require.Contains(t, entitlements, "wails setup entitlements - Configure macOS entitlements")

	setupHelp := runCLIHelp(t, "setup", "-help")
	require.Contains(t, setupHelp, "docker")
	require.Contains(t, setupHelp, "signing")
	require.Contains(t, setupHelp, "entitlements")
}

func TestTopLevelDockerIsNotRegisteredAndLegacySigningStillWorks(t *testing.T) {
	rootHelp := runCLIHelp(t, "-help")
	require.Contains(t, rootHelp, "setup")
	require.NotContains(t, rootHelp, "Pull or build the wails-cross image")

	legacy := runCLIHelp(t, "signing", "-help")
	require.Contains(t, legacy, "Configure code signing")
}
