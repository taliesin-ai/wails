package pipeline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func kindsOf(plan Plan) map[NodeKind]bool {
	kinds := map[NodeKind]bool{}
	for _, node := range plan.Nodes {
		kinds[node.Kind] = true
	}
	return kinds
}

func TestServerBuildPlanIsAPlainServerBinary(t *testing.T) {
	config := testConfig(t)
	config.Build.OutputDirectory = "bin"

	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", Server: true})
	require.NoError(t, err)

	kinds := kindsOf(plan)
	for _, kind := range []NodeKind{GeneratePlatformAssets, AssembleApplication, PackageArtifact, SignArtifact, MergeUniversalBinaries} {
		assert.False(t, kinds[kind], "server plans must not contain %s", kind)
	}
	assert.True(t, kinds[GenerateBindings], "bindings still run for embedded assets")
	assert.True(t, kinds[BuildFrontend], "the frontend still builds for embedded assets")

	compile := plan.Nodes["target:linux/amd64:compile"]
	spec := compile.Spec.(CompileSpec)
	assert.Equal(t, ".wails/build/default/linux-amd64/artifacts/app-server", compile.Output)
	assert.True(t, spec.Server)
	assert.Empty(t, spec.Assets, "server builds generate no platform assets")
	assert.True(t, spec.Production)
	assert.True(t, spec.TrimPath)
	assert.True(t, spec.Strip)
	assert.False(t, spec.VCSInfo)
	assert.Equal(t, "native", spec.Toolchain)
	assert.Equal(t, "0", spec.Environment["CGO_ENABLED"], "server compiles default to CGO off")
	assert.Contains(t, spec.Tags, "server")
	assert.Contains(t, spec.Tags, "production")

	assert.Equal(t, []string{"collect:artifacts"}, planRoots(plan))
	publish := publishNodeFor(t, plan)
	assert.Equal(t, "bin/app-server", publish.Output)
	assert.Equal(t, "bin/app-server", publish.Spec.(PublishSpec).Destination)
	assert.Equal(t, ArtifactBinary, publish.Artifact.Kind)
}

// publishNodeFor returns the single publish node of a plan.
func publishNodeFor(t *testing.T, plan Plan) Node {
	t.Helper()
	var found []Node
	for _, node := range plan.Nodes {
		if node.Kind == PublishArtifact {
			found = append(found, node)
		}
	}
	require.Len(t, found, 1)
	return found[0]
}

func planRoots(plan Plan) []string {
	roots := make([]string, 0, len(plan.Roots))
	for _, key := range plan.Roots {
		roots = append(roots, string(key))
	}
	return roots
}

func TestServerBuildPlanNamesWindowsBinaryWithExtension(t *testing.T) {
	config := testConfig(t)
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "windows", TargetArch: "amd64", Server: true})
	require.NoError(t, err)
	kinds := kindsOf(plan)
	assert.False(t, kinds[GeneratePlatformAssets], "windows GUI builds generate syso/resource assets; server builds must not")
	compile := plan.Nodes["target:windows/amd64:compile"].Spec.(CompileSpec)
	assert.Equal(t, ".wails/build/default/windows-amd64/artifacts/app-server.exe", compile.Output)
	assert.Empty(t, compile.Assets)
	assert.Equal(t, "bin/app-server.exe", publishNodeFor(t, plan).Output)
}

func TestServerBuildPlanKeepsDarwinAsPlainBinary(t *testing.T) {
	config := testConfig(t)
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "darwin", TargetArch: "arm64", Server: true})
	require.NoError(t, err)
	kinds := kindsOf(plan)
	assert.False(t, kinds[AssembleApplication], "server builds never assemble a .app bundle")
	assert.False(t, kinds[GeneratePlatformAssets])
	publish := publishNodeFor(t, plan)
	assert.Equal(t, "bin/app-server", publish.Output)
	compile := plan.Nodes["target:darwin/arm64:compile"].Spec.(CompileSpec)
	assert.Equal(t, "0", compile.Environment["CGO_ENABLED"], "the Darwin GUI CGO override must not reach server compiles")
}

func TestServerBuildPlanCompileEnvironmentIsOverridable(t *testing.T) {
	config := testConfig(t)
	config.Build.Environment = map[string]string{"CGO_ENABLED": "1", "SHARED": "build"}
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", Server: true})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"CGO_ENABLED": "1", "SHARED": "build"}, plan.Nodes["target:linux/amd64:compile"].Spec.(CompileSpec).Environment)
	assert.Equal(t, map[string]string{"CGO_ENABLED": "0"}, serverEnvironment(nil, nil))
	assert.Equal(t, map[string]string{"CGO_ENABLED": "0", "SHARED": "build"}, serverEnvironment(map[string]string{"SHARED": "build"}, nil))
	assert.Equal(t, map[string]string{"CGO_ENABLED": "0", "SHARED": "target"}, serverEnvironment(map[string]string{"SHARED": "build"}, map[string]string{"SHARED": "target"}))
}

func TestServerBuildPlanDropsWindowsGUISubsystemLinkFlag(t *testing.T) {
	config := testConfig(t)
	config.Build.Go.LinkerFlags = []string{"-X", "main.version=1", "-H", "windowsgui"}
	config.Targets.Windows.AMD64.LinkerFlags = []string{"-H=windowsgui", "-X main.build=2 -H windowsgui"}
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "windows", TargetArch: "amd64", Server: true})
	require.NoError(t, err)
	compile := plan.Nodes["target:windows/amd64:compile"].Spec.(CompileSpec)
	assert.Equal(t, []string{"-X main.build=2"}, compile.LinkerFlags)

	gui, err := PlanBuild(config, Request{Verb: "build", TargetOS: "windows", TargetArch: "amd64"})
	require.NoError(t, err)
	assert.Contains(t, gui.Nodes["target:windows/amd64:compile"].Spec.(CompileSpec).LinkerFlags, "-H=windowsgui", "GUI builds keep the manifest's linker flags unchanged")

	assert.Equal(t, []string{"-X", "main.version=1"}, withoutWindowsGUILinkFlag([]string{"-X", "main.version=1", "-H", "windowsgui"}))
	assert.Equal(t, []string{"-w", "-s"}, withoutWindowsGUILinkFlag([]string{"-w", "-s"}))
}

func TestServerBuildPlanIgnoresZigAndDockerToolchainSettings(t *testing.T) {
	config := testConfig(t)
	config.Targets.Linux.AMD64.Toolchain = "zig"
	config.Targets.Windows.AMD64.Toolchain = "docker"
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", Server: true})
	require.NoError(t, err)
	assert.Equal(t, "native", plan.Nodes["target:linux/amd64:compile"].Spec.(CompileSpec).Toolchain)
	plan, err = PlanBuild(config, Request{Verb: "build", TargetOS: "windows", TargetArch: "amd64", Server: true})
	require.NoError(t, err)
	assert.Equal(t, "native", plan.Nodes["target:windows/amd64:compile"].Spec.(CompileSpec).Toolchain)
}

func TestServerCompilesNeedNoCrossToolchain(t *testing.T) {
	// A darwin host with only Go and npm: no zig, Docker, Podman or C compiler.
	host := NewHostCapabilities("darwin", "arm64", []string{"go", "npm", "node"}, nil)
	config := testConfig(t)

	_, err := PlanBuildForHost(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64"}, host)
	assert.Error(t, err, "the GUI linux build needs a container or zig toolchain on this host")

	for _, target := range []string{"linux/amd64", "linux/arm64", "windows/amd64", "darwin/amd64"} {
		targetOS, targetArch, _ := strings.Cut(target, "/")
		plan, err := PlanBuildForHost(config, Request{Verb: "build", TargetOS: targetOS, TargetArch: targetArch, Server: true}, host)
		require.NoError(t, err, target)
		spec := plan.Nodes[NodeKey("target:"+target+":compile")].Spec.(CompileSpec)
		assert.Equal(t, "native", spec.Toolchain, target)
	}
}

func TestServerDevelopmentPlanUsesTaskfileDevFlags(t *testing.T) {
	config := testConfig(t)
	config.Dev.Tags = []string{"devtools"}
	config.Build.TrimPath = true
	config.Build.Strip = true
	plan, err := PlanBuild(config, Request{Verb: "run", TargetOS: "linux", TargetArch: "amd64", Development: true, Server: true})
	require.NoError(t, err)
	compile := plan.Nodes["target:linux/amd64:compile"]
	spec := compile.Spec.(CompileSpec)
	assert.Equal(t, ".wails/dev/linux-amd64/app-server", compile.Output)
	assert.Equal(t, []string{"server", "devtools"}, spec.Tags)
	assert.NotContains(t, spec.Tags, "production")
	assert.Equal(t, []string{"all=-l"}, spec.CompilerFlags)
	assert.False(t, spec.TrimPath)
	assert.False(t, spec.Strip)
	assert.False(t, spec.Production)
	assert.Equal(t, []string{"collect:artifacts"}, planRoots(plan))
	assert.False(t, kindsOf(plan)[PublishArtifact], "development server builds run in place and publish nothing")
}

func TestServerBuildPlanRejectsUnsupportedRequests(t *testing.T) {
	config := testConfig(t)
	_, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", Server: true, Formats: []string{"deb"}})
	assert.ErrorContains(t, err, "--formats does not apply")
	_, err = PlanBuild(config, Request{Verb: "sign", TargetOS: "linux", TargetArch: "amd64", Server: true})
	assert.ErrorContains(t, err, "build and run")
	_, err = PlanBuild(config, Request{Verb: "build", TargetOS: "ios", TargetArch: "arm64", Server: true})
	assert.ErrorContains(t, err, "desktop-only")
	_, err = PlanBuild(config, Request{Verb: "build", TargetOS: "android", TargetArch: "arm64", Server: true})
	assert.ErrorContains(t, err, "desktop-only")
	_, err = PlanBuild(config, Request{Verb: "build", TargetOS: "darwin", TargetArch: "universal", Server: true})
	assert.ErrorContains(t, err, "universal")
}

func TestServerAndGUIBuildsHaveDistinctCompileIdentities(t *testing.T) {
	config := testConfig(t)
	gui, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64"})
	require.NoError(t, err)
	server, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", Server: true})
	require.NoError(t, err)
	guiSpec := gui.Nodes["target:linux/amd64:compile"].Spec.(CompileSpec)
	serverSpec := server.Nodes["target:linux/amd64:compile"].Spec.(CompileSpec)
	assert.NotEqual(t, guiSpec.Output, serverSpec.Output)
	assert.NotEqual(t, guiSpec.Tags, serverSpec.Tags)
	assert.NotEqual(t, guiSpec, serverSpec, "the compile spec is hashed into the action key, so the two builds must not share cache entries")
}
