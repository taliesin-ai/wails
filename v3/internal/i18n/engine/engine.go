// Package engine performs MDX-safe translation. Structure preservation is a
// first-class concern: frontmatter is split and only whitelisted keys are
// translated; fenced/inline code and MDX/JSX are masked before the LLM sees them
// and restored after; glossary terms are pinned; internal links are rewritten
// per-locale by the backend; and the result is round-trip validated against the
// source structure. The same protect/restore runs for manual and machine paths.
//
// See SPEC.md section 8.
package engine

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/cache"
	"github.com/wailsapp/wails/v3/internal/i18n/llm"
)

// Engine orchestrates translation against a backend and cache directory.
type Engine struct {
	be       backend.DocsBackend
	cacheDir string
	glossary []string
}

// New constructs an Engine.
func New(be backend.DocsBackend, cacheDir string, glossary []string) *Engine {
	return &Engine{be: be, cacheDir: cacheDir, glossary: glossary}
}

// Glossary returns the configured do-not-translate terms.
func (e *Engine) Glossary() []string { return e.glossary }

// --- Structure protection ---

var (
	fencedCodeRe = regexp.MustCompile("(?s)```.*?```")
	inlineCodeRe = regexp.MustCompile("`[^`\n]+`")
	importRe     = regexp.MustCompile(`(?m)^import\s+.*$`)
	// JSX/MDX tags: opening, closing, or self-closing components. Heuristic.
	jsxTagRe = regexp.MustCompile(`</?[A-Za-z][\w.]*(?:\s[^>]*)?/?>`)
)

const phToken = "[[PH:%d]]"

var phFindRe = regexp.MustCompile(`\[\[PH:(\d+)\]\]`)

// Protect masks code fences, inline code, import lines and JSX tags, returning
// the masked text and the ordered list of original fragments.
func Protect(text string) (string, []string) {
	var tokens []string
	mask := func(re *regexp.Regexp, s string) string {
		return re.ReplaceAllStringFunc(s, func(m string) string {
			tok := fmt.Sprintf(phToken, len(tokens))
			tokens = append(tokens, m)
			return tok
		})
	}
	// Order matters: fenced code first (may contain backticks/tags), then imports,
	// then JSX tags, then inline code.
	out := mask(fencedCodeRe, text)
	out = mask(importRe, out)
	out = mask(jsxTagRe, out)
	out = mask(inlineCodeRe, out)
	return out, tokens
}

// Restore puts the original fragments back in place of their tokens.
func Restore(text string, tokens []string) string {
	return phFindRe.ReplaceAllStringFunc(text, func(m string) string {
		var i int
		fmt.Sscanf(m, phToken, &i)
		if i >= 0 && i < len(tokens) {
			return tokens[i]
		}
		return m
	})
}

// --- Chunk-on-overflow ---

// maxChunkChars bounds how much masked text we send in one LLM call. It is sized
// well under the providers' output ceilings (~10k tokens at ~4 chars/token, so a
// translation that expands ~2x still fits comfortably). Most documents are
// smaller than this and translate whole - which keeps the model's context intact
// for terminology and flow. Only genuinely large files are split, on blank-line
// boundaries, trading some cross-paragraph coherence for not being truncated.
// This runs in the engine so it protects every provider, streaming or not.
const maxChunkChars = 40000

var blankLineRe = regexp.MustCompile(`\n{2,}`)

// translateMasked translates already-masked text. If it fits in one call it is
// sent whole; otherwise it is split on blank-line boundaries (safe because
// Protect collapses fenced code and JSX to single inline tokens) and the parts
// are translated independently and rejoined. The returned text is still masked -
// callers Restore once over the whole result so placeholder indices stay global.
func translateMasked(ctx context.Context, prov llm.Provider, masked string, base llm.TranslateInput) (string, error) {
	if len(masked) <= maxChunkChars {
		base.Text = masked
		return prov.Translate(ctx, base)
	}
	chunks := splitChunks(masked, maxChunkChars)
	parts := make([]string, len(chunks))
	for i, c := range chunks {
		in := base
		in.Text = c
		out, err := prov.Translate(ctx, in)
		if err != nil {
			return "", err
		}
		parts[i] = out
	}
	return strings.Join(parts, "\n\n"), nil
}

// splitChunks greedily packs blank-line-separated paragraphs into chunks no
// larger than budget. A single paragraph that already exceeds budget becomes its
// own (oversized) chunk rather than being split mid-sentence.
func splitChunks(s string, budget int) []string {
	paras := blankLineRe.Split(s, -1)
	var chunks []string
	var cur strings.Builder
	for _, p := range paras {
		if cur.Len() > 0 && cur.Len()+len(p)+2 > budget {
			chunks = append(chunks, cur.String())
			cur.Reset()
		}
		if cur.Len() > 0 {
			cur.WriteString("\n\n")
		}
		cur.WriteString(p)
	}
	if cur.Len() > 0 {
		chunks = append(chunks, cur.String())
	}
	return chunks
}

// --- Blocks (presentation only) ---

// Block is an on-the-fly Markdown segment for the side-by-side view. It is never
// the persisted unit of translation.
type Block struct {
	Kind string `json:"kind"` // heading | code | paragraph | list | jsx
	Text string `json:"text"`
}

// Segment splits a document body into display blocks. Fenced code stays whole;
// other content is grouped by blank lines.
func Segment(body string) []Block {
	lines := strings.Split(body, "\n")
	var blocks []Block
	var cur []string
	inFence := false

	flush := func() {
		if len(cur) == 0 {
			return
		}
		text := strings.Join(cur, "\n")
		blocks = append(blocks, Block{Kind: classify(text), Text: text})
		cur = nil
	}

	for _, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "```") {
			if !inFence {
				flush()
				inFence = true
				cur = append(cur, ln)
			} else {
				cur = append(cur, ln)
				inFence = false
				flush()
			}
			continue
		}
		if inFence {
			cur = append(cur, ln)
			continue
		}
		if trimmed == "" {
			flush()
			continue
		}
		cur = append(cur, ln)
	}
	flush()
	return blocks
}

func classify(text string) string {
	t := strings.TrimSpace(text)
	switch {
	case strings.HasPrefix(t, "```"):
		return "code"
	case strings.HasPrefix(t, "#"):
		return "heading"
	case strings.HasPrefix(t, "<"):
		return "jsx"
	case strings.HasPrefix(t, "- "), strings.HasPrefix(t, "* "), regexp.MustCompile(`^\d+\.\s`).MatchString(t):
		return "list"
	default:
		return "paragraph"
	}
}

// --- Diffs ---

// DiffRow is one aligned source/target (or old/new source) pair.
type DiffRow struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	Kind    string `json:"kind"`
	Changed bool   `json:"changed"`
}

// Diff is a block-aligned diff for the review screen.
type Diff struct {
	Rows []DiffRow `json:"rows"`
}

// AlignDiff aligns two block lists by index (sufficient for the file-level model).
func AlignDiff(left, right []Block) Diff {
	n := len(left)
	if len(right) > n {
		n = len(right)
	}
	d := Diff{}
	for i := 0; i < n; i++ {
		var l, r Block
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		d.Rows = append(d.Rows, DiffRow{
			Source:  l.Text,
			Target:  r.Text,
			Kind:    firstNonEmpty(l.Kind, r.Kind),
			Changed: l.Text != r.Text,
		})
	}
	return d
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// --- Frontmatter (whitelisted) ---

var fmKeyRe = regexp.MustCompile(`(?m)^(\s*)(title|description):\s*(.+)$`)

// translatableFrontmatter applies fn to whitelisted frontmatter values only,
// leaving keys/IDs/slugs untouched.
func translateFrontmatter(fmText string, fn func(string) string) string {
	return fmKeyRe.ReplaceAllStringFunc(fmText, func(m string) string {
		sm := fmKeyRe.FindStringSubmatch(m)
		indent, key, val := sm[1], sm[2], sm[3]
		unquoted := strings.Trim(val, `"'`)
		translated := fn(unquoted)
		if val != unquoted { // was quoted
			return fmt.Sprintf(`%s%s: "%s"`, indent, key, strings.ReplaceAll(translated, `"`, `\"`))
		}
		return fmt.Sprintf("%s%s: %s", indent, key, translated)
	})
}

// --- Validation ---

// ValidateStructure checks that key structural counts match between source and
// translated bodies. Returned warnings are surfaced in the review screen.
func ValidateStructure(srcBody, outBody string) []string {
	var warns []string
	if a, b := countMatches(fencedCodeRe, srcBody), countMatches(fencedCodeRe, outBody); a != b {
		warns = append(warns, fmt.Sprintf("code fence count changed (%d -> %d)", a, b))
	}
	if a, b := countHeadings(srcBody), countHeadings(outBody); a != b {
		warns = append(warns, fmt.Sprintf("heading count changed (%d -> %d)", a, b))
	}
	if a, b := countMatches(jsxTagRe, srcBody), countMatches(jsxTagRe, outBody); a != b {
		warns = append(warns, fmt.Sprintf("JSX tag count changed (%d -> %d)", a, b))
	}
	return append(warns, validateStepsStructure(srcBody, outBody)...)
}

func countMatches(re *regexp.Regexp, s string) int { return len(re.FindAllString(s, -1)) }

// CountCodeFences counts complete fenced code blocks with the same measure
// ValidateStructure uses, so callers gating a write on fence parity (the MCP
// write tool treats a mismatch as truncated output) agree with the warnings.
func CountCodeFences(s string) int { return countMatches(fencedCodeRe, s) }
func countHeadings(s string) int {
	return len(regexp.MustCompile(`(?m)^#{1,6}\s`).FindAllString(s, -1))
}

// --- Translation pipeline ---

// FileResult is the outcome of translating a single file.
type FileResult struct {
	Path     string   `json:"path"`
	Warnings []string `json:"warnings,omitempty"`
	Err      string   `json:"error,omitempty"`
}

// TranslateFile runs the full protect/translate/restore/validate/write pipeline
// for one file and stamps the cache + source snapshot.
func (e *Engine) TranslateFile(ctx context.Context, prov llm.Provider, loc backend.Locale, src backend.SourceRef, model string) FileResult {
	res := FileResult{Path: src.Path}

	srcDoc, err := e.be.ReadDoc(backend.DocRef{Locale: backend.Locale{IsSource: true}, Source: src})
	if err != nil {
		res.Err = err.Error()
		return res
	}

	base := llm.TranslateInput{
		SourceLang: e.be.SourceLocale(), TargetLang: loc.Lang,
		TargetLabel: loc.Label, Glossary: e.glossary, Model: model,
	}
	tr := func(text string) string {
		if strings.TrimSpace(text) == "" {
			return text
		}
		masked, tokens := Protect(text)
		out, terr := translateMasked(ctx, prov, masked, base)
		if terr != nil {
			res.Err = terr.Error()
			return text
		}
		return Restore(out, tokens)
	}

	outFM := translateFrontmatter(srcDoc.Frontmatter, tr)
	outBody := tr(srcDoc.Body)
	if res.Err != "" {
		return res
	}

	res.Warnings = ValidateStructure(srcDoc.Body, outBody)

	outDoc := backend.Doc{Frontmatter: outFM, Body: outBody}
	outDoc = e.be.RewriteInternalLinks(outDoc, loc)

	if err := e.be.WriteDoc(backend.DocRef{Locale: loc, Source: src}, outDoc); err != nil {
		res.Err = err.Error()
		return res
	}

	// Stamp cache + snapshot so staleness and source-drift diffs work later.
	if err := e.Stamp(loc, src, []byte(srcDoc.Raw)); err != nil {
		res.Warnings = append(res.Warnings, "cache stamp failed: "+err.Error())
	}
	return res
}

// SourceDrift returns the body of the source as it was at translation time
// (from the snapshot) and the current source body, for a source-vs-source diff
// on stale files (SPEC.md section 8). Errors if no snapshot exists.
func (e *Engine) SourceDrift(loc backend.Locale, src backend.SourceRef) (oldBody, curBody string, err error) {
	c, err := cache.Load(e.cacheDir, loc.Code)
	if err != nil {
		return "", "", err
	}
	entry, ok := c[src.Path]
	if !ok || entry.SourceSnapshot == "" {
		return "", "", fmt.Errorf("no source snapshot for %s", src.Path)
	}
	oldRaw, err := cache.ReadSnapshot(e.cacheDir, entry.SourceSnapshot)
	if err != nil {
		return "", "", err
	}
	curDoc, err := e.be.ReadDoc(backend.DocRef{Locale: backend.Locale{IsSource: true}, Source: src})
	if err != nil {
		return "", "", err
	}
	return stripFrontmatter(string(oldRaw)), curDoc.Body, nil
}

// stripFrontmatter removes a leading YAML frontmatter block for diffing bodies.
func stripFrontmatter(raw string) string {
	if !strings.HasPrefix(raw, "---\n") && !strings.HasPrefix(raw, "---\r\n") {
		return raw
	}
	rest := raw[strings.IndexByte(raw, '\n')+1:]
	if loc := regexp.MustCompile(`\r?\n---\r?\n`).FindStringIndex(rest); loc != nil {
		return rest[loc[1]:]
	}
	return raw
}

// Stamp records the source hash + snapshot for a translated file in the cache.
func (e *Engine) Stamp(loc backend.Locale, src backend.SourceRef, sourceRaw []byte) error {
	c, err := cache.Load(e.cacheDir, loc.Code)
	if err != nil {
		return err
	}
	if err := cache.WriteSnapshot(e.cacheDir, src.Path, sourceRaw); err != nil {
		return err
	}
	c[src.Path] = cache.Entry{
		Hash:           cache.Hash(sourceRaw),
		SourceSnapshot: cache.SnapshotPath(src.Path),
		TranslatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	return c.Save(e.cacheDir, loc.Code)
}
