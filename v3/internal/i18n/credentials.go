package i18n

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/internal/i18n/llm"
)

// credentialsFile is the on-disk shape of an exported LLM credential file.
type credentialsFile struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"baseURL,omitempty"`
	APIKey   string `json:"apiKey,omitempty"`
}

// defaultCredentialsPath is a per-user config location OUTSIDE any repo.
func defaultCredentialsPath() string {
	if cfg, err := os.UserConfigDir(); err == nil {
		return filepath.Join(cfg, "wails3", "translate-credentials.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".wails3-translate-credentials.json")
}

// repoRoot walks up from start looking for a repo marker (.git or go.work).
func repoRoot(start string) string {
	dir := start
	for {
		for _, marker := range []string{".git", "go.work"} {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// guardOutsideRepo returns an error if target resolves to a path inside the
// repository checkout (so credentials can never be committed). This is the hard
// guard required by SPEC.md section 10.
func (a *App) guardOutsideRepo(target string) error {
	abs, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	abs = filepath.Clean(abs)

	roots := map[string]bool{}
	if r := repoRoot(a.be.Root()); r != "" {
		roots[r] = true
	}
	// The docs root itself and its parent also count (covers nested worktrees).
	roots[a.be.Root()] = true

	for root := range roots {
		rootAbs, _ := filepath.Abs(root)
		rootAbs = filepath.Clean(rootAbs)
		rel, err := filepath.Rel(rootAbs, abs)
		if err != nil {
			continue
		}
		if rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." && !filepath.IsAbs(rel)) {
			return fmt.Errorf("refusing to save credentials inside the repository (%s) - they could be committed. Choose a location outside %s", abs, rootAbs)
		}
	}
	return nil
}

// POST /api/llm/credentials/export {path}
func (a *App) handleExportCredentials(rw http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Path == "" {
		body.Path = defaultCredentialsPath()
	}
	if err := a.guardOutsideRepo(body.Path); err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	cfg := a.llm.Config()
	out := credentialsFile{Provider: cfg.Provider, Model: cfg.Model, BaseURL: cfg.BaseURL, APIKey: cfg.APIKey}
	data, _ := json.MarshalIndent(out, "", "  ")
	if err := os.MkdirAll(filepath.Dir(body.Path), 0o700); err != nil {
		writeErr(rw, http.StatusInternalServerError, err.Error())
		return
	}
	if err := os.WriteFile(body.Path, data, 0o600); err != nil {
		writeErr(rw, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(rw, map[string]string{"status": "saved", "path": body.Path})
}

// POST /api/llm/credentials/load {path}
func (a *App) handleLoadCredentials(rw http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Path == "" {
		body.Path = defaultCredentialsPath()
	}
	data, err := os.ReadFile(body.Path)
	if err != nil {
		writeErr(rw, http.StatusNotFound, err.Error())
		return
	}
	var in credentialsFile
	if err := json.Unmarshal(data, &in); err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	a.llm.SetConfig(llm.Config{Provider: in.Provider, Model: in.Model, BaseURL: in.BaseURL, APIKey: in.APIKey})
	writeJSON(rw, a.llm.Status())
}

// defaultCredentialsPathHandler exposes the default path so the UI can prefill it.
func (a *App) handleCredentialsDefault(rw http.ResponseWriter, r *http.Request) {
	writeJSON(rw, map[string]string{"path": defaultCredentialsPath()})
}
