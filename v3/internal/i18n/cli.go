package i18n

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/backend/starlight"
	"github.com/wailsapp/wails/v3/internal/i18n/catalog"
	"github.com/wailsapp/wails/v3/internal/i18n/engine"
	"github.com/wailsapp/wails/v3/internal/i18n/review"
)

type HelperMode string

const (
	HelperModeStatus    HelperMode = "status"
	HelperModeListStale HelperMode = "list-stale"
	HelperModeRecord    HelperMode = "record"
)

type HelperOptions struct {
	Mode   HelperMode
	Dir    string
	Locale string
	File   string
	All    bool
	Stdout io.Writer
}

type helperSession struct {
	be       backend.DocsBackend
	cacheDir string
	cat      *catalog.Catalog
	eng      *engine.Engine
	rev      *review.Store
}

func RunHelper(opts HelperOptions) error {
	out := opts.Stdout
	if out == nil {
		out = os.Stdout
	}
	s, err := newHelperSession(opts.Dir)
	if err != nil {
		return err
	}
	switch opts.Mode {
	case HelperModeStatus:
		return s.printStatus(opts, out)
	case HelperModeListStale:
		return s.printStale(opts, out)
	case HelperModeRecord:
		return s.record(opts, out)
	default:
		return fmt.Errorf("unknown translate helper mode %q", opts.Mode)
	}
}

func newHelperSession(dir string) (*helperSession, error) {
	root, _ := findDocsRoot(dir)
	if root == "" {
		return nil, fmt.Errorf("could not find a docs site (Astro Starlight). Run from your project or pass --dir <docs root>")
	}
	be := starlight.New(root)
	cacheDir := filepath.Join(be.Root(), ".translation-cache")
	rev := review.NewStore(cacheDir)
	return &helperSession{
		be:       be,
		cacheDir: cacheDir,
		cat:      catalog.New(be, cacheDir, rev),
		eng:      engine.New(be, cacheDir, loadGlossary(be.Root())),
		rev:      rev,
	}, nil
}

func (s *helperSession) printStatus(opts HelperOptions, out io.Writer) error {
	locs, err := s.resolveLocales(opts, true)
	if err != nil {
		return err
	}
	for _, loc := range locs {
		status, err := s.cat.LocaleStatus(loc)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s (%s): up-to-date=%d stale=%d missing=%d orphaned=%d / %d\n",
			status.Code, status.Label, status.UpToDate, status.Stale, status.Missing, status.Orphaned, status.Total)
	}
	return nil
}

func (s *helperSession) printStale(opts HelperOptions, out io.Writer) error {
	locs, err := s.resolveLocales(opts, false)
	if err != nil {
		return err
	}
	prefixLocale := opts.All || len(locs) > 1
	for _, loc := range locs {
		files, err := s.cat.FileStatuses(loc)
		if err != nil {
			return err
		}
		for _, file := range files {
			if file.Freshness == catalog.Stale || file.Freshness == catalog.Missing {
				if prefixLocale {
					fmt.Fprintf(out, "%s\t%s\n", loc.Code, file.Path)
					continue
				}
				fmt.Fprintln(out, file.Path)
			}
		}
	}
	return nil
}

func (s *helperSession) record(opts HelperOptions, out io.Writer) error {
	if opts.All {
		return fmt.Errorf("--record needs exactly one --locale <code>")
	}
	src, err := backend.ParseSourceRef(opts.File)
	if err != nil {
		return err
	}
	loc, err := s.locale(opts.Locale)
	if err != nil {
		return err
	}
	srcDoc, err := s.be.ReadDoc(backend.DocRef{Locale: backend.Locale{IsSource: true}, Source: src})
	if err != nil {
		return fmt.Errorf("source page not found: %s: %w", src.Path, err)
	}
	if _, err := s.be.ReadDoc(backend.DocRef{Locale: loc, Source: src}); err != nil {
		return fmt.Errorf("translated file not found (write it first): %s/%s: %w", loc.Code, src.Path, err)
	}
	if err := s.eng.Stamp(loc, src, []byte(srcDoc.Raw)); err != nil {
		return err
	}
	fmt.Fprintf(out, "recorded %s/%s\n", loc.Code, src.Path)
	return nil
}

func (s *helperSession) resolveLocales(opts HelperOptions, defaultAll bool) ([]backend.Locale, error) {
	if opts.All || opts.Locale == "" && defaultAll {
		return s.nonSourceLocales()
	}
	loc, err := s.locale(opts.Locale)
	if err != nil {
		return nil, err
	}
	return []backend.Locale{loc}, nil
}

func (s *helperSession) nonSourceLocales() ([]backend.Locale, error) {
	locs, err := s.be.ListLocales()
	if err != nil {
		return nil, err
	}
	out := make([]backend.Locale, 0, len(locs))
	for _, loc := range locs {
		if !loc.IsSource {
			out = append(out, loc)
		}
	}
	return out, nil
}

func (s *helperSession) locale(code string) (backend.Locale, error) {
	if code == "" {
		return backend.Locale{}, fmt.Errorf("specify --locale <code>")
	}
	locs, err := s.be.ListLocales()
	if err != nil {
		return backend.Locale{}, err
	}
	for _, loc := range locs {
		if loc.Code == code && !loc.IsSource {
			return loc, nil
		}
	}
	return backend.Locale{}, fmt.Errorf("unknown locale %q", code)
}
