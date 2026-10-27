package commands

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

const (
	serverDockerfileName     = "Dockerfile.server"
	serverDockerfileLegacy   = "build/docker/Dockerfile.server"
	serverDockerfileGenerate = ".wails/docker/Dockerfile.server"
	serverContainerPort      = 8080
	serverFrontendCopyLine   = "COPY --from=builder /app/frontend/dist /frontend/dist"
	serverFrontendCopySource = "/app/frontend/dist"
)

// ServerDockerOptions configures `wails3 server docker`. Defaults come from
// NewServerDockerOptions and mirror the build:docker Taskfile task.
type ServerDockerOptions struct {
	Tag          string `name:"tag" description:"Image tag (default <binary_name>:latest)"`
	CGO          bool   `name:"cgo" description:"Build with CGO_ENABLED=1; requires a libc-compatible -go-image and -runtime-image"`
	GoImage      string `name:"go-image" description:"Builder image for the Go build stage"`
	RuntimeImage string `name:"runtime-image" description:"Runtime image for the final stage"`
	Run          bool   `name:"run" description:"Run the image after building it"`
	Port         int    `name:"port" description:"Host port mapped to the container's port 8080 when -run is set"`
}

// NewServerDockerOptions returns the default options for `wails3 server docker`.
func NewServerDockerOptions() ServerDockerOptions {
	return ServerDockerOptions{
		GoImage:      "golang:alpine",
		RuntimeImage: "gcr.io/distroless/static-debian12",
		Port:         serverContainerPort,
	}
}

// serverDockerOperations isolates every side effect so tests can run the full
// sequence without Docker or a Go toolchain.
type serverDockerOperations struct {
	// probe runs a silent Docker command and reports whether it succeeded.
	probe func(root string, args ...string) (string, error)
	// docker runs a Docker command with its output attached to the terminal.
	docker func(ctx context.Context, root string, args ...string) error
	// buildFrontend runs the production frontend and bindings for the server
	// build; the compiled host binary it also produces is not used.
	buildFrontend func(ctx context.Context, loaded *manifest.Loaded) error
}

// ServerDocker builds the frontend for server mode and packages the project
// into a Docker image using the Dockerfile.server template.
func ServerDocker(options *ServerDockerOptions, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("wails3 server docker takes no arguments: %s", strings.Join(args, " "))
	}
	if options.Port <= 0 || options.Port > 65535 {
		return fmt.Errorf("-port must be between 1 and 65535, got %d", options.Port)
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, _, err := manifest.Discover(root); err != nil {
		return fmt.Errorf("server docker requires a wails.yaml project: %w", err)
	}
	loaded, err := manifest.Load(root, "")
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return serverDockerWithOperations(ctx, options, loaded, serverDockerOperations{
		probe: func(root string, args ...string) (string, error) {
			return runManifestCommand(ctx, root, nil, "docker", args...)
		},
		docker: func(ctx context.Context, root string, args ...string) error {
			return runApplicationProcess(ctx, root, "docker", args, nil)
		},
		buildFrontend: func(ctx context.Context, loaded *manifest.Loaded) error {
			options := serverDockerFrontendRequest(loaded)
			options.Context = ctx
			_, err := runManifestPipelineResult(options)
			return err
		},
	})
}

func serverDockerWithOperations(ctx context.Context, options *ServerDockerOptions, loaded *manifest.Loaded, operations serverDockerOperations) error {
	root := loaded.Config.Root
	// Fail before any build work when Docker is unreachable.
	if _, err := operations.probe(root, "info"); err != nil {
		return fmt.Errorf("docker is required for wails3 server docker, but `docker info` failed (is the Docker daemon running?): %w", err)
	}
	dockerfile, err := resolveServerDockerfile(root, loaded.Config.Frontend)
	if err != nil {
		return err
	}
	tag := options.Tag
	if tag == "" {
		tag = loaded.Config.Project.BinaryName + ":latest"
	}
	if err := operations.buildFrontend(ctx, loaded); err != nil {
		return err
	}
	if err := operations.docker(ctx, root, serverDockerBuildArgs(tag, dockerfile, options)...); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}
	if !options.Run {
		return nil
	}
	return operations.docker(ctx, root, serverDockerRunArgs(tag, options.Port)...)
}

// serverDockerFrontendRequest is the pipeline request that produces the
// frontend and bindings embedded by the image. The image compiles linux/amd64
// inside its builder stage, so the host-side request targets that platform too:
// server builds never need a host-native or cross toolchain, so this works on
// any host. The host binary the request also publishes is not used by the image.
func serverDockerFrontendRequest(loaded *manifest.Loaded) manifestRunOptions {
	return manifestRunOptions{Verb: "build", Loaded: loaded, TargetOS: "linux", TargetArch: "amd64", Server: true}
}

// serverDockerBuildArgs returns the argv after "docker" for the image build.
// The build context is the project root, as in the build:docker Taskfile task.
func serverDockerBuildArgs(tag, dockerfile string, options *ServerDockerOptions) []string {
	cgo := "0"
	if options.CGO {
		cgo = "1"
	}
	return []string{
		"build",
		"--build-arg", "CGO_ENABLED=" + cgo,
		"--build-arg", "GO_IMAGE=" + options.GoImage,
		"--build-arg", "RUNTIME_IMAGE=" + options.RuntimeImage,
		"-t", tag,
		"-f", dockerfile,
		".",
	}
}

// serverDockerRunArgs returns the argv after "docker" for `docker run`.
func serverDockerRunArgs(tag string, port int) []string {
	return []string{"run", "--rm", "-p", strconv.Itoa(port) + ":" + strconv.Itoa(serverContainerPort), tag}
}

// resolveServerDockerfile chooses the Dockerfile in precedence order: a
// user-owned build/docker/Dockerfile.server (legacy layout), then the embedded
// template written to .wails/docker/Dockerfile.server. The returned path is
// relative to root, which is also the build context.
func resolveServerDockerfile(root string, frontend manifest.Frontend) (string, error) {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(serverDockerfileLegacy)))
	if err == nil && info.Mode().IsRegular() {
		return serverDockerfileLegacy, nil
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	template, err := buildAssets.ReadFile("build_assets/docker/" + serverDockerfileName)
	if err != nil {
		return "", fmt.Errorf("read embedded %s: %w", serverDockerfileName, err)
	}
	content := renderServerDockerfile(string(template), frontend)
	destination := filepath.Join(root, filepath.FromSlash(serverDockerfileGenerate))
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(destination, []byte(content), 0o644); err != nil {
		return "", err
	}
	return serverDockerfileGenerate, nil
}

// renderServerDockerfile adapts the template to a wake project layout. The
// template copies the frontend output from frontend/dist; wake projects may
// place it elsewhere, and a headless project has no output at all.
func renderServerDockerfile(template string, frontend manifest.Frontend) string {
	if frontend.Disabled {
		lines := strings.Split(template, "\n")
		kept := lines[:0]
		for _, line := range lines {
			if strings.TrimSpace(line) == serverFrontendCopyLine {
				continue
			}
			kept = append(kept, line)
		}
		return strings.Join(kept, "\n")
	}
	output := frontendOutputPathForDocker(frontend.Directory, frontend.OutputDirectory)
	if output == "frontend/dist" {
		return template
	}
	return strings.Replace(template, serverFrontendCopySource, "/app/"+output, 1)
}

// frontendOutputPathForDocker mirrors the planner's frontend output resolution,
// so the Dockerfile copies the directory the frontend build actually writes.
func frontendOutputPathForDocker(directory, output string) string {
	directory = filepath.ToSlash(filepath.Clean(directory))
	output = filepath.ToSlash(filepath.Clean(output))
	if output == directory || strings.HasPrefix(output, directory+"/") {
		return output
	}
	return filepath.ToSlash(filepath.Join(directory, output))
}
