package pipeline

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nodeOfKind(t *testing.T, plan Plan, kind NodeKind) Node {
	t.Helper()
	for _, node := range plan.Nodes {
		if node.Kind == kind {
			return node
		}
	}
	require.Failf(t, "missing node", "no %s node in plan", kind)
	return Node{}
}

func hasNodeOfKind(plan Plan, kind NodeKind) bool {
	for _, node := range plan.Nodes {
		if node.Kind == kind {
			return true
		}
	}
	return false
}

func TestDevBuildPublishesDevelopmentBinaryToOutputDirectory(t *testing.T) {
	config := testConfig(t)
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", DevBuild: true})
	require.NoError(t, err)

	publish := nodeOfKind(t, plan, PublishArtifact)
	assert.Equal(t, filepath.ToSlash(filepath.Join(config.Build.OutputDirectory, config.Project.BinaryName)), publish.Output)
	assert.Contains(t, plan.Artifacts, publish.Key)

	compile := nodeOfKind(t, plan, CompileApplication)
	spec := compile.Spec.(CompileSpec)
	assert.False(t, spec.Production)
	assert.NotContains(t, spec.Tags, "production")
	assert.Equal(t, []string{"all=-l"}, spec.CompilerFlags)
	assert.False(t, spec.TrimPath)
	assert.False(t, spec.Strip)
	assert.False(t, spec.VCSInfo)
	assert.True(t, strings.HasPrefix(compile.Output, ".wails/dev/"), compile.Output)
}

func TestDevBuildAndProductionIdentitiesDoNotCollide(t *testing.T) {
	config := testConfig(t)
	devPlan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", DevBuild: true})
	require.NoError(t, err)
	prodPlan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64"})
	require.NoError(t, err)

	devCompile := nodeOfKind(t, devPlan, CompileApplication)
	prodCompile := nodeOfKind(t, prodPlan, CompileApplication)
	assert.NotEqual(t, devCompile.Spec, prodCompile.Spec)
	assert.NotEqual(t, devCompile.Output, prodCompile.Output)
	assert.True(t, strings.HasPrefix(prodCompile.Output, ".wails/build/"), prodCompile.Output)

	// Both publish to the same user-visible path, so the later build replaces the
	// earlier one: the publish action keys differ because their sources differ.
	devPublish := nodeOfKind(t, devPlan, PublishArtifact)
	prodPublish := nodeOfKind(t, prodPlan, PublishArtifact)
	assert.Equal(t, devPublish.Output, prodPublish.Output)
	assert.NotEqual(t, devPublish.Spec, prodPublish.Spec)
}

func TestDevBuildSupportsObfuscationOnExplicitRequestOnly(t *testing.T) {
	config := testConfig(t)
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", DevBuild: true, Obfuscated: true})
	require.NoError(t, err)
	spec := nodeOfKind(t, plan, CompileApplication).Spec.(CompileSpec)
	assert.True(t, spec.Obfuscated)
	assert.Contains(t, spec.Tags, "wails_obfuscated")
	assert.NotContains(t, spec.Tags, "production")

	plan, err = PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", DevBuild: true})
	require.NoError(t, err)
	spec = nodeOfKind(t, plan, CompileApplication).Spec.(CompileSpec)
	assert.False(t, spec.Obfuscated)
}

func TestDevBuildAddsAdhocSignedDevelopmentBundleOnDarwin(t *testing.T) {
	config := testConfig(t)
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "darwin", TargetArch: "arm64", DevBuild: true})
	require.NoError(t, err)

	assemble := nodeOfKind(t, plan, AssembleApplication)
	spec := assemble.Spec.(PackageSpec)
	assert.True(t, spec.Development)
	assert.True(t, spec.AdhocSign)
	assert.True(t, strings.HasPrefix(assemble.Output, ".wails/dev/"), assemble.Output)

	publish := nodeOfKind(t, plan, PublishArtifact)
	assert.True(t, strings.HasPrefix(publish.Output, config.Build.OutputDirectory+"/"), publish.Output)
	assert.True(t, strings.HasSuffix(publish.Output, ".app"), publish.Output)
}

func TestDevBuildUsesDevelopmentFrontendAndFullTidyInputs(t *testing.T) {
	config := testConfig(t)
	config.Build.Tidy = true
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", DevBuild: true})
	require.NoError(t, err)

	frontend := plan.Nodes["frontend:build"].Spec.(FrontendSpec)
	assert.False(t, frontend.Production)
	assert.True(t, strings.HasSuffix(frontend.Command, ":dev"), frontend.Command)
	assert.Equal(t, "frontend-source", plan.Nodes["frontend:build"].Inputs[0].Label)

	tidy := plan.Nodes["go:tidy"]
	assert.Equal(t, "go-tidy-sources", tidy.Inputs[1].Label)
}

func TestDevelopmentSessionStillDoesNotPublish(t *testing.T) {
	config := testConfig(t)
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64", Development: true})
	require.NoError(t, err)
	assert.False(t, hasNodeOfKind(plan, PublishArtifact))
	assert.True(t, strings.HasPrefix(nodeOfKind(t, plan, CompileApplication).Output, ".wails/dev/"))
}
