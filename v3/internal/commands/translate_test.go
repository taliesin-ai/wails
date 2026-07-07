package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeTranslateCommandDocs(t *testing.T) string {
	t.Helper()
	docs := filepath.Join(t.TempDir(), "docs")
	writeTranslateFile(t, filepath.Join(docs, "astro.config.mjs"), `import starlight from '@astrojs/starlight';
export default defineConfig({
  integrations: [starlight({
    locales: {
      root: { label: "English", lang: "en", dir: "ltr" },
      de: { label: "Deutsch", lang: "de", dir: "ltr" },
    },
  })],
})
`)
	writeTranslateFile(t, filepath.Join(docs, "package.json"), `{"dependencies":{"@astrojs/starlight":"^0"}}`)
	writeTranslateFile(t, filepath.Join(docs, "src/content/docs/index.mdx"), "Hello")
	return docs
}

func writeTranslateFile(t *testing.T, path string, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
}

func captureTranslateStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	runErr := fn()
	require.NoError(t, w.Close())
	os.Stdout = orig

	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, r)
	require.NoError(t, copyErr)
	require.NoError(t, r.Close())
	return buf.String(), runErr
}

func TestTranslateStatusModeDoesNotPrintWorkbenchWarning(t *testing.T) {
	// Given
	docs := makeTranslateCommandDocs(t)

	// When
	out, err := captureTranslateStdout(t, func() error {
		return Translate(&TranslateOptions{Dir: docs, Status: true, All: true})
	})

	// Then
	require.NoError(t, err)
	require.Contains(t, out, "de (Deutsch):")
	require.NotContains(t, strings.ToLower(out), "experimental")
	require.NotContains(t, out, "Translation workbench running")
}

func TestTranslateRejectsMultipleHelperModes(t *testing.T) {
	// Given
	docs := makeTranslateCommandDocs(t)

	// When
	err := Translate(&TranslateOptions{Dir: docs, Status: true, ListStale: true, Locale: "de"})

	// Then
	require.ErrorContains(t, err, "choose only one translate helper mode")
}
