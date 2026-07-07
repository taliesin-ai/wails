// Package i18n implements `wails3 translate`: a Crowdin-style documentation
// translation workbench. Like v3/internal/setupwizard, it runs a local HTTP
// server, serves an embedded frontend, and opens the user's browser. The Go
// core is split behind the DocsBackend seam (see SPEC.md section 3).
package i18n

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/internal/browser"
	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/backend/starlight"
	"github.com/wailsapp/wails/v3/internal/i18n/catalog"
	"github.com/wailsapp/wails/v3/internal/i18n/engine"
	"github.com/wailsapp/wails/v3/internal/i18n/llm"
	"github.com/wailsapp/wails/v3/internal/i18n/review"
)

//go:embed all:frontend/dist
var frontendFS embed.FS

//go:embed frontend/locales/*.json
var localesFS embed.FS

// Options configure a translate session.
type Options struct {
	Dir       string // docs root override; auto-detected when empty
	Locale    string // jump straight to a locale workspace
	Headless  bool   // engine-only, no window (CI bridge)
	NoBrowser bool   // start the server but don't open a browser (tests/CI)
	Clone     bool   // when no docs site is found, clone Wails without prompting
}

// App is a running translate workbench.
type App struct {
	opts     Options
	be       backend.DocsBackend
	cacheDir string
	cat      *catalog.Catalog
	rev      *review.Store
	eng      *engine.Engine
	llm      *llm.Manager
	preview  *previewManager

	server   *http.Server
	done     chan struct{}
	doneOnce sync.Once

	jobsMu sync.Mutex
	jobs   map[string]*translateJob
}

// Run starts the workbench: detect the backend, wire the core, serve the UI and
// open the browser. It blocks until the window is closed.
func (a *App) Run(opts Options) error {
	a.opts = opts
	a.done = make(chan struct{})

	root, conf := findDocsRoot(opts.Dir)
	if root == "" {
		var err error
		root, conf, err = a.offerCloneWailsDocs(opts)
		if err != nil {
			return err
		}
	}
	if root == "" {
		return fmt.Errorf("could not find a docs site (Astro Starlight). Run from your project, pass --dir <docs root>, or use --clone to fetch the Wails docs")
	}
	a.be = starlight.New(root)
	a.cacheDir = filepath.Join(a.be.Root(), ".translation-cache")
	a.rev = review.NewStore(a.cacheDir)
	a.cat = catalog.New(a.be, a.cacheDir, a.rev)
	a.eng = engine.New(a.be, a.cacheDir, loadGlossary(a.be.Root()))
	a.llm = llm.NewManager()
	a.preview = newPreviewManager(a.be)

	if opts.Headless {
		return a.runHeadless()
	}

	// Prefer a stable port so restarting `wails3 translate` keeps the same URL and
	// an already-open browser tab keeps working (no dead-tab churn while
	// iterating). Fall back to an ephemeral port if it's already taken.
	listener, err := net.Listen("tcp", "127.0.0.1:9737")
	if err != nil {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return fmt.Errorf("find port: %w", err)
		}
	}
	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	mux := http.NewServeMux()
	a.routes(mux)
	a.server = &http.Server{Handler: recoverMiddleware(mux)}

	go func() {
		if err := a.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "translate server error: %v\n", err)
		}
	}()

	fmt.Printf("Translation workbench running at %s\n", url)
	fmt.Printf("Detected docs backend: %s (%s) at %s\n", a.be.Name(), conf, root)

	if !opts.NoBrowser {
		if err := browser.OpenURL(url); err != nil {
			fmt.Printf("Open %s in your browser to begin.\n", url)
		}
	}

	<-a.done
	a.preview.stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.server.Shutdown(ctx)
}

func (a *App) signalDone() { a.doneOnce.Do(func() { close(a.done) }) }

// routes registers the HTTP API and the embedded frontend.
func (a *App) routes(mux *http.ServeMux) {
	mux.HandleFunc("/api/backend", a.handleBackend)
	mux.HandleFunc("/api/locales", a.handleLocales)
	mux.HandleFunc("/api/tree", a.handleTree)
	mux.HandleFunc("/api/file", a.handleFile)     // GET ?locale&path ; POST save
	mux.HandleFunc("/api/blocks", a.handleBlocks) // GET ?locale&path
	mux.HandleFunc("/api/diff", a.handleDiff)
	mux.HandleFunc("/api/sourcediff", a.handleSourceDiff)
	mux.HandleFunc("/api/review", a.handleReview) // POST set viewed/approved
	mux.HandleFunc("/api/models", a.handleModels)
	mux.HandleFunc("/api/llm/config", a.handleLLMConfig)
	mux.HandleFunc("/api/llm/models", a.handleProviderModels)
	mux.HandleFunc("/api/llm/credentials/export", a.handleExportCredentials)
	mux.HandleFunc("/api/llm/credentials/load", a.handleLoadCredentials)
	mux.HandleFunc("/api/llm/credentials/default", a.handleCredentialsDefault)
	mux.HandleFunc("/api/translate", a.handleTranslate)
	mux.HandleFunc("/api/translate/events", a.handleTranslateEvents)
	mux.HandleFunc("/api/locale/add", a.handleAddLocale)
	mux.HandleFunc("/api/preview/start", a.handlePreviewStart)
	mux.HandleFunc("/api/preview/stop", a.handlePreviewStop)
	mux.HandleFunc("/api/preview/events", a.handlePreviewEvents)
	mux.HandleFunc("/api/submit", a.handleSubmit)
	mux.HandleFunc("/api/i18n/", a.handleUICatalog) // /api/i18n/<locale>
	mux.HandleFunc("/api/uilocales", a.handleUILocales)
	mux.HandleFunc("/api/close", a.handleClose)

	dist, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(dist))
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(dist, p); err != nil {
			r.URL.Path = "/" // SPA fallback
		}
		if ct := staticContentType(r.URL.Path); ct != "" {
			rw.Header().Set("Content-Type", ct)
		}
		fileServer.ServeHTTP(rw, r)
	})
}

func (a *App) handleClose(rw http.ResponseWriter, r *http.Request) {
	writeJSON(rw, map[string]string{"status": "closing"})
	a.signalDone()
}

// findDocsRoot locates a Starlight docs root. With an explicit override it is
// validated and used; otherwise it probes cwd, cwd/docs, and walks up parents
// (and their docs/ subdir).
func findDocsRoot(override string) (string, backend.Confidence) {
	if override != "" {
		if ok, conf := starlight.Detect(override); ok {
			abs, _ := filepath.Abs(override)
			return abs, conf
		}
		// Maybe they pointed at the repo root containing docs/.
		if ok, conf := starlight.Detect(filepath.Join(override, "docs")); ok {
			abs, _ := filepath.Abs(filepath.Join(override, "docs"))
			return abs, conf
		}
		return "", backend.NoConfidence
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", backend.NoConfidence
	}
	for {
		for _, cand := range []string{dir, filepath.Join(dir, "docs")} {
			if ok, conf := starlight.Detect(cand); ok {
				abs, _ := filepath.Abs(cand)
				return abs, conf
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", backend.NoConfidence
		}
		dir = parent
	}
}

// recoverMiddleware ensures a panic in any handler returns a 500 (and keeps the
// connection well-behaved) instead of aborting the connection, so a single bad
// request can never look like "server gone" to the frontend.
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Fprintf(os.Stderr, "translate: recovered from panic on %s: %v\n", r.URL.Path, rec)
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(rw, `{"error":"internal error: %v"}`, rec)
			}
		}()
		next.ServeHTTP(rw, r)
	})
}

// staticContentType pins MIME types so rendering never depends on the host's
// MIME database (mirrors setupwizard).
func staticContentType(path string) string {
	switch {
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".html"), path == "/":
		return "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".js"), strings.HasSuffix(path, ".mjs"):
		return "text/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".json"):
		return "application/json"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".webp"):
		return "image/webp"
	case strings.HasSuffix(path, ".ico"):
		return "image/x-icon"
	case strings.HasSuffix(path, ".woff2"):
		return "font/woff2"
	}
	return ""
}
