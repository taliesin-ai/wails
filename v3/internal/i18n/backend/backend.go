// Package backend defines the DocsBackend seam: the single interface that
// isolates everything which knows *where docs live and how they are shaped*.
// Everything else in the i18n subsystem (catalog, cache, engine, llm, review,
// the HTTP API and the frontend) depends only on this package, never on a
// concrete docs generator. Adding a future docs backend is therefore an
// additive change - write one adapter implementing DocsBackend - not a rewrite.
//
// See v3/internal/i18n/SPEC.md section 7 for the design rationale.
package backend

// Confidence expresses how sure a backend is that it owns a given docs root.
type Confidence int

const (
	NoConfidence Confidence = iota
	Low
	Medium
	High
)

func (c Confidence) String() string {
	switch c {
	case High:
		return "high"
	case Medium:
		return "medium"
	case Low:
		return "low"
	default:
		return "none"
	}
}

// Locale is a declared documentation locale.
type Locale struct {
	Code       string `json:"code"`       // directory/key code, e.g. "de", "zh-cn"
	Label      string `json:"label"`      // human label as declared by the site, e.g. "Deutsch"
	Lang       string `json:"lang"`       // BCP-47 tag, e.g. "de", "zh-CN"
	Dir        string `json:"dir"`        // "ltr" | "rtl"
	NativeName string `json:"nativeName"` // optional native name
	IsSource   bool   `json:"isSource"`   // true for the source locale (English/root)
}

// SourceRef identifies a translatable source document by its path relative to
// the source content root, e.g. "quick-start/why-wails.mdx".
type SourceRef struct {
	Path string `json:"path"`
}

// DocRef points at a concrete document on disk: a source document when
// Locale.IsSource is true, otherwise the translation of Source for Locale.
type DocRef struct {
	Locale Locale
	Source SourceRef
}

// DocFormat describes what structure-preservation the engine must apply.
type DocFormat string

const (
	FormatMD  DocFormat = "md"
	FormatMDX DocFormat = "mdx"
)

// Doc is a loaded document with its YAML frontmatter split out from the body.
// Raw is the exact bytes on disk; Frontmatter excludes the surrounding "---"
// fences and is empty when the document has none.
type Doc struct {
	Frontmatter string `json:"frontmatter"`
	Body        string `json:"body"`
	Raw         string `json:"raw"`
}

// Command is an external command a backend asks the tool to run (e.g. the live
// preview dev server).
type Command struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
	Dir  string   `json:"dir"`
}

// AddLocaleOpts carries the metadata needed to wire a new locale into a site.
type AddLocaleOpts struct {
	Label string `json:"label"`
	Lang  string `json:"lang"`
	Dir   string `json:"dir"`
}

// DocsBackend is the portability seam. The Starlight adapter is the only
// concrete implementation today; see SPEC.md section 7.
type DocsBackend interface {
	// Identity & location.
	Name() string         // e.g. "starlight"
	Root() string         // absolute docs root this adapter is bound to
	SourceLocale() string // e.g. "en" (or the root locale code)

	// Content layout.
	ListLocales() ([]Locale, error)              // declared locales + labels/native names
	ListSourceDocs() ([]SourceRef, error)        // translatable files (honours ignore globs)
	SourcePath(src SourceRef) string             // absolute path of the source doc
	TargetPath(loc Locale, src SourceRef) string // absolute path of the translation
	ReadDoc(ref DocRef) (Doc, error)             // load source/target, frontmatter split out
	WriteDoc(ref DocRef, doc Doc) error          // persist a translation

	// Conventions the engine must respect.
	Format() DocFormat                               // md | mdx
	RewriteInternalLinks(doc Doc, loc Locale) Doc    // per-locale link prefixing
	AddLocale(code string, opts AddLocaleOpts) error // wire a new locale into the site config

	// Build / preview.
	PreviewCommand(loc Locale) (Command, bool) // dev-server command, ok=false if unavailable
}
