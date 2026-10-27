package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

// TestGeneratedHookScriptsAreNewFilesAndActivationNeedsThem checks the write
// rules: a dry run creates nothing, a migration creates an executable script,
// and activation fails when the script the draft references is missing.
func TestGeneratedHookScriptsAreNewFilesAndActivationNeedsThem(t *testing.T) {
	var edits customCorpusEdits
	readCustomCorpusJSON(t, filepath.Join(customCorpusDirectory, "C07-pre-build-step", "edits.json"), &edits)
	root := initLegacyStockProject(t, "react")
	applyCustomCorpusEdits(t, root, edits)
	script := filepath.Join(root, "build", "hooks", "before_build.sh")
	withMigrationWorkingDirectory(t, root)

	dry := captureMigrationStdout(t, func() { require.NoError(t, Migrate(&MigrateOptions{DryRun: true})) })
	assert.Contains(t, dry, "Would write build/hooks/before_build.sh")
	assert.NoFileExists(t, script, "a dry run must not create hook scripts")
	assert.NoFileExists(t, filepath.Join(root, manifest.MigratedFilename))

	written := captureMigrationStdout(t, func() { require.NoError(t, Migrate(&MigrateOptions{})) })
	assert.Contains(t, written, "Wrote build/hooks/before_build.sh")
	info, err := os.Stat(script)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
	}

	captureMigrationStdout(t, func() { require.NoError(t, Migrate(&MigrateOptions{Activate: true})) })
	assert.FileExists(t, filepath.Join(root, manifest.Filename))
}

func TestActivationFailsWhenTheDraftsHookScriptIsMissing(t *testing.T) {
	var edits customCorpusEdits
	readCustomCorpusJSON(t, filepath.Join(customCorpusDirectory, "C07-pre-build-step", "edits.json"), &edits)
	root := initLegacyStockProject(t, "react")
	applyCustomCorpusEdits(t, root, edits)
	withMigrationWorkingDirectory(t, root)

	captureMigrationStdout(t, func() { require.NoError(t, Migrate(&MigrateOptions{})) })
	require.NoError(t, os.Remove(filepath.Join(root, "build", "hooks", "before_build.sh")))

	var activateErr error
	captureMigrationStdout(t, func() { activateErr = Migrate(&MigrateOptions{Activate: true}) })
	require.Error(t, activateErr, "activation must not succeed without the referenced hook script")
	assert.NoFileExists(t, filepath.Join(root, manifest.Filename))
}
