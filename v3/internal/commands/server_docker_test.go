package commands

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/flags"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

func loadServerTestProject(t *testing.T) (string, *manifest.Loaded) {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, manifest.WriteMinimal(root, manifest.Project{Name: "app", ProductName: "App", Identifier: "com.example.app", Version: "1.0.0"}))
	loaded, err := manifest.Load(root, "")
	require.NoError(t, err)
	return root, loaded
}

func TestServerDockerFrontendRequestTargetsLinuxServer(t *testing.T) {
	_, loaded := loadServerTestProject(t)
	request := serverDockerFrontendRequest(loaded)
	assert.Equal(t, "build", request.Verb)
	assert.Equal(t, "linux", request.TargetOS)
	assert.Equal(t, "amd64", request.TargetArch)
	assert.True(t, request.Server)
	assert.False(t, request.Development)
	assert.Same(t, loaded, request.Loaded)
}

func TestServerDockerBuildArgs(t *testing.T) {
	options := NewServerDockerOptions()
	assert.Equal(t, []string{
		"build",
		"--build-arg", "CGO_ENABLED=0",
		"--build-arg", "GO_IMAGE=golang:alpine",
		"--build-arg", "RUNTIME_IMAGE=gcr.io/distroless/static-debian12",
		"-t", "app:latest",
		"-f", ".wails/docker/Dockerfile.server",
		".",
	}, serverDockerBuildArgs("app:latest", ".wails/docker/Dockerfile.server", &options))

	options.CGO = true
	options.GoImage = "golang:bookworm"
	options.RuntimeImage = "gcr.io/distroless/base-debian12"
	args := serverDockerBuildArgs("svc:v2", "build/docker/Dockerfile.server", &options)
	assert.Contains(t, args, "CGO_ENABLED=1")
	assert.Contains(t, args, "GO_IMAGE=golang:bookworm")
	assert.Contains(t, args, "RUNTIME_IMAGE=gcr.io/distroless/base-debian12")
	assert.Equal(t, "svc:v2", args[len(args)-4])
}

func TestServerDockerRunArgs(t *testing.T) {
	assert.Equal(t, []string{"run", "--rm", "-p", "9000:8080", "app:latest"}, serverDockerRunArgs("app:latest", 9000))
	assert.Equal(t, []string{"run", "--rm", "-p", "8080:8080", "app:latest"}, serverDockerRunArgs("app:latest", serverContainerPort))
}

func TestNewServerDockerOptionsDefaults(t *testing.T) {
	options := NewServerDockerOptions()
	assert.Equal(t, "golang:alpine", options.GoImage)
	assert.Equal(t, "gcr.io/distroless/static-debian12", options.RuntimeImage)
	assert.Equal(t, 8080, options.Port)
	assert.False(t, options.CGO)
	assert.False(t, options.Run)
	assert.Empty(t, options.Tag)
}

func TestResolveServerDockerfilePrefersLegacyUserFile(t *testing.T) {
	root, _ := loadServerTestProject(t)
	legacy := filepath.Join(root, "build", "docker", "Dockerfile.server")
	require.NoError(t, os.MkdirAll(filepath.Dir(legacy), 0o755))
	require.NoError(t, os.WriteFile(legacy, []byte("FROM scratch\n"), 0o644))

	path, err := resolveServerDockerfile(root, manifest.Frontend{Directory: "frontend", OutputDirectory: "dist"})
	require.NoError(t, err)
	assert.Equal(t, "build/docker/Dockerfile.server", path)
	_, statErr := os.Stat(filepath.Join(root, ".wails", "docker", "Dockerfile.server"))
	assert.ErrorIs(t, statErr, os.ErrNotExist, "a user-owned Dockerfile must not be shadowed by a generated copy")
}

func TestResolveServerDockerfileWritesEmbeddedTemplate(t *testing.T) {
	root, _ := loadServerTestProject(t)
	path, err := resolveServerDockerfile(root, manifest.Frontend{Directory: "frontend", OutputDirectory: "dist"})
	require.NoError(t, err)
	assert.Equal(t, ".wails/docker/Dockerfile.server", path)
	content, err := os.ReadFile(filepath.Join(root, ".wails", "docker", "Dockerfile.server"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "ARG RUNTIME_IMAGE=gcr.io/distroless/static-debian12")
	assert.Contains(t, string(content), "COPY --from=builder /app/frontend/dist /frontend/dist")
}

func TestRenderServerDockerfileFollowsFrontendLayout(t *testing.T) {
	template := "x\n" + serverFrontendCopyLine + "\ny\n"
	assert.Equal(t, template, renderServerDockerfile(template, manifest.Frontend{Directory: "frontend", OutputDirectory: "dist"}))
	assert.Equal(t, "x\n"+"COPY --from=builder /app/ui/build /frontend/dist"+"\ny\n",
		renderServerDockerfile(template, manifest.Frontend{Directory: "ui", OutputDirectory: "build"}))
	assert.Equal(t, "x\n"+"COPY --from=builder /app/frontend/dist /frontend/dist"+"\ny\n",
		renderServerDockerfile(template, manifest.Frontend{Directory: "frontend", OutputDirectory: "frontend/dist"}))
	headless := renderServerDockerfile(template, manifest.Frontend{Disabled: true})
	assert.False(t, strings.Contains(headless, "/frontend/dist"), "a headless project has no frontend output to copy")
	assert.Contains(t, headless, "x\n")
	assert.Contains(t, headless, "y\n")
}

type serverDockerCall struct {
	kind string
	args []string
}

func TestServerDockerRunsPreflightBuildThenImage(t *testing.T) {
	_, loaded := loadServerTestProject(t)
	root := loaded.Config.Root
	var calls []serverDockerCall
	operations := serverDockerOperations{
		probe: func(dir string, args ...string) (string, error) {
			calls = append(calls, serverDockerCall{"probe", append([]string{dir}, args...)})
			return "", nil
		},
		buildFrontend: func(ctx context.Context, l *manifest.Loaded) error {
			calls = append(calls, serverDockerCall{kind: "frontend"})
			return nil
		},
		docker: func(ctx context.Context, dir string, args ...string) error {
			calls = append(calls, serverDockerCall{"docker", append([]string{dir}, args...)})
			return nil
		},
	}
	options := NewServerDockerOptions()
	options.Run = true
	options.Port = 9000
	require.NoError(t, serverDockerWithOperations(context.Background(), &options, loaded, operations))

	require.Len(t, calls, 4)
	assert.Equal(t, serverDockerCall{"probe", []string{root, "info"}}, calls[0])
	assert.Equal(t, serverDockerCall{kind: "frontend"}, calls[1])
	assert.Equal(t, "docker", calls[2].kind)
	assert.Equal(t, root, calls[2].args[0])
	assert.Equal(t, serverDockerBuildArgs("app:latest", ".wails/docker/Dockerfile.server", &options), calls[2].args[1:])
	assert.Equal(t, []string{root, "run", "--rm", "-p", "9000:8080", "app:latest"}, calls[3].args)
}

func TestServerDockerFailsBeforeBuildingWhenDockerIsUnavailable(t *testing.T) {
	_, loaded := loadServerTestProject(t)
	built := false
	operations := serverDockerOperations{
		probe: func(string, ...string) (string, error) { return "", errors.New("cannot connect") },
		buildFrontend: func(context.Context, *manifest.Loaded) error {
			built = true
			return nil
		},
		docker: func(context.Context, string, ...string) error {
			t.Fatal("docker build must not run when the daemon is unreachable")
			return nil
		},
	}
	options := NewServerDockerOptions()
	err := serverDockerWithOperations(context.Background(), &options, loaded, operations)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "`docker info` failed")
	assert.False(t, built)
	_, statErr := os.Stat(filepath.Join(loaded.Config.Root, ".wails"))
	assert.ErrorIs(t, statErr, os.ErrNotExist, "no generated files are written when Docker is unavailable")
}

func TestServerDockerReportsFailedImageBuild(t *testing.T) {
	_, loaded := loadServerTestProject(t)
	operations := serverDockerOperations{
		probe:         func(string, ...string) (string, error) { return "", nil },
		buildFrontend: func(context.Context, *manifest.Loaded) error { return nil },
		docker: func(_ context.Context, _ string, args ...string) error {
			return errors.New("exit status 1")
		},
	}
	options := NewServerDockerOptions()
	err := serverDockerWithOperations(context.Background(), &options, loaded, operations)
	assert.ErrorContains(t, err, "docker build failed")
}

func TestServerDockerRejectsArguments(t *testing.T) {
	options := NewServerDockerOptions()
	err := ServerDocker(&options, []string{"extra"})
	assert.ErrorContains(t, err, "takes no arguments")
	options.Port = 70000
	err = ServerDocker(&options, nil)
	assert.ErrorContains(t, err, "-port")
}

func TestRunServerSelectsDevelopmentServerBuild(t *testing.T) {
	build := manifestRunOptions{Verb: "run"}
	require.NoError(t, configureServerRun("linux", &build))
	assert.True(t, build.Server)
	assert.True(t, build.Development, "run --server uses the development policy like the Taskfile run:server task")

	assert.ErrorContains(t, configureServerRun("ios", &manifestRunOptions{}), "desktop targets")
	assert.ErrorContains(t, configureServerRun("android", &manifestRunOptions{}), "desktop targets")
}

func TestBuildServerRequiresWakeProject(t *testing.T) {
	ran := false
	operations := manifestCommandOperations{
		active: func() (bool, error) { return false, nil },
		run:    func(manifestRunOptions) error { ran = true; return nil },
		task:   func(string, []string) error { t.Fatal("legacy task must not run for --server"); return nil },
	}
	err := buildWithOperations(&flags.Build{Server: true}, nil, operations)
	assert.ErrorContains(t, err, "--server requires a wails.yaml project")
	assert.False(t, ran)
}

func TestBuildServerPassesServerRequestToPlanner(t *testing.T) {
	var got manifestRunOptions
	operations := manifestCommandOperations{
		active: func() (bool, error) { return true, nil },
		run:    func(options manifestRunOptions) error { got = options; return nil },
		task:   func(string, []string) error { return nil },
	}
	err := buildWithOperations(&flags.Build{Server: true, Targets: "linux/amd64,windows/amd64", Tags: "extra"}, nil, operations)
	require.NoError(t, err)
	assert.True(t, got.Server)
	assert.Equal(t, "build", got.Verb)
	assert.Len(t, got.Targets, 2)
	assert.Equal(t, []string{"extra"}, got.Tags)
}
