package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

func TestTidyNodeRunsGoModTidyInProjectRoot(t *testing.T) {
	root := t.TempDir()
	var calls []string
	var gotDirectory string
	previous := manifestGoModTidy
	manifestGoModTidy = func(_ context.Context, dir string, _ []string) (string, error) {
		calls = append(calls, "go mod tidy")
		gotDirectory = dir
		return "tidied", nil
	}
	t.Cleanup(func() { manifestGoModTidy = previous })

	handler := &manifestHandler{root: root}
	result, err := handler.Run(context.Background(), pipeline.Node{Key: "go:tidy", Kind: pipeline.TidyGoModules, Spec: pipeline.TidySpec{}})
	require.NoError(t, err)
	assert.Equal(t, []string{"go mod tidy"}, calls)
	assert.Equal(t, root, gotDirectory)
	assert.Equal(t, "tidied", result.Detail)
}

func TestTidyNodePropagatesFailure(t *testing.T) {
	previous := manifestGoModTidy
	manifestGoModTidy = func(context.Context, string, []string) (string, error) {
		return "boom", errors.New("exit status 1")
	}
	t.Cleanup(func() { manifestGoModTidy = previous })

	handler := &manifestHandler{root: t.TempDir()}
	result, err := handler.Run(context.Background(), pipeline.Node{Key: "go:tidy", Kind: pipeline.TidyGoModules, Spec: pipeline.TidySpec{}})
	require.Error(t, err)
	assert.Equal(t, "boom", result.Detail)
}
