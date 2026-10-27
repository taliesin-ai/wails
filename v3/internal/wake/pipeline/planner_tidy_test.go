package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func labelsOf(node Node) []string {
	labels := make([]string, 0, len(node.Inputs))
	for _, input := range node.Inputs {
		labels = append(labels, input.Label)
	}
	return labels
}

func TestPlannerOmitsTidyWhenDisabled(t *testing.T) {
	config := testConfig(t)
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64"})
	require.NoError(t, err)
	_, ok := plan.Nodes["go:tidy"]
	assert.False(t, ok)
	assert.NotContains(t, plan.Nodes["frontend:bindings"].Dependencies, NodeKey("go:tidy"))
	assert.NotContains(t, plan.Nodes["target:linux/amd64:compile"].Dependencies, NodeKey("go:tidy"))
}

func TestPlannerTidyRunsBeforeBindingsAndCompilation(t *testing.T) {
	config := testConfig(t)
	config.Build.Tidy = true
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64"})
	require.NoError(t, err)

	tidy, ok := plan.Nodes["go:tidy"]
	require.True(t, ok)
	assert.Equal(t, TidyGoModules, tidy.Kind)
	assert.Equal(t, "Tidy Go modules", tidy.Label)
	assert.Equal(t, ProjectScope, tidy.Scope)
	assert.Equal(t, CacheReceipt, tidy.Cache)
	assert.Equal(t, "go.mod", tidy.Marker)
	assert.Equal(t, []string{"go-module", "go-tidy-sources"}, labelsOf(tidy)[:2])

	assert.Contains(t, plan.Nodes["frontend:bindings"].Dependencies, NodeKey("go:tidy"))
	assert.Contains(t, plan.Nodes["target:linux/amd64:compile"].Dependencies, NodeKey("go:tidy"))
	assert.IsType(t, TidySpec{}, tidy.Spec)
}

func TestPlannerTidyDevelopmentKeysOnModuleFilesOnly(t *testing.T) {
	config := testConfig(t)
	config.Build.Tidy = true
	plan, err := PlanBuild(config, Request{Verb: "build", TargetOS: "darwin", TargetArch: "arm64", Development: true})
	require.NoError(t, err)

	tidy, ok := plan.Nodes["go:tidy"]
	require.True(t, ok)
	assert.Equal(t, []string{"go-module"}, labelsOf(tidy))
	assert.Contains(t, plan.Nodes["target:darwin/arm64:compile"].Dependencies, NodeKey("go:tidy"))
}
