package i18n

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/engine"
	"github.com/wailsapp/wails/v3/internal/i18n/llm"
	"github.com/wailsapp/wails/v3/internal/i18n/review"
)

// writeJSON writes v as a JSON response.
func writeJSON(rw http.ResponseWriter, v any) {
	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(v)
}

func writeErr(rw http.ResponseWriter, code int, msg string) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)
	_ = json.NewEncoder(rw).Encode(map[string]string{"error": msg})
}

// locale resolves a locale code to its declared backend.Locale.
func (a *App) locale(code string) (backend.Locale, bool) {
	locs, err := a.be.ListLocales()
	if err != nil {
		return backend.Locale{}, false
	}
	for _, l := range locs {
		if l.Code == code {
			return l, true
		}
	}
	return backend.Locale{}, false
}

// GET /api/backend
func (a *App) handleBackend(rw http.ResponseWriter, r *http.Request) {
	writeJSON(rw, map[string]any{
		"name":         a.be.Name(),
		"root":         a.be.Root(),
		"sourceLocale": a.be.SourceLocale(),
		"format":       string(a.be.Format()),
		"detected":     true,
	})
}

// GET /api/locales
func (a *App) handleLocales(rw http.ResponseWriter, r *http.Request) {
	statuses, err := a.cat.LocaleStatuses()
	if err != nil {
		writeErr(rw, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(rw, statuses)
}

// GET /api/tree?locale=de
func (a *App) handleTree(rw http.ResponseWriter, r *http.Request) {
	loc, ok := a.locale(r.URL.Query().Get("locale"))
	if !ok {
		writeErr(rw, http.StatusBadRequest, "unknown locale")
		return
	}
	files, err := a.cat.FileStatuses(loc)
	if err != nil {
		writeErr(rw, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(rw, files)
}

type fileView struct {
	Path      string         `json:"path"`
	Source    string         `json:"source"`
	Target    string         `json:"target"`
	HasTarget bool           `json:"hasTarget"`
	SrcBlocks []engine.Block `json:"sourceBlocks"`
	TgtBlocks []engine.Block `json:"targetBlocks"`
}

// GET /api/file?locale=de&path=index.mdx  -> source+target+blocks
// POST /api/file {locale,path,content}    -> save target
func (a *App) handleFile(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		a.handleSaveFile(rw, r)
		return
	}
	loc, ok := a.locale(r.URL.Query().Get("locale"))
	if !ok {
		writeErr(rw, http.StatusBadRequest, "unknown locale")
		return
	}
	src, err := backend.ParseSourceRef(r.URL.Query().Get("path"))
	if err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}

	srcDoc, err := a.be.ReadDoc(backend.DocRef{Locale: backend.Locale{IsSource: true}, Source: src})
	if err != nil {
		writeErr(rw, http.StatusNotFound, "source not found: "+err.Error())
		return
	}
	view := fileView{Path: src.Path, Source: srcDoc.Raw, SrcBlocks: engine.Segment(srcDoc.Body)}
	if tgtDoc, err := a.be.ReadDoc(backend.DocRef{Locale: loc, Source: src}); err == nil {
		view.Target = tgtDoc.Raw
		view.HasTarget = true
		view.TgtBlocks = engine.Segment(tgtDoc.Body)
	}
	writeJSON(rw, view)
}

type saveReq struct {
	Locale  string `json:"locale"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (a *App) handleSaveFile(rw http.ResponseWriter, r *http.Request) {
	var req saveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	loc, ok := a.locale(req.Locale)
	if !ok {
		writeErr(rw, http.StatusBadRequest, "unknown locale")
		return
	}
	src, err := backend.ParseSourceRef(req.Path)
	if err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	// Write raw content as the target file.
	if err := a.be.WriteDoc(backend.DocRef{Locale: loc, Source: src}, backend.Doc{Raw: req.Content, Body: req.Content}); err != nil {
		writeErr(rw, http.StatusInternalServerError, err.Error())
		return
	}
	// Re-stamp the cache against the current source so freshness is up to date.
	if srcDoc, err := a.be.ReadDoc(backend.DocRef{Locale: backend.Locale{IsSource: true}, Source: src}); err == nil {
		_ = a.eng.Stamp(loc, src, []byte(srcDoc.Raw))
	}
	// A hand-saved file moves to needs_review unless already approved.
	if status, _, ok := a.rev.Status(loc.Code, src.Path); !ok || status == "untranslated" || status == "machine" {
		_ = a.rev.SetStatus(loc.Code, src.Path, "needs_review", review.Provenance{Method: "manual"})
	}
	writeJSON(rw, map[string]string{"status": "saved"})
}

// GET /api/blocks?locale=de&path=index.mdx
func (a *App) handleBlocks(rw http.ResponseWriter, r *http.Request) {
	loc, ok := a.locale(r.URL.Query().Get("locale"))
	if !ok {
		writeErr(rw, http.StatusBadRequest, "unknown locale")
		return
	}
	src, err := backend.ParseSourceRef(r.URL.Query().Get("path"))
	if err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	srcDoc, err := a.be.ReadDoc(backend.DocRef{Locale: backend.Locale{IsSource: true}, Source: src})
	if err != nil {
		writeErr(rw, http.StatusNotFound, err.Error())
		return
	}
	out := map[string]any{"sourceBlocks": engine.Segment(srcDoc.Body)}
	if tgtDoc, err := a.be.ReadDoc(backend.DocRef{Locale: loc, Source: src}); err == nil {
		out["targetBlocks"] = engine.Segment(tgtDoc.Body)
	}
	writeJSON(rw, out)
}

// GET /api/diff?locale=de&path=index.mdx  -> source vs generated translation
func (a *App) handleDiff(rw http.ResponseWriter, r *http.Request) {
	loc, ok := a.locale(r.URL.Query().Get("locale"))
	if !ok {
		writeErr(rw, http.StatusBadRequest, "unknown locale")
		return
	}
	src, err := backend.ParseSourceRef(r.URL.Query().Get("path"))
	if err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	srcDoc, err := a.be.ReadDoc(backend.DocRef{Locale: backend.Locale{IsSource: true}, Source: src})
	if err != nil {
		writeErr(rw, http.StatusNotFound, err.Error())
		return
	}
	var tgtBlocks []engine.Block
	if tgtDoc, err := a.be.ReadDoc(backend.DocRef{Locale: loc, Source: src}); err == nil {
		tgtBlocks = engine.Segment(tgtDoc.Body)
	}
	writeJSON(rw, engine.AlignDiff(engine.Segment(srcDoc.Body), tgtBlocks))
}

// GET /api/sourcediff?locale=de&path=index.mdx -> old source snapshot vs current
func (a *App) handleSourceDiff(rw http.ResponseWriter, r *http.Request) {
	loc, ok := a.locale(r.URL.Query().Get("locale"))
	if !ok {
		writeErr(rw, http.StatusBadRequest, "unknown locale")
		return
	}
	src, err := backend.ParseSourceRef(r.URL.Query().Get("path"))
	if err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	old, cur, err := a.eng.SourceDrift(loc, src)
	if err != nil {
		writeErr(rw, http.StatusOK, "") // tolerate; frontend shows "no snapshot"
		return
	}
	writeJSON(rw, engine.AlignDiff(engine.Segment(old), engine.Segment(cur)))
}

type reviewReq struct {
	Locale   string `json:"locale"`
	Path     string `json:"path"`
	Viewed   *bool  `json:"viewed,omitempty"`
	Approved *bool  `json:"approved,omitempty"`
}

// POST /api/review
func (a *App) handleReview(rw http.ResponseWriter, r *http.Request) {
	var req reviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	if req.Viewed != nil {
		_ = a.rev.SetViewed(req.Locale, req.Path, *req.Viewed)
	}
	if req.Approved != nil {
		_ = a.rev.SetApproved(req.Locale, req.Path, *req.Approved)
	}
	writeJSON(rw, a.rev.Get(req.Locale, req.Path))
}

// GET /api/models?lang=ja
func (a *App) handleModels(rw http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	writeJSON(rw, map[string]any{
		"recommendation": a.llm.Recommend(lang),
		"presets":        a.llm.Presets(),
	})
}

// GET /api/llm/models?provider=anthropic|openrouter|ollama[&url=...]
// Returns the selectable model list for a provider. For ollama, with no url it
// scans the local default endpoint; with a url it lists that server's models.
func (a *App) handleProviderModels(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch p := r.URL.Query().Get("provider"); p {
	case "anthropic":
		models := llm.AnthropicModels()
		writeJSON(rw, map[string]any{"models": models, "suggested": llm.SuggestedModels(p, models)})
	case "openrouter":
		models, err := llm.OpenRouterModels(ctx)
		writeJSON(rw, map[string]any{"models": models, "suggested": llm.SuggestedModels(p, models), "error": errStr(err)})
	case "ollama":
		if url := r.URL.Query().Get("url"); url != "" {
			models, err := llm.OllamaModels(ctx, url)
			writeJSON(rw, map[string]any{"models": models, "suggested": llm.SuggestedModels(p, models), "url": url, "found": err == nil && len(models) > 0, "error": errStr(err)})
			return
		}
		url, models, found := llm.OllamaScan(ctx)
		writeJSON(rw, map[string]any{"models": models, "suggested": llm.SuggestedModels(p, models), "url": url, "found": found})
	default:
		writeErr(rw, http.StatusBadRequest, "unknown provider")
	}
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// GET/POST /api/llm/config
func (a *App) handleLLMConfig(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			Provider string `json:"provider"`
			Model    string `json:"model"`
			BaseURL  string `json:"baseURL"`
			APIKey   string `json:"apiKey"`
			Preset   string `json:"preset"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(rw, http.StatusBadRequest, err.Error())
			return
		}
		if body.Preset != "" {
			if err := a.llm.ApplyPreset(body.Preset); err != nil {
				writeErr(rw, http.StatusBadRequest, err.Error())
				return
			}
		} else {
			cfg := a.llm.Config()
			cfg.Provider = body.Provider
			cfg.Model = body.Model
			cfg.BaseURL = body.BaseURL
			cfg.APIKey = body.APIKey // empty leaves existing key in place (see SetConfig)
			cfg.Preset = ""
			a.llm.SetConfig(cfg)
		}
	}
	writeJSON(rw, a.llm.Status()) // never includes the raw key
}

// POST /api/translate {locale, scope, paths?, model?}
func (a *App) handleTranslate(rw http.ResponseWriter, r *http.Request) {
	var req translateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	id, err := a.startJob(req)
	if err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(rw, map[string]string{"jobId": id})
}

// POST /api/locale/add {code, label, lang, dir}
func (a *App) handleAddLocale(rw http.ResponseWriter, r *http.Request) {
	var body struct {
		Code, Label, Lang, Dir string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	if body.Code == "" {
		writeErr(rw, http.StatusBadRequest, "code required")
		return
	}
	if err := a.be.AddLocale(body.Code, backend.AddLocaleOpts{Label: body.Label, Lang: body.Lang, Dir: body.Dir}); err != nil {
		writeErr(rw, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(rw, map[string]string{"status": "added"})
}

// preview handlers
func (a *App) handlePreviewStart(rw http.ResponseWriter, r *http.Request) {
	loc, _ := a.locale(r.URL.Query().Get("locale"))
	url, err := a.preview.start(loc)
	if err != nil {
		writeErr(rw, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(rw, map[string]string{"url": url})
}

func (a *App) handlePreviewStop(rw http.ResponseWriter, r *http.Request) {
	a.preview.stop()
	writeJSON(rw, map[string]string{"status": "stopped"})
}

func (a *App) handlePreviewEvents(rw http.ResponseWriter, r *http.Request) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		writeErr(rw, http.StatusInternalServerError, "SSE unsupported")
		return
	}
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	ch := a.preview.subscribe()
	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-ch:
			data, _ := json.Marshal(ev)
			fmt.Fprintf(rw, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// GET /api/i18n/<locale>  -> UI message catalog
func (a *App) handleUICatalog(rw http.ResponseWriter, r *http.Request) {
	loc := strings.TrimPrefix(r.URL.Path, "/api/i18n/")
	if loc == "" {
		loc = "en"
	}
	data, err := localesFS.ReadFile("frontend/locales/" + loc + ".json")
	if err != nil {
		// fall back to en
		data, err = localesFS.ReadFile("frontend/locales/en.json")
		if err != nil {
			writeErr(rw, http.StatusNotFound, "no catalog")
			return
		}
	}
	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write(data)
}

// GET /api/uilocales -> list of available UI locales
func (a *App) handleUILocales(rw http.ResponseWriter, r *http.Request) {
	entries, _ := localesFS.ReadDir("frontend/locales")
	var out []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".json") {
			out = append(out, strings.TrimSuffix(name, ".json"))
		}
	}
	writeJSON(rw, out)
}
