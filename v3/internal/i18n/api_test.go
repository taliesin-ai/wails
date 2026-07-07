package i18n

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/i18n/backend/starlight"
	"github.com/wailsapp/wails/v3/internal/i18n/catalog"
	"github.com/wailsapp/wails/v3/internal/i18n/engine"
	"github.com/wailsapp/wails/v3/internal/i18n/llm"
	"github.com/wailsapp/wails/v3/internal/i18n/review"
)

func makeAPIApp(t *testing.T) (*App, string) {
	t.Helper()
	docs := makeHelperDocs(t)
	be := starlight.New(docs)
	cacheDir := filepath.Join(docs, ".translation-cache")
	rev := review.NewStore(cacheDir)
	return &App{
		be:       be,
		cacheDir: cacheDir,
		cat:      catalog.New(be, cacheDir, rev),
		rev:      rev,
		eng:      engine.New(be, cacheDir, nil),
		llm:      llm.NewManager(),
	}, docs
}

func TestAPIGetEndpointsRejectUnsafeSourcePaths(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"file", "/api/file?locale=de&path=../secret.mdx"},
		{"blocks", "/api/blocks?locale=de&path=../secret.mdx"},
		{"diff", "/api/diff?locale=de&path=../secret.mdx"},
		{"source diff", "/api/sourcediff?locale=de&path=../secret.mdx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			app, _ := makeAPIApp(t)
			handlers := map[string]http.HandlerFunc{
				"file":        app.handleFile,
				"blocks":      app.handleBlocks,
				"diff":        app.handleDiff,
				"source diff": app.handleSourceDiff,
			}

			// When
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			handlers[tt.name](rr, req)

			// Then
			require.Equal(t, http.StatusBadRequest, rr.Code)
			require.Contains(t, rr.Body.String(), "invalid source-relative path")
		})
	}
}

func TestAPISaveFileRejectsUnsafeSourcePathBeforeWriting(t *testing.T) {
	// Given
	app, docs := makeAPIApp(t)
	body := bytes.NewBufferString(`{"locale":"de","path":"../secret.mdx","content":"escape"}`)

	// When
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/file", body)
	app.handleFile(rr, req)

	// Then
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "invalid source-relative path")
	require.NoFileExists(t, filepath.Join(docs, "src", "content", "docs", "secret.mdx"))
}

func TestAPITranslateRejectsUnsafePathScopeBeforeStartingJob(t *testing.T) {
	// Given
	app, _ := makeAPIApp(t)
	body := bytes.NewBufferString(`{"locale":"de","scope":"paths","paths":["../secret.mdx"]}`)

	// When
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/translate", body)
	app.handleTranslate(rr, req)

	// Then
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "invalid source-relative path")
}
