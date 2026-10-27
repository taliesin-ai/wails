package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

// argumentAfter returns the value following flag in arguments, or "" when the
// flag is absent.
func argumentAfter(arguments []string, flag string) string {
	for index, argument := range arguments {
		if argument == flag && index+1 < len(arguments) {
			return arguments[index+1]
		}
	}
	return ""
}

func TestCompileGoArgsAddsWindowsGUISubsystemToProductionBuilds(t *testing.T) {
	_, args := compileGoArgs(pipeline.CompileSpec{TargetOS: "windows", TargetArch: "amd64", Production: true, Strip: true})
	assert.Equal(t, "-w -s -H=windowsgui", argumentAfter(args, "-ldflags"))

	_, args = compileGoArgs(pipeline.CompileSpec{TargetOS: "windows", TargetArch: "arm64", Production: true, LinkerFlags: []string{"-X", "main.v=1"}})
	assert.Equal(t, "-X main.v=1 -H=windowsgui", argumentAfter(args, "-ldflags"))
}

func TestCompileGoArgsLeavesDevelopmentAndOtherPlatformsOnTheConsoleSubsystem(t *testing.T) {
	_, args := compileGoArgs(pipeline.CompileSpec{TargetOS: "windows", TargetArch: "amd64", Production: false, Strip: true})
	assert.Equal(t, "-w -s", argumentAfter(args, "-ldflags"))

	_, args = compileGoArgs(pipeline.CompileSpec{TargetOS: "linux", TargetArch: "amd64", Production: true, Strip: true})
	assert.Equal(t, "-w -s", argumentAfter(args, "-ldflags"))
}

func TestCompileGoArgsDoesNotDuplicateUserWindowsGUIFlag(t *testing.T) {
	for _, linker := range [][]string{{"-H=windowsgui"}, {"-H", "windowsgui"}, {"-s", "-H windowsgui"}} {
		_, args := compileGoArgs(pipeline.CompileSpec{TargetOS: "windows", TargetArch: "amd64", Production: true, LinkerFlags: linker})
		ldflags := argumentAfter(args, "-ldflags")
		assert.Equal(t, 1, strings.Count(ldflags, "windowsgui"), "linker flags %q produced %q", linker, ldflags)
		assert.Equal(t, 1, strings.Count(ldflags, "-H"), "linker flags %q produced %q", linker, ldflags)
	}
}

func TestLinkerFlagsSetWindowsGUIRecognisesBothSpellings(t *testing.T) {
	assert.True(t, linkerFlagsSetWindowsGUI([]string{"-H=windowsgui"}))
	assert.True(t, linkerFlagsSetWindowsGUI([]string{"-H", "windowsgui"}))
	assert.False(t, linkerFlagsSetWindowsGUI([]string{"-H", "linux"}))
	assert.False(t, linkerFlagsSetWindowsGUI([]string{"-windowsgui"}))
}

func TestNativeDarwinCompileExportsDeploymentTargetByDefault(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture invokes Unix shell tools")
	}
	root := t.TempDir()
	tools := t.TempDir()
	record := filepath.Join(root, "go-env.txt")
	script := "#!/bin/sh\n/usr/bin/env > \"$GO_RECORD\"\noutput=\nwhile [ \"$#\" -gt 0 ]; do if [ \"$1\" = \"-o\" ]; then output=$2; shift; fi; shift; done\nmkdir -p \"$(dirname \"$output\")\"\nprintf binary > \"$output\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(tools, "go"), []byte(script), 0o755))
	t.Setenv("PATH", tools)
	t.Setenv("GO_RECORD", record)
	previousHost := manifestHostOS
	manifestHostOS = "darwin"
	t.Cleanup(func() { manifestHostOS = previousHost })

	_, err := (&manifestHandler{root: root}).compile(t.Context(), pipeline.CompileSpec{
		TargetOS: "darwin", TargetArch: "arm64", Output: ".wails/build/app", Toolchain: "native", MinimumVersion: "12.0",
	})
	require.NoError(t, err)
	env := readTestFile(t, record)
	assert.Contains(t, env, "\nMACOSX_DEPLOYMENT_TARGET=12.0\n")
	assert.Contains(t, env, "\nCGO_CFLAGS=-mmacosx-version-min=12.0\n")
	assert.Contains(t, env, "\nCGO_LDFLAGS=-mmacosx-version-min=12.0\n")
	assert.Contains(t, env, "\nCGO_ENABLED=1\n")
}

func TestNativeWindowsCrossCompileDisablesCGOOffWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture invokes Unix shell tools")
	}
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n\ngo 1.23\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "windows"), 0o755))
	icon, err := buildAssets.ReadFile("build_assets/windows/icon.ico")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "windows", "icon.ico"), icon, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "windows", "wails.exe.manifest"), []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0"></assembly>`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "windows", "info.json"), []byte(`{"fixed":{"file_version":"1.0.0.0"},"info":{"0000":{"ProductVersion":"1.0.0"}}}`), 0o644))
	tools := t.TempDir()
	record := filepath.Join(root, "go-env.txt")
	script := "#!/bin/sh\n/usr/bin/env > \"$GO_RECORD\"\noutput=\nwhile [ \"$#\" -gt 0 ]; do if [ \"$1\" = \"-o\" ]; then output=$2; shift; fi; shift; done\nmkdir -p \"$(dirname \"$output\")\"\nprintf binary > \"$output\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(tools, "go"), []byte(script), 0o755))
	t.Setenv("PATH", tools)
	t.Setenv("GO_RECORD", record)
	previousHost := manifestHostOS
	manifestHostOS = "darwin"
	t.Cleanup(func() { manifestHostOS = previousHost })

	_, err = (&manifestHandler{root: root}).compile(t.Context(), pipeline.CompileSpec{
		TargetOS: "windows", TargetArch: "amd64", Output: ".wails/build/app.exe", Toolchain: "native", Production: true,
	})
	require.NoError(t, err)
	env := readTestFile(t, record)
	assert.Contains(t, env, "\nCGO_ENABLED=0\n")
	assert.Contains(t, env, "\nGOOS=windows\n")
}

func TestDockerCacheArgumentsMountModuleAndLinuxBuildCaches(t *testing.T) {
	arguments := dockerCacheArguments("linux", "/home/dev/go/pkg/mod", "/home/dev/.cache/go-build", "")
	assert.Equal(t, "/home/dev/go/pkg/mod:/go/pkg/mod", argumentAfter(arguments, "-v"))
	assert.Contains(t, arguments, "/home/dev/.cache/go-build:/tmp/go-build")
	assert.Contains(t, arguments, "GOMODCACHE=/go/pkg/mod")
	assert.Contains(t, arguments, "GOCACHE=/tmp/go-build")

	arguments = dockerCacheArguments("darwin", "/Users/dev/go/pkg/mod", "/Users/dev/Library/Caches/go-build", "")
	assert.Contains(t, arguments, "/Users/dev/go/pkg/mod:/go/pkg/mod")
	assert.NotContains(t, arguments, "/Users/dev/Library/Caches/go-build:/tmp/go-build", "build cache stays container-local off Linux")

	arguments = dockerCacheArguments("linux", "/src/mod", "", ":Z")
	assert.Contains(t, arguments, "/src/mod:/go/pkg/mod:Z")
	assert.Len(t, arguments, 6, "no build cache mount without a host cache directory")
}

func TestDockerIdentityArgumentsRunAsInvokingUserOnLinuxOnly(t *testing.T) {
	assert.Equal(t, []string{"--user", "1000:1001", "-e", "HOME=/tmp"}, dockerIdentityArguments("linux", 1000, 1001))
	assert.Empty(t, dockerIdentityArguments("darwin", 501, 20))
	assert.Empty(t, dockerIdentityArguments("windows", 0, 0))
}

func TestDockerCompileArgumentsIncludeCachesIdentityAndCFlags(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture invokes Unix shell tools; native Windows paths are tested separately")
	}
	root := t.TempDir()
	tools := t.TempDir()
	record := filepath.Join(root, "docker-args.txt")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$DOCKER_RECORD\"\noutput=\nwhile [ \"$#\" -gt 0 ]; do if [ \"$1\" = \"-o\" ]; then output=$2; shift; fi; shift; done\nmkdir -p \"$(dirname \"$output\")\"\nprintf binary > \"$output\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(tools, "docker"), []byte(script), 0o755))
	t.Setenv("PATH", tools)
	t.Setenv("DOCKER_RECORD", record)
	previousHost := manifestHostOS
	manifestHostOS = "linux"
	t.Cleanup(func() { manifestHostOS = previousHost })

	_, err := (&manifestHandler{root: root}).compile(t.Context(), pipeline.CompileSpec{
		TargetOS: "darwin", TargetArch: "arm64", Output: ".wails/build/app", Toolchain: "docker", MinimumVersion: "12.0", Production: true,
	})
	require.NoError(t, err)
	arguments := strings.Split(strings.TrimSpace(readTestFile(t, record)), "\n")
	assert.Contains(t, arguments, "--user")
	assert.Contains(t, arguments, "GOMODCACHE=/go/pkg/mod")
	assert.Contains(t, arguments, "GOCACHE=/tmp/go-build")
	assert.Contains(t, arguments, "CGO_CFLAGS=-w -mmacosx-version-min=12.0")
	assert.Contains(t, arguments, "MACOSX_DEPLOYMENT_TARGET=12.0")
}

func TestSetupDockerCommandPullsOrBuildsTheCrossImage(t *testing.T) {
	name, arguments := dockerSetupCommand("docker", false, "")
	assert.Equal(t, "docker", name)
	assert.Equal(t, []string{"pull", "ghcr.io/wailsapp/wails-cross:latest"}, arguments)

	name, arguments = dockerSetupCommand("podman", true, "/tmp/ctx")
	assert.Equal(t, "podman", name)
	assert.Equal(t, []string{"build", "-t", "wails-cross", "-f", filepath.Join("/tmp/ctx", "Dockerfile.cross"), "/tmp/ctx"}, arguments)
}

func TestSetupContainerRuntimePrefersDockerThenPodman(t *testing.T) {
	onlyPodman := func(name string) (string, error) {
		if name == "podman" {
			return "/usr/bin/podman", nil
		}
		return "", os.ErrNotExist
	}
	runtime, err := setupContainerRuntime(onlyPodman)
	require.NoError(t, err)
	assert.Equal(t, "podman", runtime)

	everything := func(name string) (string, error) { return "/usr/bin/" + name, nil }
	runtime, err = setupContainerRuntime(everything)
	require.NoError(t, err)
	assert.Equal(t, "docker", runtime)

	_, err = setupContainerRuntime(func(string) (string, error) { return "", os.ErrNotExist })
	assert.ErrorContains(t, err, "requires Docker or Podman")
}

func TestSetupDockerRunsThroughInjectedRunner(t *testing.T) {
	haveDocker := func(name string) (string, error) {
		if name == "docker" {
			return "/usr/bin/docker", nil
		}
		return "", os.ErrNotExist
	}

	var calls [][]string
	record := func(name string, arguments ...string) error {
		calls = append(calls, append([]string{name}, arguments...))
		return nil
	}
	require.NoError(t, setupDocker(false, haveDocker, record))
	assert.Equal(t, [][]string{{"docker", "pull", "ghcr.io/wailsapp/wails-cross:latest"}}, calls)

	calls = nil
	var contextFile string
	buildRunner := func(name string, arguments ...string) error {
		calls = append(calls, append([]string{name}, arguments...))
		contextFile = arguments[4]
		data, err := os.ReadFile(contextFile)
		if assert.NoError(t, err) {
			assert.Contains(t, string(data), "FROM golang")
		}
		return nil
	}
	require.NoError(t, setupDocker(true, haveDocker, buildRunner))
	require.Len(t, calls, 1)
	assert.Equal(t, "docker", calls[0][0])
	assert.Equal(t, "build", calls[0][1])
	assert.Equal(t, "wails-cross", calls[0][3])
	_, statErr := os.Stat(filepath.Dir(contextFile))
	assert.True(t, os.IsNotExist(statErr), "the temporary build context is removed after the build")

	failing := func(string, ...string) error { return os.ErrPermission }
	err := setupDocker(false, haveDocker, failing)
	assert.ErrorContains(t, err, "docker pull failed")
}
