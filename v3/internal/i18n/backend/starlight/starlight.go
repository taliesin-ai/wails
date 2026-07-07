// Package starlight implements backend.DocsBackend for an Astro Starlight docs
// site (the current Wails docs at docs/). It is the only concrete backend; all
// Starlight-specific knowledge - content layout, locale declaration, MDX
// frontmatter, the /<locale>/ link rule, the npm run dev preview - lives here.
//
// See SPEC.md section 7.
package starlight

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
)

// Starlight is the Astro Starlight adapter, bound to a docs root.
type Starlight struct {
	root        string // absolute docs root (the dir containing astro.config.mjs)
	contentRoot string // root/src/content/docs
}

// configFile is the marker file used for detection and locale parsing.
const configFile = "astro.config.mjs"

// New constructs an adapter bound to the given docs root.
func New(root string) *Starlight {
	abs, _ := filepath.Abs(root)
	return &Starlight{
		root:        abs,
		contentRoot: filepath.Join(abs, "src", "content", "docs"),
	}
}

// Detect probes whether dir is a Starlight docs root. It checks for
// astro.config.mjs and an @astrojs/starlight reference (in the config or
// package.json). Returns the confidence and the resolved root.
func Detect(dir string) (bool, backend.Confidence) {
	cfg := filepath.Join(dir, configFile)
	data, err := os.ReadFile(cfg)
	if err != nil {
		return false, backend.NoConfidence
	}
	if strings.Contains(string(data), "@astrojs/starlight") || strings.Contains(string(data), "starlight(") {
		return true, backend.High
	}
	// astro present but starlight reference not found inline; check package.json.
	if pkg, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		if strings.Contains(string(pkg), "@astrojs/starlight") {
			return true, backend.High
		}
	}
	return true, backend.Low
}

func (s *Starlight) Name() string              { return "starlight" }
func (s *Starlight) Root() string              { return s.root }
func (s *Starlight) SourceLocale() string      { return "en" }
func (s *Starlight) Format() backend.DocFormat { return backend.FormatMDX }

// ContentRoot exposes the docs content directory (used by the catalog/cache for
// snapshot and cache-dir resolution by the caller).
func (s *Starlight) ContentRoot() string { return s.contentRoot }

// ListLocales parses the locales:{...} block from astro.config.mjs. The "root"
// key is the source locale (English).
func (s *Starlight) ListLocales() ([]backend.Locale, error) {
	data, err := os.ReadFile(filepath.Join(s.root, configFile))
	if err != nil {
		return nil, err
	}
	return parseLocales(string(data)), nil
}

var (
	localesBlockRe = regexp.MustCompile(`(?s)locales:\s*\{(.*?)\n\s*\},`)
	localeLineRe   = regexp.MustCompile(`(?m)^\s*(?:"([^"]+)"|([A-Za-z][\w-]*))\s*:\s*\{\s*label:\s*"([^"]*)",\s*lang:\s*"([^"]*)",\s*dir:\s*"([^"]*)"`)
)

func parseLocales(config string) []backend.Locale {
	block := config
	if m := localesBlockRe.FindStringSubmatch(config); m != nil {
		block = m[1]
	}
	var out []backend.Locale
	for _, m := range localeLineRe.FindAllStringSubmatch(block, -1) {
		key := m[1]
		if key == "" {
			key = m[2]
		}
		loc := backend.Locale{
			Label: m[3], Lang: m[4], Dir: m[5],
		}
		if key == "root" {
			loc.Code = "en"
			loc.IsSource = true
		} else {
			loc.Code = key
		}
		out = append(out, loc)
	}
	return out
}

// localeDirs returns the set of on-disk subdirectory names that hold
// translations, so ListSourceDocs can exclude them.
func (s *Starlight) localeDirs() map[string]bool {
	dirs := map[string]bool{}
	locs, _ := s.ListLocales()
	for _, l := range locs {
		if !l.IsSource {
			dirs[l.Code] = true
		}
	}
	return dirs
}

// ignoreDirs are directories under the content root that are never translatable
// source docs.
var ignoreDirs = map[string]bool{
	"blog": true, // Starlight blog posts are author-specific, not translated docs
}

// ListSourceDocs walks the content root and returns every .md/.mdx file that is
// not inside a locale subdirectory or an ignored directory.
func (s *Starlight) ListSourceDocs() ([]backend.SourceRef, error) {
	localeDirs := s.localeDirs()
	var out []backend.SourceRef
	err := filepath.Walk(s.contentRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".mdx" {
			return nil
		}
		rel, err := filepath.Rel(s.contentRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		top := strings.SplitN(rel, "/", 2)[0]
		if localeDirs[top] || ignoreDirs[top] {
			return nil
		}
		out = append(out, backend.SourceRef{Path: rel})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Starlight) SourcePath(src backend.SourceRef) string {
	return filepath.Join(s.contentRoot, filepath.FromSlash(src.Path))
}

func (s *Starlight) TargetPath(loc backend.Locale, src backend.SourceRef) string {
	if loc.IsSource {
		return s.SourcePath(src)
	}
	return filepath.Join(s.contentRoot, loc.Code, filepath.FromSlash(src.Path))
}

func (s *Starlight) ReadDoc(ref backend.DocRef) (backend.Doc, error) {
	path := s.TargetPath(ref.Locale, ref.Source)
	data, err := os.ReadFile(path)
	if err != nil {
		return backend.Doc{}, err
	}
	return SplitDoc(string(data)), nil
}

func (s *Starlight) WriteDoc(ref backend.DocRef, doc backend.Doc) error {
	path := s.TargetPath(ref.Locale, ref.Source)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(joinDoc(doc)), 0o644)
}

// SplitDoc separates YAML frontmatter (between the first two "---" lines) from
// the body. Raw is always the full content. Exported because callers writing
// externally-produced content (the MCP write path) must split it before link
// rewriting so frontmatter is never rewritten.
func SplitDoc(raw string) backend.Doc {
	d := backend.Doc{Raw: raw}
	if !strings.HasPrefix(raw, "---\n") && !strings.HasPrefix(raw, "---\r\n") {
		d.Body = raw
		return d
	}
	rest := raw[strings.IndexByte(raw, '\n')+1:]
	end := regexp.MustCompile(`\r?\n---\r?\n`).FindStringIndex(rest)
	if end == nil {
		d.Body = raw
		return d
	}
	d.Frontmatter = rest[:end[0]]
	d.Body = rest[end[1]:]
	return d
}

// joinDoc reassembles a doc. If a doc has frontmatter, write the fences back.
func joinDoc(d backend.Doc) string {
	if d.Frontmatter == "" {
		if d.Raw != "" && d.Body == "" {
			return d.Raw
		}
		return d.Body
	}
	return "---\n" + d.Frontmatter + "\n---\n" + d.Body
}

// internalLinkRe matches markdown links/images whose target is a site-internal
// absolute path ("/..."), capturing the leading "](" or "](" forms.
var internalLinkRe = regexp.MustCompile(`(\]\()(/[^)\s]*)`)

// RewriteInternalLinks prefixes site-internal absolute links with /<locale> so
// a translated page links to the translated target (the i18n(id) link-prefix
// convention). External links, anchors and already-prefixed links are left
// alone.
func (s *Starlight) RewriteInternalLinks(doc backend.Doc, loc backend.Locale) backend.Doc {
	if loc.IsSource {
		return doc
	}
	prefix := "/" + loc.Code
	rewrite := func(body string) string {
		return internalLinkRe.ReplaceAllStringFunc(body, func(m string) string {
			sm := internalLinkRe.FindStringSubmatch(m)
			lead, target := sm[1], sm[2]
			// Already locale-prefixed?
			if strings.HasPrefix(target, prefix+"/") || target == prefix {
				return m
			}
			return lead + prefix + target
		})
	}
	doc.Body = rewrite(doc.Body)
	if doc.Raw != "" {
		doc.Raw = joinDoc(doc)
	}
	return doc
}

// AddLocale appends a locale entry to the locales:{...} block in astro.config.mjs.
// Best-effort textual edit; returns an error if the block cannot be located.
func (s *Starlight) AddLocale(code string, opts backend.AddLocaleOpts) error {
	cfgPath := filepath.Join(s.root, configFile)
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}
	config := string(data)
	loc := localesBlockRe.FindStringSubmatchIndex(config)
	if loc == nil {
		return fmt.Errorf("could not find locales block in %s; add the locale manually", configFile)
	}
	dir := opts.Dir
	if dir == "" {
		dir = "ltr"
	}
	entry := fmt.Sprintf("        %q: { label: %q, lang: %q, dir: %q },\n", code, opts.Label, opts.Lang, dir)
	// Insert just before the block's closing "\n  },".
	insertAt := loc[3] // end of the captured inner block (before "\n  },")
	updated := config[:insertAt] + "\n" + strings.TrimRight(entry, "\n") + config[insertAt:]
	return os.WriteFile(cfgPath, []byte(updated), 0o644)
}

// PreviewCommand returns the Astro dev-server command for live preview.
func (s *Starlight) PreviewCommand(_ backend.Locale) (backend.Command, bool) {
	return backend.Command{Name: "npm", Args: []string{"run", "dev"}, Dir: s.root}, true
}

// compile-time assertion that the adapter satisfies the seam.
var _ backend.DocsBackend = (*Starlight)(nil)
