package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/catalog"
	"github.com/wailsapp/wails/v3/internal/i18n/review"
)

// translateReq starts a machine-translation job.
type translateReq struct {
	Locale string   `json:"locale"`
	Scope  string   `json:"scope"` // all | missing | stale | paths
	Paths  []string `json:"paths"`
	Model  string   `json:"model"`
}

// jobEvent is one progress event streamed to the UI.
type jobEvent struct {
	Type     string   `json:"type"` // start | file_start | file_done | error | done
	JobID    string   `json:"jobId"`
	Path     string   `json:"path,omitempty"`
	Index    int      `json:"index,omitempty"`
	Total    int      `json:"total,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Error    string   `json:"error,omitempty"`
}

type translateJob struct {
	id     string
	mu     sync.Mutex
	log    []jobEvent
	subs   []chan jobEvent
	closed bool
}

func (j *translateJob) emit(ev jobEvent) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.log = append(j.log, ev)
	for _, ch := range j.subs {
		select {
		case ch <- ev:
		default:
		}
	}
	if ev.Type == "done" {
		j.closed = true
		for _, ch := range j.subs {
			close(ch)
		}
		j.subs = nil
	}
}

func (j *translateJob) subscribe() chan jobEvent {
	j.mu.Lock()
	defer j.mu.Unlock()
	ch := make(chan jobEvent, 64)
	for _, ev := range j.log { // replay history
		ch <- ev
	}
	if j.closed {
		close(ch)
		return ch
	}
	j.subs = append(j.subs, ch)
	return ch
}

// startJob resolves the file set for the scope, builds the provider, and runs
// the translation in the background, emitting progress events.
func (a *App) startJob(req translateReq) (string, error) {
	loc, ok := a.locale(req.Locale)
	if !ok {
		return "", fmt.Errorf("unknown locale %q", req.Locale)
	}
	var todo []backend.SourceRef
	switch req.Scope {
	case "paths":
		for _, p := range req.Paths {
			src, err := backend.ParseSourceRef(p)
			if err != nil {
				return "", err
			}
			todo = append(todo, src)
		}
	case "missing":
		files, err := a.cat.FileStatuses(loc)
		if err != nil {
			return "", err
		}
		for _, f := range files {
			if f.Freshness == catalog.Missing {
				todo = append(todo, backend.SourceRef{Path: f.Path})
			}
		}
	case "stale":
		files, err := a.cat.FileStatuses(loc)
		if err != nil {
			return "", err
		}
		for _, f := range files {
			if f.Freshness == catalog.Stale {
				todo = append(todo, backend.SourceRef{Path: f.Path})
			}
		}
	default: // all = missing + stale
		files, err := a.cat.FileStatuses(loc)
		if err != nil {
			return "", err
		}
		for _, f := range files {
			if f.Freshness == catalog.Missing || f.Freshness == catalog.Stale {
				todo = append(todo, backend.SourceRef{Path: f.Path})
			}
		}
	}

	prov, err := a.llm.Provider()
	if err != nil {
		return "", err
	}

	model := req.Model
	if model == "" {
		model = a.llm.Config().Model
	}

	job := &translateJob{id: fmt.Sprintf("job-%d", time.Now().UnixNano())}
	a.jobsMu.Lock()
	if a.jobs == nil {
		a.jobs = map[string]*translateJob{}
	}
	a.jobs[job.id] = job
	a.jobsMu.Unlock()

	go func() {
		ctx := context.Background()
		job.emit(jobEvent{Type: "start", JobID: job.id, Total: len(todo)})
		for i, src := range todo {
			job.emit(jobEvent{Type: "file_start", JobID: job.id, Path: src.Path, Index: i + 1, Total: len(todo)})
			res := a.eng.TranslateFile(ctx, prov, loc, src, model)
			if res.Err != "" {
				job.emit(jobEvent{Type: "error", JobID: job.id, Path: src.Path, Index: i + 1, Total: len(todo), Error: res.Err})
				continue
			}
			_ = a.rev.SetStatus(loc.Code, src.Path, "machine", review.Provenance{Method: "llm", Model: model})
			job.emit(jobEvent{Type: "file_done", JobID: job.id, Path: src.Path, Index: i + 1, Total: len(todo), Warnings: res.Warnings})
		}
		job.emit(jobEvent{Type: "done", JobID: job.id, Total: len(todo)})
	}()

	return job.id, nil
}

// GET /api/translate/events?jobId=...
func (a *App) handleTranslateEvents(rw http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("jobId")
	a.jobsMu.Lock()
	job := a.jobs[id]
	a.jobsMu.Unlock()
	if job == nil {
		writeErr(rw, http.StatusNotFound, "unknown job")
		return
	}
	flusher, ok := rw.(http.Flusher)
	if !ok {
		writeErr(rw, http.StatusInternalServerError, "SSE unsupported")
		return
	}
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	ch := job.subscribe()
	for {
		select {
		case <-r.Context().Done():
			return
		case ev, open := <-ch:
			if !open {
				return
			}
			data, _ := json.Marshal(ev)
			fmt.Fprintf(rw, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
