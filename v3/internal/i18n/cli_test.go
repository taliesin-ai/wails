package i18n

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/i18n/cache"
)

func makeHelperDocs(t *testing.T) string {
	t.Helper()
	docs := filepath.Join(t.TempDir(), "docs")
	writeFile(t, filepath.Join(docs, "astro.config.mjs"), `import starlight from '@astrojs/starlight';
export default defineConfig({
  integrations: [starlight({
    locales: {
      root: { label: "English", lang: "en", dir: "ltr" },
      de: { label: "Deutsch", lang: "de", dir: "ltr" },
      fr: { label: "Français", lang: "fr", dir: "ltr" },
    },
  })],
})
`)
	writeFile(t, filepath.Join(docs, "package.json"), `{"dependencies":{"@astrojs/starlight":"^0"}}`)
	writeFile(t, filepath.Join(docs, "src/content/docs/guide/a.mdx"), "alpha")
	writeFile(t, filepath.Join(docs, "src/content/docs/guide/b.mdx"), "beta v2")
	writeFile(t, filepath.Join(docs, "src/content/docs/guide/c.mdx"), "gamma")
	writeFile(t, filepath.Join(docs, "src/content/docs/de/guide/a.mdx"), "alpha de")
	writeFile(t, filepath.Join(docs, "src/content/docs/de/guide/b.mdx"), "beta de")

	c := cache.Cache{
		"guide/a.mdx": {Hash: cache.Hash([]byte("alpha"))},
		"guide/b.mdx": {Hash: cache.Hash([]byte("beta v1"))},
	}
	require.NoError(t, c.Save(filepath.Join(docs, ".translation-cache"), "de"))
	return docs
}

func writeFile(t *testing.T, path string, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
}

func TestRunHelperStatusAllPrintsOneLinePerNonSourceLocale(t *testing.T) {
	// Given
	docs := makeHelperDocs(t)
	var out bytes.Buffer

	// When
	err := RunHelper(HelperOptions{Mode: HelperModeStatus, Dir: docs, All: true, Stdout: &out})

	// Then
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	require.Len(t, lines, 2)
	require.Equal(t, "de (Deutsch): up-to-date=1 stale=1 missing=1 orphaned=0 / 3", lines[0])
	require.Equal(t, "fr (Français): up-to-date=0 stale=0 missing=3 orphaned=0 / 3", lines[1])
}

func TestRunHelperListStalePrintsOnlyStaleAndMissingPaths(t *testing.T) {
	// Given
	docs := makeHelperDocs(t)
	var out bytes.Buffer

	// When
	err := RunHelper(HelperOptions{Mode: HelperModeListStale, Dir: docs, Locale: "de", Stdout: &out})

	// Then
	require.NoError(t, err)
	require.Equal(t, "guide/b.mdx\nguide/c.mdx", strings.TrimSpace(out.String()))
}

func TestRunHelperListStaleAllPrefixesLocale(t *testing.T) {
	// Given
	docs := makeHelperDocs(t)
	var out bytes.Buffer

	// When
	err := RunHelper(HelperOptions{Mode: HelperModeListStale, Dir: docs, All: true, Stdout: &out})

	// Then
	require.NoError(t, err)
	require.Equal(t, strings.Join([]string{
		"de\tguide/b.mdx",
		"de\tguide/c.mdx",
		"fr\tguide/a.mdx",
		"fr\tguide/b.mdx",
		"fr\tguide/c.mdx",
	}, "\n"), strings.TrimSpace(out.String()))
}

func TestRunHelperListStaleUnknownLocaleReturnsClearError(t *testing.T) {
	// Given
	docs := makeHelperDocs(t)

	// When
	err := RunHelper(HelperOptions{Mode: HelperModeListStale, Dir: docs, Locale: "zz"})

	// Then
	require.ErrorContains(t, err, `unknown locale "zz"`)
}

func TestRunHelperRecordStampsCacheAndSourceSnapshot(t *testing.T) {
	// Given
	docs := makeHelperDocs(t)
	writeFile(t, filepath.Join(docs, "src/content/docs/guide/record.mdx"), "record source")
	writeFile(t, filepath.Join(docs, "src/content/docs/de/guide/record.mdx"), "record target")

	// When
	err := RunHelper(HelperOptions{Mode: HelperModeRecord, Dir: docs, Locale: "de", File: "guide/record.mdx"})

	// Then
	require.NoError(t, err)
	c, err := cache.Load(filepath.Join(docs, ".translation-cache"), "de")
	require.NoError(t, err)
	require.Equal(t, cache.Hash([]byte("record source")), c["guide/record.mdx"].Hash)
	require.Equal(t, filepath.Join(cache.SnapshotDir, "guide/record.mdx"), c["guide/record.mdx"].SourceSnapshot)
	snapshotPath := filepath.Join(docs, ".translation-cache", c["guide/record.mdx"].SourceSnapshot)
	snapshot, err := os.ReadFile(snapshotPath)
	require.NoError(t, err)
	require.Equal(t, "record source", string(snapshot))
}

func TestRunHelperRecordRejectsUnsafeFilePaths(t *testing.T) {
	tests := []string{
		"",
		"/abs.mdx",
		"../secret.mdx",
		"guide/../secret.mdx",
		`guide\secret.mdx`,
		"guide//a.mdx",
		"./guide/a.mdx",
		"guide/./a.mdx",
		"C:/secret.mdx",
	}

	for _, file := range tests {
		t.Run(file, func(t *testing.T) {
			// Given
			docs := makeHelperDocs(t)

			// When
			err := RunHelper(HelperOptions{Mode: HelperModeRecord, Dir: docs, Locale: "de", File: file})

			// Then
			require.ErrorContains(t, err, "invalid source-relative path")
		})
	}
}
