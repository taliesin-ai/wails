package i18n

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/cache"
	"github.com/wailsapp/wails/v3/internal/i18n/catalog"
	"github.com/wailsapp/wails/v3/internal/i18n/review"
)

func readFileString(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

func newTestMCPServer(t *testing.T, dir string) *mcpServer {
	t.Helper()
	s := newMCPServer(MCPOptions{Dir: dir})
	s.logw = io.Discard
	return s
}

func mcpTestRequest(t *testing.T, method string, params any) *mcpJSONRPCRequest {
	t.Helper()
	req := &mcpJSONRPCRequest{JSONRPC: "2.0", ID: json.RawMessage("1"), Method: method}
	if params != nil {
		raw, err := json.Marshal(params)
		require.NoError(t, err)
		req.Params = raw
	}
	return req
}

// mcpTestCall invokes a tool through the dispatch layer and returns the text
// payload plus the isError flag.
func mcpTestCall(t *testing.T, s *mcpServer, tool string, args map[string]any) (string, bool) {
	t.Helper()
	resp := s.handleMessage(mcpTestRequest(t, "tools/call", map[string]any{"name": tool, "arguments": args}))
	require.NotNil(t, resp)
	require.Nil(t, resp.Error)
	result, ok := resp.Result.(map[string]any)
	require.True(t, ok)
	content, ok := result["content"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, content, 1)
	text, ok := content[0]["text"].(string)
	require.True(t, ok)
	isError, ok := result["isError"].(bool)
	require.True(t, ok)
	return text, isError
}

// mcpTestRPCResponse is the wire shape tests decode from serve() output lines.
type mcpTestRPCResponse struct {
	ID     json.RawMessage `json:"id"`
	Result struct {
		ProtocolVersion string         `json:"protocolVersion"`
		ServerInfo      map[string]any `json:"serverInfo"`
		Instructions    string         `json:"instructions"`
		Content         []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	} `json:"result"`
	Error *mcpJSONRPCError `json:"error"`
}

func TestMCPServeHandshake(t *testing.T) {
	// Given
	docs := makeHelperDocs(t)
	s := newTestMCPServer(t, docs)
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
	}, "\n")
	var out bytes.Buffer

	// When
	err := s.serve(strings.NewReader(input), &out)

	// Then
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	require.Len(t, lines, 2, "the notification must produce no output")
	var init mcpTestRPCResponse
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &init))
	require.Equal(t, "2025-03-26", init.Result.ProtocolVersion, "client protocol version must be echoed")
	require.Equal(t, "wails-translate", init.Result.ServerInfo["name"])
	require.NotEmpty(t, init.Result.Instructions)
}

func TestMCPToolsListNames(t *testing.T) {
	// Given
	s := newTestMCPServer(t, makeHelperDocs(t))

	// When
	resp := s.handleMessage(mcpTestRequest(t, "tools/list", nil))

	// Then
	require.NotNil(t, resp)
	require.Nil(t, resp.Error)
	tools := resp.Result.(map[string]any)["tools"].([]map[string]any)
	var names []string
	for _, tool := range tools {
		names = append(names, tool["name"].(string))
	}
	require.Equal(t, []string{
		"setup_docs", "translation_status", "list_files",
		"get_translation_context", "write_translation", "add_locale",
	}, names)
}

func TestMCPUnknownMethodErrors(t *testing.T) {
	// Given
	s := newTestMCPServer(t, makeHelperDocs(t))

	// When
	resp := s.handleMessage(mcpTestRequest(t, "resources/list", nil))

	// Then
	require.NotNil(t, resp)
	require.NotNil(t, resp.Error)
	require.Equal(t, mcpCodeMethodNotFound, resp.Error.Code)
}

func TestMCPTranslationStatusAndListFiles(t *testing.T) {
	// Given
	s := newTestMCPServer(t, makeHelperDocs(t))

	// When
	statusText, statusErr := mcpTestCall(t, s, "translation_status", map[string]any{"locale": "de"})
	filesText, filesErr := mcpTestCall(t, s, "list_files", map[string]any{"locale": "de"})

	// Then
	require.False(t, statusErr)
	var status catalog.LocaleStatus
	require.NoError(t, json.Unmarshal([]byte(statusText), &status))
	require.Equal(t, 3, status.Total)
	require.Equal(t, 1, status.UpToDate)
	require.Equal(t, 1, status.Stale)
	require.Equal(t, 1, status.Missing)

	require.False(t, filesErr)
	var files []catalog.FileStatus
	require.NoError(t, json.Unmarshal([]byte(filesText), &files))
	var paths []string
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	require.Equal(t, []string{"guide/b.mdx", "guide/c.mdx"}, paths, "needs_work = stale + missing")
}

func TestMCPGetTranslationContextIncludesDrift(t *testing.T) {
	// Given a stale file whose snapshot captures the source as translated
	s := newTestMCPServer(t, makeHelperDocs(t))
	require.NotNil(t, s.session)
	loc, err := s.session.locale("de")
	require.NoError(t, err)
	require.NoError(t, s.session.eng.Stamp(loc, backend.SourceRef{Path: "guide/b.mdx"}, []byte("beta v1")))

	// When
	text, isErr := mcpTestCall(t, s, "get_translation_context", map[string]any{"locale": "de", "path": "guide/b.mdx"})

	// Then
	require.False(t, isErr)
	var ctx mcpTranslationContext
	require.NoError(t, json.Unmarshal([]byte(text), &ctx))
	require.Equal(t, catalog.Stale, ctx.Freshness)
	require.NotNil(t, ctx.Source)
	require.Equal(t, "beta v2", *ctx.Source)
	require.NotNil(t, ctx.Target)
	require.Equal(t, "beta de", *ctx.Target)
	require.NotEmpty(t, ctx.Glossary)
	require.NotNil(t, ctx.Drift, "stale + snapshot must yield a drift diff")
	require.Equal(t, 1, ctx.Drift.ChangedCount)
	require.Equal(t, "beta v1", ctx.Drift.ChangedBlocks[0].Old)
	require.Equal(t, "beta v2", ctx.Drift.ChangedBlocks[0].New)
}

func TestMCPWriteTranslationHappyPathFlipsFileUpToDate(t *testing.T) {
	// Given
	docs := makeHelperDocs(t)
	s := newTestMCPServer(t, docs)

	// When
	text, isErr := mcpTestCall(t, s, "write_translation", map[string]any{
		"locale": "de", "path": "guide/c.mdx", "content": "gamma auf Deutsch", "model": "claude-test",
	})

	// Then
	require.False(t, isErr, text)
	var result struct {
		Written      string   `json:"written"`
		ReviewStatus string   `json:"reviewStatus"`
		Warnings     []string `json:"warnings"`
	}
	require.NoError(t, json.Unmarshal([]byte(text), &result))
	require.Equal(t, "needs_review", result.ReviewStatus)
	require.Empty(t, result.Warnings)

	target := filepath.Join(docs, "src", "content", "docs", "de", "guide", "c.mdx")
	require.Equal(t, target, result.Written)
	data, err := readFileString(target)
	require.NoError(t, err)
	require.Equal(t, "gamma auf Deutsch", data)

	cacheDir := filepath.Join(docs, ".translation-cache")
	entries, err := cache.Load(cacheDir, "de")
	require.NoError(t, err)
	entry := entries["guide/c.mdx"]
	require.Equal(t, cache.Hash([]byte("gamma")), entry.Hash, "stamp must hash the current source")
	require.NotEmpty(t, entry.SourceSnapshot)

	record := review.NewStore(cacheDir).Get("de", "guide/c.mdx")
	require.Equal(t, "needs_review", record.Status)
	require.Equal(t, review.Provenance{Method: "mcp", Model: "claude-test"}, record.Provenance)

	filesText, _ := mcpTestCall(t, s, "list_files", map[string]any{"locale": "de", "filter": "up_to_date"})
	require.Contains(t, filesText, "guide/c.mdx")
}

func TestMCPWriteTranslationRejectsFenceMismatchUnlessForced(t *testing.T) {
	// Given a source page containing a code fence
	docs := makeHelperDocs(t)
	writeFile(t, filepath.Join(docs, "src/content/docs/guide/d.mdx"), "intro\n\n```go\ncode\n```\n")
	s := newTestMCPServer(t, docs)
	truncated := "Einleitung\n" // fence lost: looks like truncated output
	target := filepath.Join(docs, "src", "content", "docs", "de", "guide", "d.mdx")

	// When / Then: rejected, nothing written
	text, isErr := mcpTestCall(t, s, "write_translation", map[string]any{
		"locale": "de", "path": "guide/d.mdx", "content": truncated,
	})
	require.True(t, isErr)
	require.Contains(t, text, "code fence count differs")
	_, err := readFileString(target)
	require.Error(t, err, "a rejected write must leave no file behind")

	// When / Then: force writes anyway
	_, isErr = mcpTestCall(t, s, "write_translation", map[string]any{
		"locale": "de", "path": "guide/d.mdx", "content": truncated, "force": true,
	})
	require.False(t, isErr)
	data, err := readFileString(target)
	require.NoError(t, err)
	require.Equal(t, truncated, data)
}

func TestMCPWriteTranslationRejectsPathTraversal(t *testing.T) {
	// Given
	s := newTestMCPServer(t, makeHelperDocs(t))

	// When
	text, isErr := mcpTestCall(t, s, "write_translation", map[string]any{
		"locale": "de", "path": "../evil.mdx", "content": "x",
	})

	// Then
	require.True(t, isErr, text)
}

func TestMCPServeLargeSingleLineWriteRequest(t *testing.T) {
	// Given a write_translation request comfortably past bufio.Scanner's 64KB
	// default token limit (the framing regression this server must not have)
	docs := makeHelperDocs(t)
	writeFile(t, filepath.Join(docs, "src/content/docs/guide/big.mdx"), "big source")
	s := newTestMCPServer(t, docs)
	content := strings.Repeat("große Wörter ", 12000) // ~160KB
	call, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": 7, "method": "tools/call",
		"params": map[string]any{
			"name": "write_translation",
			"arguments": map[string]any{
				"locale": "de", "path": "guide/big.mdx", "content": content,
			},
		},
	})
	require.NoError(t, err)
	require.Greater(t, len(call), 64*1024)
	var out bytes.Buffer

	// When
	require.NoError(t, s.serve(bytes.NewReader(append(call, '\n')), &out))

	// Then
	var resp mcpTestRPCResponse
	require.NoError(t, json.Unmarshal(out.Bytes(), &resp))
	require.Nil(t, resp.Error)
	require.False(t, resp.Result.IsError, "large request must round-trip: %s", out.String())
	data, err := readFileString(filepath.Join(docs, "src", "content", "docs", "de", "guide", "big.mdx"))
	require.NoError(t, err)
	require.Equal(t, content, data)
}

func TestMCPToolsBeforeSetupAskForSetupDocs(t *testing.T) {
	// Given a server started where no docs site exists
	s := newTestMCPServer(t, t.TempDir())
	require.Nil(t, s.session)

	// When
	text, isErr := mcpTestCall(t, s, "translation_status", map[string]any{})

	// Then
	require.True(t, isErr)
	require.Contains(t, text, "setup_docs")
}

func TestMCPSetupDocsClonesAndUnblocksTools(t *testing.T) {
	// Given no docs site anywhere near cwd, and the clone seams pointed at a
	// local fake remote and a temp cache
	t.Chdir(t.TempDir())
	remote := makeFakeWailsRemote(t)
	cacheDir := filepath.Join(t.TempDir(), "wails-src")
	origURL, origDir := wailsRepoURL, wailsDocsCacheDir
	wailsRepoURL = remote
	wailsDocsCacheDir = func() (string, error) { return cacheDir, nil }
	t.Cleanup(func() { wailsRepoURL, wailsDocsCacheDir = origURL, origDir })
	s := newTestMCPServer(t, "")
	require.Nil(t, s.session)

	// When
	text, isErr := mcpTestCall(t, s, "setup_docs", map[string]any{})

	// Then
	require.False(t, isErr, text)
	var setup struct {
		DocsRoot string           `json:"docsRoot"`
		Backend  string           `json:"backend"`
		Cloned   bool             `json:"cloned"`
		Locales  []backend.Locale `json:"locales"`
	}
	require.NoError(t, json.Unmarshal([]byte(text), &setup))
	require.True(t, setup.Cloned)
	require.Equal(t, filepath.Join(cacheDir, "docs"), setup.DocsRoot)
	require.Equal(t, "starlight", setup.Backend)
	require.Len(t, setup.Locales, 2)

	// And other tools are unblocked against the cloned docs
	statusText, statusErr := mcpTestCall(t, s, "translation_status", map[string]any{"locale": "de"})
	require.False(t, statusErr, statusText)
	require.Contains(t, statusText, `"missing": 1`)

	// And a second setup_docs reuses the session (no re-clone)
	text2, isErr2 := mcpTestCall(t, s, "setup_docs", map[string]any{})
	require.False(t, isErr2)
	require.Contains(t, text2, `"cloned": false`)
}
