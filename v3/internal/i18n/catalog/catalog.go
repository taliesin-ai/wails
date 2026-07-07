// Package catalog cross-references the backend's current source list against
// each locale's committed cache to derive freshness state per file and per
// locale. Freshness is *derived* on every launch (SPEC.md section 5.6); it is
// distinct from the stored review status (machine / needs_review / approved),
// which is layered on top by the review package.
package catalog

import (
	"os"
	"sort"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/cache"
)

// Freshness is the derived staleness state of a single translated file.
type Freshness string

const (
	UpToDate Freshness = "up_to_date"
	Stale    Freshness = "stale"
	Missing  Freshness = "missing"
	Orphaned Freshness = "orphaned"
)

// FileStatus is the per-file freshness plus the review fields (filled by the
// review package when available).
type FileStatus struct {
	Path         string    `json:"path"`
	Freshness    Freshness `json:"freshness"`
	HasSnapshot  bool      `json:"hasSnapshot"`            // a source snapshot exists -> a source diff is possible
	ReviewStatus string    `json:"reviewStatus,omitempty"` // untranslated|machine|needs_review|approved
	Viewed       bool      `json:"viewed"`
	TranslatedAt string    `json:"translatedAt,omitempty"`
}

// LocaleStatus is the dashboard summary for one locale.
type LocaleStatus struct {
	Code         string `json:"code"`
	Label        string `json:"label"`
	Lang         string `json:"lang"`
	Dir          string `json:"dir"`
	NativeName   string `json:"nativeName,omitempty"`
	Total        int    `json:"total"`
	UpToDate     int    `json:"upToDate"`
	Stale        int    `json:"stale"`
	Missing      int    `json:"missing"`
	Orphaned     int    `json:"orphaned"`
	Approved     int    `json:"approved"`
	Machine      int    `json:"machine"`
	FreshnessPct int    `json:"freshnessPct"`
}

// ReviewLookup lets the catalog fold personal review state into its output
// without importing the review package (avoids an import cycle and keeps review
// state optional). Implemented by review.Store.
type ReviewLookup interface {
	Status(locale, path string) (status string, viewed bool, ok bool)
}

// Catalog computes freshness against a backend and its cache directory.
type Catalog struct {
	be       backend.DocsBackend
	cacheDir string
	review   ReviewLookup // may be nil
}

// New creates a Catalog. review may be nil.
func New(be backend.DocsBackend, cacheDir string, review ReviewLookup) *Catalog {
	return &Catalog{be: be, cacheDir: cacheDir, review: review}
}

// sourceHash returns the current hash of a source document, or "" if it cannot
// be read.
func (c *Catalog) sourceHash(src backend.SourceRef) string {
	data, err := os.ReadFile(c.be.SourcePath(src))
	if err != nil {
		return ""
	}
	return cache.Hash(data)
}

// FileStatuses derives the freshness of every file for a locale, including
// orphans (cache/target entries whose source is gone).
func (c *Catalog) FileStatuses(loc backend.Locale) ([]FileStatus, error) {
	cacheData, err := cache.Load(c.cacheDir, loc.Code)
	if err != nil {
		return nil, err
	}
	srcs, err := c.be.ListSourceDocs()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, len(srcs))
	out := make([]FileStatus, 0, len(srcs))

	for _, src := range srcs {
		seen[src.Path] = true
		fs := FileStatus{Path: src.Path, ReviewStatus: "untranslated"}
		entry, inCache := cacheData[src.Path]

		// A translation file is considered present if a target file exists OR a
		// cache entry exists.
		targetExists := fileExists(c.be.TargetPath(loc, src))

		switch {
		case !inCache && !targetExists:
			fs.Freshness = Missing
		case !inCache && targetExists:
			// Translated by hand without a cache stamp -> treat as up to date but
			// unverified; surface as stale so it gets a hash stamp on next save.
			fs.Freshness = Stale
		default:
			fs.HasSnapshot = entry.SourceSnapshot != ""
			fs.TranslatedAt = entry.TranslatedAt
			if entry.Hash != "" && entry.Hash == c.sourceHash(src) {
				fs.Freshness = UpToDate
			} else {
				fs.Freshness = Stale
			}
		}

		if c.review != nil {
			if status, viewed, ok := c.review.Status(loc.Code, src.Path); ok {
				fs.ReviewStatus = status
				fs.Viewed = viewed
			}
		}
		out = append(out, fs)
	}

	// Orphans: cache entries whose source no longer exists.
	for path, entry := range cacheData {
		if seen[path] {
			continue
		}
		fs := FileStatus{
			Path:         path,
			Freshness:    Orphaned,
			HasSnapshot:  entry.SourceSnapshot != "",
			TranslatedAt: entry.TranslatedAt,
			ReviewStatus: "needs_review",
		}
		if c.review != nil {
			if status, viewed, ok := c.review.Status(loc.Code, path); ok {
				fs.ReviewStatus = status
				fs.Viewed = viewed
			}
		}
		out = append(out, fs)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// LocaleStatus computes the dashboard summary for one locale.
func (c *Catalog) LocaleStatus(loc backend.Locale) (LocaleStatus, error) {
	files, err := c.FileStatuses(loc)
	if err != nil {
		return LocaleStatus{}, err
	}
	ls := LocaleStatus{
		Code: loc.Code, Label: loc.Label, Lang: loc.Lang, Dir: loc.Dir, NativeName: loc.NativeName,
	}
	for _, f := range files {
		ls.Total++
		switch f.Freshness {
		case UpToDate:
			ls.UpToDate++
		case Stale:
			ls.Stale++
		case Missing:
			ls.Missing++
		case Orphaned:
			ls.Orphaned++
		}
		switch f.ReviewStatus {
		case "approved":
			ls.Approved++
		case "machine":
			ls.Machine++
		}
	}
	// Orphans are not part of the translatable total for the freshness %.
	denom := ls.Total - ls.Orphaned
	if denom > 0 {
		ls.FreshnessPct = ls.UpToDate * 100 / denom
	}
	return ls, nil
}

// LocaleStatuses returns dashboard summaries for every declared non-source locale.
func (c *Catalog) LocaleStatuses() ([]LocaleStatus, error) {
	locales, err := c.be.ListLocales()
	if err != nil {
		return nil, err
	}
	out := make([]LocaleStatus, 0, len(locales))
	for _, loc := range locales {
		if loc.IsSource {
			continue
		}
		ls, err := c.LocaleStatus(loc)
		if err != nil {
			return nil, err
		}
		out = append(out, ls)
	}
	return out, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
