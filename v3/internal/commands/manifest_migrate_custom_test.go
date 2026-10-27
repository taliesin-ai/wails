package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

// customCorpusEdit changes one file of a stock project. Old must occur exactly
// once unless All is set; Pattern is a regular expression used instead of Old.
type customCorpusEdit struct {
	File    string `json:"file"`
	Old     string `json:"old"`
	Pattern string `json:"pattern"`
	New     string `json:"new"`
	All     bool   `json:"all"`
}

type customCorpusEdits struct {
	Edits  []customCorpusEdit `json:"edits"`
	Remove []string           `json:"remove"`
}

type customCorpusExpect struct {
	Complete        bool           `json:"complete"`
	Codes           []string       `json:"codes"`
	AbsentCodes     []string       `json:"absent_codes"`
	MessageContains []string       `json:"message_contains"`
	Manifest        map[string]any `json:"manifest"`
	Activate        bool           `json:"activate"`
	PlanNodes       []string       `json:"plan_nodes"`
	// Verb selects the plan checked for PlanNodes; empty means build.
	Verb string `json:"verb"`
	// ScriptContains maps a generated script to fragments it must contain. A
	// key ending in ":absent" lists fragments it must not contain.
	ScriptContains map[string][]string `json:"script_contains"`
	// KnownGap names a case whose translation is not implemented yet. Such a
	// case is reported as skipped instead of failing the corpus.
	KnownGap string `json:"known_gap"`
}

const customCorpusDirectory = "testdata/migration/custom"

func TestCustomTaskfileCorpus(t *testing.T) {
	entries, err := os.ReadDir(customCorpusDirectory)
	require.NoError(t, err)
	passed, total := 0, 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		total++
		name := entry.Name()
		if runCustomCorpusCase(t, filepath.Join(customCorpusDirectory, name), name) {
			passed++
		}
	}
	require.Equal(t, 16, total, "corpus must contain every planned case")
	t.Logf("custom corpus: %d/%d complete as expected", passed, total)
}

// runCustomCorpusCase reports whether the case ran and met its expectations.
// Known gaps are skipped and do not count as complete.
func runCustomCorpusCase(t *testing.T, directory, name string) bool {
	t.Helper()
	gap := false
	passed := t.Run(name, func(t *testing.T) {
		var expect customCorpusExpect
		readCustomCorpusJSON(t, filepath.Join(directory, "expect.json"), &expect)
		if expect.KnownGap != "" {
			gap = true
			t.Skipf("known gap: %s", expect.KnownGap)
		}
		var edits customCorpusEdits
		readCustomCorpusJSON(t, filepath.Join(directory, "edits.json"), &edits)

		root := initLegacyStockProject(t, "react")
		applyCustomCorpusEdits(t, root, edits)

		report, doc, err := analyseMigration(root)
		require.NoError(t, err)
		codes := migrationDiagnosticCodes(report)
		if report.Complete != expect.Complete {
			for _, diagnostic := range report.Diagnostics {
				if diagnostic.Severity != "info" {
					t.Logf("%s %s [%s]: %s", diagnostic.Severity, diagnostic.File, diagnostic.Task, diagnostic.Message)
				}
			}
		}
		assert.Equal(t, expect.Complete, report.Complete, "complete flag")
		for _, code := range expect.Codes {
			assert.Contains(t, codes, code)
		}
		for _, code := range expect.AbsentCodes {
			assert.NotContains(t, codes, code)
		}
		messages := make([]string, 0, len(report.Diagnostics))
		for _, diagnostic := range report.Diagnostics {
			location := diagnostic.File
			if diagnostic.Task != "" {
				location += " [" + diagnostic.Task + "]"
			}
			messages = append(messages, location+": "+diagnostic.Message)
		}
		for _, text := range expect.MessageContains {
			assert.True(t, containsAny(messages, text), "no diagnostic message contains %q", text)
		}
		assertCustomCorpusManifest(t, doc, expect.Manifest)

		if !expect.Complete {
			withMigrationWorkingDirectory(t, root)
			captureMigrationStdout(t, func() { require.NoError(t, Migrate(&MigrateOptions{})) })
			assert.FileExists(t, filepath.Join(root, manifest.MigratedFilename), "blocked migrations still write the draft")
		}
		if expect.Activate {
			withMigrationWorkingDirectory(t, root)
			captureMigrationStdout(t, func() {
				require.NoError(t, Migrate(&MigrateOptions{}))
				require.NoError(t, Migrate(&MigrateOptions{Activate: true}))
			})
			for file, fragments := range expect.ScriptContains {
				checkCustomCorpusScript(t, root, file, fragments)
			}
			loaded, err := manifest.Load(root, "")
			require.NoError(t, err)
			verb := expect.Verb
			if verb == "" {
				verb = "build"
			}
			plan, err := pipeline.PlanBuild(loaded.Config, pipeline.Request{Verb: verb, TargetOS: "darwin", TargetArch: "arm64"})
			require.NoError(t, err)
			for _, node := range expect.PlanNodes {
				_, ok := plan.Nodes[pipeline.NodeKey(node)]
				assert.True(t, ok, "plan is missing node %s", node)
			}
		}
	})
	return passed && !gap
}

// checkCustomCorpusScript reads a generated hook script and checks fragments.
// Scripts must be executable on POSIX hosts.
func checkCustomCorpusScript(t *testing.T, root, file string, fragments []string) {
	t.Helper()
	absent := strings.HasSuffix(file, ":absent")
	if absent {
		file = strings.TrimSuffix(file, ":absent")
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
	require.NoError(t, err, "hook script %s was not generated", file)
	for _, fragment := range fragments {
		if absent {
			assert.NotContains(t, string(data), fragment, file)
		} else {
			assert.Contains(t, string(data), fragment, file)
		}
	}
	if !absent && !strings.HasPrefix(string(data), "#!/bin/sh\n") {
		t.Errorf("%s must start with #!/bin/sh", file)
	}
	if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(file))); err == nil && !absent && info.Mode().Perm()&0o100 == 0 {
		t.Errorf("%s is not executable", file)
	}
}

func readCustomCorpusJSON(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, target), path)
}

func applyCustomCorpusEdits(t *testing.T, root string, edits customCorpusEdits) {
	t.Helper()
	for _, edit := range edits.Edits {
		path := filepath.Join(root, filepath.FromSlash(edit.File))
		data, err := os.ReadFile(path)
		require.NoError(t, err, edit.File)
		contents := string(data)
		if edit.Pattern != "" {
			expression := regexp.MustCompile(edit.Pattern)
			matches := expression.FindAllStringIndex(contents, -1)
			require.NotEmpty(t, matches, "pattern %q matched nothing in %s", edit.Pattern, edit.File)
			require.True(t, edit.All || len(matches) == 1, "pattern %q matched %d times in %s", edit.Pattern, len(matches), edit.File)
			contents = expression.ReplaceAllLiteralString(contents, edit.New)
		} else {
			count := strings.Count(contents, edit.Old)
			require.NotZero(t, count, "%q not found in %s", edit.Old, edit.File)
			require.True(t, edit.All || count == 1, "%q occurs %d times in %s", edit.Old, count, edit.File)
			contents = strings.ReplaceAll(contents, edit.Old, edit.New)
		}
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
	}
	for _, removed := range edits.Remove {
		require.NoError(t, os.RemoveAll(filepath.Join(root, filepath.FromSlash(removed))))
	}
}

// assertCustomCorpusManifest checks dotted paths against the JSON form of the
// migrated document, so expectations name the manifest field, not Go types.
func assertCustomCorpusManifest(t *testing.T, doc manifest.Document, expected map[string]any) {
	t.Helper()
	if len(expected) == 0 {
		return
	}
	data, err := json.Marshal(doc)
	require.NoError(t, err)
	var document map[string]any
	require.NoError(t, json.Unmarshal(data, &document))
	paths := make([]string, 0, len(expected))
	for path := range expected {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		actual, ok := lookupDottedPath(document, path)
		if !assert.True(t, ok, "manifest field %s is absent", path) {
			continue
		}
		assert.Equal(t, expected[path], actual, path)
	}
}

// lookupDottedPath matches each segment case-insensitively, because document
// fields are Go names (Build) while the manifest uses lower-case keys (build).
func lookupDottedPath(document map[string]any, path string) (any, bool) {
	var current any = document
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		found := false
		for key, value := range object {
			if strings.EqualFold(key, part) {
				current, found = value, true
				break
			}
		}
		if !found {
			return nil, false
		}
	}
	return current, true
}

func containsAny(values []string, text string) bool {
	for _, value := range values {
		if strings.Contains(value, text) {
			return true
		}
	}
	return false
}
