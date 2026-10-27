package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withPuppertinoServer points the Puppertino fetcher at an httptest server and
// returns the number of requests it served.
func withPuppertinoServer(t *testing.T, status int) *atomic.Int32 {
	t.Helper()
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if status != http.StatusOK {
			http.Error(w, "nope", status)
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/dist/css/full.css"):
			_, _ = w.Write([]byte("/* puppertino */"))
		case strings.HasSuffix(r.URL.Path, "/LICENSE"):
			_, _ = w.Write([]byte("MIT"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	previous := puppertinoRawBaseURL
	puppertinoRawBaseURL = server.URL
	t.Cleanup(func() { puppertinoRawBaseURL = previous })
	return &hits
}

func writeFrontend(t *testing.T, index string) string {
	t.Helper()
	frontend := filepath.Join(t.TempDir(), "frontend")
	require.NoError(t, os.MkdirAll(frontend, 0o755))
	if index != "" {
		require.NoError(t, os.WriteFile(filepath.Join(frontend, "index.html"), []byte(index), 0o644))
	}
	return frontend
}

const puppertinoIndexFixture = `<!doctype html>
<html>
  <head>
    <link rel="stylesheet" href="/style.css"/>
  </head>
  <body>
    <button class="btn">Go</button>
  </body>
</html>`

func TestToolPuppertinoDownloadsPatchesAndIsIdempotent(t *testing.T) {
	hits := withPuppertinoServer(t, http.StatusOK)
	frontend := writeFrontend(t, puppertinoIndexFixture)

	require.NoError(t, ToolPuppertino(&ToolPuppertinoOptions{Frontend: frontend}))

	css, err := os.ReadFile(filepath.Join(frontend, "public", "puppertino", "puppertino.css"))
	require.NoError(t, err)
	assert.Equal(t, "/* puppertino */", string(css))
	license, err := os.ReadFile(filepath.Join(frontend, "public", "puppertino", "LICENSE"))
	require.NoError(t, err)
	assert.Equal(t, "MIT", string(license))

	index, err := os.ReadFile(filepath.Join(frontend, "index.html"))
	require.NoError(t, err)
	want := strings.Replace(puppertinoIndexFixture,
		`<link rel="stylesheet" href="/style.css"/>`,
		`<link rel="stylesheet" href="/style.css"/>`+"\n"+`    <link rel="stylesheet" href="/puppertino/puppertino.css"/>`, 1)
	want = strings.Replace(want, `class="btn"`, `class="p-btn p-prim-col"`, 1)
	assert.Equal(t, want, string(index))
	assert.Equal(t, int32(2), hits.Load())

	// Second run: bundled CSS is reused, nothing is re-downloaded or rewritten.
	require.NoError(t, ToolPuppertino(&ToolPuppertinoOptions{Frontend: frontend}))
	assert.Equal(t, int32(2), hits.Load())
	again, err := os.ReadFile(filepath.Join(frontend, "index.html"))
	require.NoError(t, err)
	assert.Equal(t, string(index), string(again))
}

func TestToolPuppertinoForceRefetchesCSS(t *testing.T) {
	hits := withPuppertinoServer(t, http.StatusOK)
	frontend := writeFrontend(t, "")
	cssDir := filepath.Join(frontend, "public", "puppertino")
	require.NoError(t, os.MkdirAll(cssDir, 0o755))
	cssPath := filepath.Join(cssDir, "puppertino.css")
	require.NoError(t, os.WriteFile(cssPath, []byte("old"), 0o644))

	require.NoError(t, ToolPuppertino(&ToolPuppertinoOptions{Frontend: frontend}))
	assert.Equal(t, int32(0), hits.Load(), "existing CSS must be reused without -force")

	require.NoError(t, ToolPuppertino(&ToolPuppertinoOptions{Frontend: frontend, Force: true}))
	assert.Equal(t, int32(2), hits.Load())
	css, err := os.ReadFile(cssPath)
	require.NoError(t, err)
	assert.Equal(t, "/* puppertino */", string(css))
}

func TestToolPuppertinoFailedDownloadIsNotFatal(t *testing.T) {
	withPuppertinoServer(t, http.StatusNotFound)
	frontend := writeFrontend(t, puppertinoIndexFixture)

	require.NoError(t, ToolPuppertino(&ToolPuppertinoOptions{Frontend: frontend}))
	_, err := os.Stat(filepath.Join(frontend, "public", "puppertino", "puppertino.css"))
	assert.True(t, os.IsNotExist(err))
	index, err := os.ReadFile(filepath.Join(frontend, "index.html"))
	require.NoError(t, err)
	assert.Contains(t, string(index), `href="/puppertino/puppertino.css"`)
}

func TestPatchPuppertinoIndex(t *testing.T) {
	const link = `href="/puppertino/puppertino.css"`
	for _, test := range []struct {
		name     string
		input    string
		changed  bool
		wantLink bool
		want     string
	}{
		{
			name:     "inserts link after style.css and replaces btn",
			input:    "<head>\n    <link rel=\"stylesheet\" href=\"/style.css\"/>\n</head>\n<button class=\"btn\">x</button>\n",
			changed:  true,
			wantLink: true,
			want:     "<head>\n    <link rel=\"stylesheet\" href=\"/style.css\"/>\n    <link rel=\"stylesheet\" href=\"/puppertino/puppertino.css\"/>\n</head>\n<button class=\"p-btn p-prim-col\">x</button>\n",
		},
		{
			name:     "link already present is not duplicated",
			input:    "<link rel=\"stylesheet\" href=\"/style.css\"/>\n<link rel=\"stylesheet\" href=\"/puppertino/puppertino.css\"/>\n",
			changed:  false,
			wantLink: true,
			want:     "<link rel=\"stylesheet\" href=\"/style.css\"/>\n<link rel=\"stylesheet\" href=\"/puppertino/puppertino.css\"/>\n",
		},
		{
			name:     "no style.css link means no insertion but btn still replaced",
			input:    "<button class=\"btn\">x</button>\n",
			changed:  true,
			wantLink: false,
			want:     "<button class=\"p-btn p-prim-col\">x</button>\n",
		},
		{
			name:     "already patched file is unchanged",
			input:    "<link rel=\"stylesheet\" href=\"/style.css\"/>\n<link rel=\"stylesheet\" href=\"/puppertino/puppertino.css\"/>\n<button class=\"p-btn p-prim-col\">x</button>\n",
			changed:  false,
			wantLink: true,
			want:     "<link rel=\"stylesheet\" href=\"/style.css\"/>\n<link rel=\"stylesheet\" href=\"/puppertino/puppertino.css\"/>\n<button class=\"p-btn p-prim-col\">x</button>\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "index.html")
			require.NoError(t, os.WriteFile(path, []byte(test.input), 0o644))
			changed, err := patchPuppertinoIndex(path)
			require.NoError(t, err)
			assert.Equal(t, test.changed, changed)
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, test.want, string(data))
			assert.Equal(t, test.wantLink, strings.Contains(string(data), link))
		})
	}
}

func TestPatchPuppertinoIndexMissingFileIsNoop(t *testing.T) {
	changed, err := patchPuppertinoIndex(filepath.Join(t.TempDir(), "missing.html"))
	require.NoError(t, err)
	assert.False(t, changed)
}
