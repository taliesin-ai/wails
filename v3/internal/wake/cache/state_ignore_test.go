package cache

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureStateDirectoryCreatesSelfIgnoringState(t *testing.T) {
	state := filepath.Join(t.TempDir(), ".wails")
	require.NoError(t, EnsureStateDirectory(state))

	data, err := os.ReadFile(filepath.Join(state, ".gitignore"))
	require.NoError(t, err)
	assert.Equal(t, "*\n", string(data))
}

func TestEnsureStateDirectoryLeavesExistingIgnoreFileAlone(t *testing.T) {
	state := filepath.Join(t.TempDir(), ".wails")
	require.NoError(t, os.MkdirAll(state, 0o755))
	custom := []byte("keep-this\n")
	require.NoError(t, os.WriteFile(filepath.Join(state, ".gitignore"), custom, 0o644))

	require.NoError(t, EnsureStateDirectory(state))
	data, err := os.ReadFile(filepath.Join(state, ".gitignore"))
	require.NoError(t, err)
	assert.Equal(t, custom, data)
}

func TestOpenCacheCreatesSelfIgnoringStateButReadOnlyOpenDoesNot(t *testing.T) {
	root := t.TempDir()
	_, err := OpenCacheReadOnly(root)
	require.NoError(t, err)
	assert.NoDirExists(t, filepath.Join(root, ".wails"), "read-only opens must not create project state")

	_, err = OpenCache(root)
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(root, ".wails", ".gitignore"))
	require.NoError(t, err)
	assert.Equal(t, "*\n", string(data))
}

// TestGeneratedStateDoesNotDirtyARepository runs the real git tool against a
// repository whose .gitignore no longer mentions .wails, as older projects have.
func TestGeneratedStateDoesNotDirtyARepository(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	root := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		command := exec.Command(git, args...)
		command.Dir = root
		output, err := command.CombinedOutput()
		require.NoError(t, err, string(output))
		return string(output)
	}
	run("init", "-q")
	require.NoError(t, os.WriteFile(filepath.Join(root, ".gitignore"), []byte("bin/\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "wails.yaml"), []byte("version: 3\n"), 0o644))

	_, err = OpenCache(root)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, ".wails", "generated.txt"), []byte("state"), 0o644))

	status := run("status", "--porcelain", "--untracked-files=all")
	assert.NotContains(t, status, ".wails")
	assert.Contains(t, status, "wails.yaml")
}
