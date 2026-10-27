package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

// A customised build:native whose vars block was deleted has a nil vars map.
// Migration must block it as a modified task, not panic.
func TestBuildNativeWithoutVarsBlockIsBlockedNotPanicking(t *testing.T) {
	root := initLegacyStockProject(t, "react")
	edits := customCorpusEdits{Edits: []customCorpusEdit{{
		File:    "build/darwin/Taskfile.yml",
		Pattern: `(?s)    vars:\n      BUILD_FLAGS:.*?\n    env:\n`,
		New:     "    env:\n",
	}}}
	applyCustomCorpusEdits(t, root, edits)

	var report MigrationReport
	require.NotPanics(t, func() {
		var err error
		report, _, err = analyseMigration(root)
		require.NoError(t, err)
	})
	assert.False(t, report.Complete)
	assert.Contains(t, migrationDiagnosticCodes(report), "modified-task")
	assert.Contains(t, diagnosticCodesForTask(report, "build:native"), "modified-task")
}

// An existing draft is left unchanged, so hook scripts from a changed Taskfile
// must not be written beside it: they are reported as pending instead.
func TestExistingDraftDoesNotWriteHookScripts(t *testing.T) {
	// Resolve the fixture before the test changes the working directory.
	fixture, err := filepath.Abs(filepath.Join(customCorpusDirectory, "C07-pre-build-step", "edits.json"))
	require.NoError(t, err)
	root := initLegacyStockProject(t, "react")
	withMigrationWorkingDirectory(t, root)
	captureMigrationStdout(t, func() { require.NoError(t, Migrate(&MigrateOptions{})) })
	draft := filepath.Join(root, manifest.MigratedFilename)
	before, err := os.ReadFile(draft)
	require.NoError(t, err)

	var edits customCorpusEdits
	readCustomCorpusJSON(t, fixture, &edits)
	applyCustomCorpusEdits(t, root, edits)

	output := captureMigrationStdout(t, func() { require.NoError(t, Migrate(&MigrateOptions{})) })
	assert.Contains(t, output, "Left existing inactive")
	assert.Contains(t, output, "Would write build/hooks/before_build.sh")
	assert.NoFileExists(t, filepath.Join(root, "build", "hooks", "before_build.sh"))
	after, err := os.ReadFile(draft)
	require.NoError(t, err)
	assert.Equal(t, before, after, "the existing draft must not change")
}
