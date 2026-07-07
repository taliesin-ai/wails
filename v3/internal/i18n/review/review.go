// Package review holds the personal, never-committed review state: which files
// the user has viewed/approved and how each translation was produced. It lives
// in a local, gitignored file (docs/.translation-cache/.review/<locale>.json)
// and is the user's private progress tracker - there is no team/shared mode.
//
// See SPEC.md sections 4, 10, 17.
package review

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Provenance records how a translation was produced.
type Provenance struct {
	Method string `json:"method"`          // manual | llm | mcp
	Model  string `json:"model,omitempty"` // model id when machine-produced (llm | mcp)
}

// Record is the per-file review state.
type Record struct {
	Status     string     `json:"status"` // untranslated | machine | needs_review | approved
	Viewed     bool       `json:"viewed"`
	Provenance Provenance `json:"provenance,omitempty"`
}

// reviewSubdir is the gitignored directory under the cache dir.
const reviewSubdir = ".review"

// Store is a concurrency-safe accessor for per-locale review files. Reads hit
// disk on every call - the files are tiny - so concurrent processes (the
// workbench and an MCP server on the same docs root) always see each other's
// writes; a process-lifetime memo here would clobber the other side's records
// on the next persist. The mutex serializes in-process access.
type Store struct {
	cacheDir string
	mu       sync.Mutex
}

// NewStore creates a review store rooted at the translation cache directory.
func NewStore(cacheDir string) *Store {
	return &Store{cacheDir: cacheDir}
}

func (s *Store) dir() string { return filepath.Join(s.cacheDir, reviewSubdir) }
func (s *Store) file(loc string) string {
	return filepath.Join(s.dir(), loc+".json")
}

// load reads a locale's review file fresh from disk (caller holds mu).
func (s *Store) load(loc string) map[string]Record {
	m := map[string]Record{}
	if data, err := os.ReadFile(s.file(loc)); err == nil {
		_ = json.Unmarshal(data, &m)
	}
	return m
}

// persist writes a locale's review file and guarantees it is gitignored.
func (s *Store) persist(loc string, m map[string]Record) error {
	if err := os.MkdirAll(s.dir(), 0o755); err != nil {
		return err
	}
	if err := s.ensureGitignored(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file(loc), data, 0o644)
}

// ensureGitignored makes sure the review directory can never be committed by
// dropping a self-ignoring .gitignore in it (matches the SPEC.md guarantee that
// review state is never a shared artifact).
func (s *Store) ensureGitignored() error {
	gi := filepath.Join(s.dir(), ".gitignore")
	if _, err := os.Stat(gi); err == nil {
		return nil
	}
	return os.WriteFile(gi, []byte("# Personal translation review state - never commit.\n*\n"), 0o644)
}

// Status implements catalog.ReviewLookup.
func (s *Store) Status(loc, path string) (status string, viewed bool, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, found := s.load(loc)[path]
	if !found {
		return "", false, false
	}
	return rec.Status, rec.Viewed, true
}

// Get returns the full record for a file (zero value if absent).
func (s *Store) Get(loc, path string) Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load(loc)[path]
}

// SetViewed records that the user has looked at a file.
func (s *Store) SetViewed(loc, path string, viewed bool) error {
	return s.update(loc, path, func(r *Record) { r.Viewed = viewed })
}

// SetApproved toggles approval; approving also marks the file viewed.
func (s *Store) SetApproved(loc, path string, approved bool) error {
	return s.update(loc, path, func(r *Record) {
		if approved {
			r.Status = "approved"
			r.Viewed = true
		} else if r.Status == "approved" {
			r.Status = "needs_review"
		}
	})
}

// SetStatus sets the review status (and provenance for machine output).
func (s *Store) SetStatus(loc, path, status string, prov Provenance) error {
	return s.update(loc, path, func(r *Record) {
		r.Status = status
		if prov.Method != "" {
			r.Provenance = prov
		}
	})
}

func (s *Store) update(loc, path string, fn func(*Record)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.load(loc)
	rec := m[path]
	fn(&rec)
	m[path] = rec
	return s.persist(loc, m)
}

// IsReviewPath reports whether p is inside the review dir (used by guards).
func (s *Store) IsReviewPath(p string) bool {
	abs, _ := filepath.Abs(p)
	return strings.HasPrefix(abs, s.dir())
}
