package commands

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

func TestDefaultPuppertinoFrontendUsesManifestDirectory(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, manifest.WriteMinimal(root, manifest.Project{Name: "pup", ProductName: "Pup", Identifier: "com.example.pup", Version: "1.0.0"}))
	path := filepath.Join(root, manifest.Filename)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(raw), "directory: frontend")
	require.NoError(t, os.WriteFile(path, []byte(strings.Replace(string(raw), "directory: frontend", "directory: web", 1)), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "web"), 0o755))
	t.Chdir(root)

	dir, warning := defaultPuppertinoFrontend()
	require.NoError(t, warning)
	want, err := os.Stat(filepath.Join(root, "web"))
	require.NoError(t, err)
	got, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, os.SameFile(want, got), "expected %s to be the manifest frontend directory, got %s", filepath.Join(root, "web"), dir)
}

func TestDefaultPuppertinoFrontendFallsBackOnInvalidManifest(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, manifest.Filename), []byte("version: [unterminated\n"), 0o644))
	t.Chdir(root)

	dir, warning := defaultPuppertinoFrontend()
	assert.Equal(t, "frontend", dir)
	assert.ErrorContains(t, warning, manifest.Filename)
}

func TestToolPuppertinoWithInvalidManifestWarnsAndUsesFrontend(t *testing.T) {
	withPuppertinoServer(t, http.StatusOK)
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, manifest.Filename), []byte("version: [unterminated\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "frontend"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "frontend", "index.html"), []byte(puppertinoIndexFixture), 0o644))
	t.Chdir(root)

	require.NoError(t, ToolPuppertino(&ToolPuppertinoOptions{}))
	_, err := os.Stat(filepath.Join(root, "frontend", "public", "puppertino", "puppertino.css"))
	assert.NoError(t, err)
}
