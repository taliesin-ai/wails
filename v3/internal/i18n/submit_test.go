package i18n

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeGitStatusStub(t *testing.T, porcelain string) string {
	t.Helper()
	binDir := t.TempDir()
	script := filepath.Join(binDir, "git")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\nif [ \"$1\" = \"-C\" ]; then\n  shift 2\nfi\nif [ \"$1\" = \"status\" ] && [ \"$2\" = \"--porcelain\" ]; then\n  printf '%s\\n' \""+porcelain+"\"\nfi\n"), 0o755))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return binDir
}

func TestChangedTranslationFilesIncludesSnapshotPathWhenGitStatusReportsIt(t *testing.T) {
	// Given
	app, _ := makeAPIApp(t)
	makeGitStatusStub(t, "?? docs/.translation-cache/.snapshots/guide/record.mdx")

	// When
	files := app.changedTranslationFiles("/ignored", "de")

	// Then
	require.Contains(t, files, "docs/.translation-cache/.snapshots/guide/record.mdx")
}

func TestHandleSubmitExportIncludesSnapshotDirInManualGitAddCommand(t *testing.T) {
	// Given
	app, _ := makeAPIApp(t)
	body := bytes.NewBufferString(`{"locale":"de","mode":"export"}`)

	// When
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	app.handleSubmit(rr, req)

	// Then
	require.Equal(t, http.StatusOK, rr.Code)
	var resp submitResp
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Commands, 5)
	require.Contains(t, resp.Commands[1], "docs/.translation-cache/.snapshots")
}

func TestSubmitAddPathsIncludeSnapshotDirForAutomatedPR(t *testing.T) {
	// Given
	paths := submitAddPaths("de")

	// When / Then
	require.Equal(t, []string{
		"docs/src/content/docs/de",
		"docs/.translation-cache/de.json",
		"docs/.translation-cache/.snapshots",
	}, paths)
}

func TestHandleSubmitPRStopsBeforeGitWhenDocsBuildFails(t *testing.T) {
	// Given
	app, docs := makeAPIApp(t)
	require.NoError(t, os.Mkdir(filepath.Join(filepath.Dir(docs), ".git"), 0o755))
	logPath := makeSubmitCommandStubs(t)
	t.Setenv("SUBMIT_TEST_NPM_FAIL", "1")
	body := bytes.NewBufferString(`{"locale":"de","mode":"pr"}`)

	// When
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	app.handleSubmit(rr, req)

	// Then
	require.Equal(t, http.StatusOK, rr.Code)
	var resp submitResp
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Contains(t, resp.Error, "npm run build failed")
	log := readCommandLog(t, logPath)
	require.Contains(t, log, "npm run build")
	require.NotContains(t, log, "git commit")
	require.NotContains(t, log, "git push")
	require.NotContains(t, log, "gh pr create")
}

func TestHandleSubmitPRRunsDocsBuildBeforeOpeningPR(t *testing.T) {
	// Given
	app, docs := makeAPIApp(t)
	require.NoError(t, os.Mkdir(filepath.Join(filepath.Dir(docs), ".git"), 0o755))
	logPath := makeSubmitCommandStubs(t)
	body := bytes.NewBufferString(`{"locale":"de","mode":"pr"}`)

	// When
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	app.handleSubmit(rr, req)

	// Then
	require.Equal(t, http.StatusOK, rr.Code)
	var resp submitResp
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "https://github.com/wailsapp/wails/pull/1", resp.PRURL)
	log := readCommandLog(t, logPath)
	npmIndex := strings.Index(log, "npm run build")
	prIndex := strings.Index(log, "gh pr create")
	require.NotEqual(t, -1, npmIndex, log)
	require.NotEqual(t, -1, prIndex, log)
	require.Less(t, npmIndex, prIndex, log)
}

func makeSubmitCommandStubs(t *testing.T) string {
	t.Helper()
	binDir := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "commands.log")
	t.Setenv("SUBMIT_TEST_LOG", logPath)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	writeExecutable(t, filepath.Join(binDir, "git"), `#!/bin/sh
if [ "$1" = "-C" ]; then
  shift 2
fi
if [ "$1" = "status" ] && [ "$2" = "--porcelain" ]; then
  printf ' M docs/src/content/docs/de/guide/a.mdx\n'
  exit 0
fi
printf 'git %s\n' "$*" >> "$SUBMIT_TEST_LOG"
`)
	writeExecutable(t, filepath.Join(binDir, "gh"), `#!/bin/sh
printf 'gh %s\n' "$*" >> "$SUBMIT_TEST_LOG"
printf 'https://github.com/wailsapp/wails/pull/1\n'
`)
	writeExecutable(t, filepath.Join(binDir, "npm"), `#!/bin/sh
printf 'npm %s cwd=%s\n' "$*" "$(pwd)" >> "$SUBMIT_TEST_LOG"
if [ "$SUBMIT_TEST_NPM_FAIL" = "1" ]; then
  printf 'build failed\n'
  exit 1
fi
`)
	return logPath
}

func writeExecutable(t *testing.T, path string, body string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(body), 0o755))
}

func readCommandLog(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}
