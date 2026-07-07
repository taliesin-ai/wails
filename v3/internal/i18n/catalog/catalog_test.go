package catalog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/cache"
)

// fakeBackend is a minimal in-memory/on-disk DocsBackend used to prove the seam:
// the catalog computes freshness correctly against any adapter, not just Starlight.
type fakeBackend struct {
	root    string
	sources []string
	locales []backend.Locale
}

func (f *fakeBackend) Name() string                           { return "fake" }
func (f *fakeBackend) Root() string                           { return f.root }
func (f *fakeBackend) SourceLocale() string                   { return "en" }
func (f *fakeBackend) Format() backend.DocFormat              { return backend.FormatMD }
func (f *fakeBackend) ListLocales() ([]backend.Locale, error) { return f.locales, nil }
func (f *fakeBackend) ListSourceDocs() ([]backend.SourceRef, error) {
	var out []backend.SourceRef
	for _, s := range f.sources {
		out = append(out, backend.SourceRef{Path: s})
	}
	return out, nil
}
func (f *fakeBackend) SourcePath(s backend.SourceRef) string {
	return filepath.Join(f.root, "en", s.Path)
}
func (f *fakeBackend) TargetPath(l backend.Locale, s backend.SourceRef) string {
	return filepath.Join(f.root, l.Code, s.Path)
}
func (f *fakeBackend) ReadDoc(r backend.DocRef) (backend.Doc, error)                    { return backend.Doc{}, nil }
func (f *fakeBackend) WriteDoc(backend.DocRef, backend.Doc) error                       { return nil }
func (f *fakeBackend) RewriteInternalLinks(d backend.Doc, _ backend.Locale) backend.Doc { return d }
func (f *fakeBackend) AddLocale(string, backend.AddLocaleOpts) error                    { return nil }
func (f *fakeBackend) PreviewCommand(backend.Locale) (backend.Command, bool) {
	return backend.Command{}, false
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFreshnessDerivation(t *testing.T) {
	root := t.TempDir()
	be := &fakeBackend{
		root:    root,
		sources: []string{"a.md", "b.md", "c.md"},
		locales: []backend.Locale{{Code: "en", IsSource: true}, {Code: "de", Label: "Deutsch"}},
	}
	// Sources.
	write(t, be.SourcePath(backend.SourceRef{Path: "a.md"}), "alpha v2") // changed since translation
	write(t, be.SourcePath(backend.SourceRef{Path: "b.md"}), "beta")     // up to date
	write(t, be.SourcePath(backend.SourceRef{Path: "c.md"}), "gamma")    // never translated -> missing

	cacheDir := filepath.Join(root, ".cache")
	c := cache.Cache{
		"a.md": {Hash: cache.Hash([]byte("alpha v1"))}, // stale (old hash)
		"b.md": {Hash: cache.Hash([]byte("beta"))},     // up to date
		"d.md": {Hash: cache.Hash([]byte("deleted"))},  // orphaned (no source)
	}
	if err := c.Save(cacheDir, "de"); err != nil {
		t.Fatal(err)
	}
	// Target files for a.md and b.md (so they aren't treated as missing).
	write(t, be.TargetPath(backend.Locale{Code: "de"}, backend.SourceRef{Path: "a.md"}), "alpha de")
	write(t, be.TargetPath(backend.Locale{Code: "de"}, backend.SourceRef{Path: "b.md"}), "beta de")

	cat := New(be, cacheDir, nil)
	files, err := cat.FileStatuses(backend.Locale{Code: "de", Label: "Deutsch"})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Freshness{}
	for _, f := range files {
		got[f.Path] = f.Freshness
	}
	checks := map[string]Freshness{"a.md": Stale, "b.md": UpToDate, "c.md": Missing, "d.md": Orphaned}
	for path, want := range checks {
		if got[path] != want {
			t.Errorf("%s: freshness = %q, want %q", path, got[path], want)
		}
	}

	ls, err := cat.LocaleStatus(backend.Locale{Code: "de", Label: "Deutsch"})
	if err != nil {
		t.Fatal(err)
	}
	if ls.UpToDate != 1 || ls.Stale != 1 || ls.Missing != 1 || ls.Orphaned != 1 {
		t.Errorf("counts wrong: %+v", ls)
	}
	// freshness % = up-to-date / (total - orphaned) = 1/3 = 33
	if ls.FreshnessPct != 33 {
		t.Errorf("freshnessPct = %d, want 33", ls.FreshnessPct)
	}
}

var _ backend.DocsBackend = (*fakeBackend)(nil)
