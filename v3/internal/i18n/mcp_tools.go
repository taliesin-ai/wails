package i18n

import (
	"fmt"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/backend/starlight"
	"github.com/wailsapp/wails/v3/internal/i18n/catalog"
	"github.com/wailsapp/wails/v3/internal/i18n/engine"
	"github.com/wailsapp/wails/v3/internal/i18n/review"
)

// --- schema helpers ---

func mcpObjectSchema(props map[string]any, required ...string) map[string]any {
	schema := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func mcpProp(typ, desc string) map[string]any {
	return map[string]any{"type": typ, "description": desc}
}

func mcpEnumProp(desc string, values ...string) map[string]any {
	return map[string]any{"type": "string", "enum": values, "description": desc}
}

// --- argument helpers ---

func mcpStringArg(args map[string]any, key string) string {
	s, _ := args[key].(string)
	return s
}

func mcpRequiredString(args map[string]any, key string) (string, error) {
	s := mcpStringArg(args, key)
	if s == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return s, nil
}

func mcpBoolArg(args map[string]any, key string) bool {
	b, _ := args[key].(bool)
	return b
}

func (s *mcpServer) registerTools() {
	s.tools = []*mcpTool{
		{
			Name: "setup_docs",
			Description: "Resolve the docs site to translate. Uses the docs site detected at startup " +
				"when there is one; otherwise reuses (or sparse-clones) the Wails documentation into a " +
				"local cache and translates that. Call this first if other tools report no docs site. " +
				"Returns the docs root and the declared locales.",
			Schema:  mcpObjectSchema(map[string]any{}),
			Handler: s.toolSetupDocs,
		},
		{
			Name: "translation_status",
			Description: "Per-locale translation coverage: total, up-to-date, stale, missing and orphaned " +
				"file counts plus review progress. Omit locale for all non-source locales. Use this to " +
				"pick which locale to work on.",
			Schema: mcpObjectSchema(map[string]any{
				"locale": mcpProp("string", "Locale code, e.g. \"de\". Omit for all non-source locales."),
			}),
			Handler: s.toolTranslationStatus,
		},
		{
			Name: "list_files",
			Description: "List source files for a locale with their freshness. The default filter " +
				"'needs_work' (stale + missing) is the work queue: these are the files to translate.",
			Schema: mcpObjectSchema(map[string]any{
				"locale": mcpProp("string", "Locale code, e.g. \"de\"."),
				"filter": mcpEnumProp("Which files to list; defaults to needs_work (stale + missing).",
					"needs_work", "stale", "missing", "orphaned", "up_to_date", "all"),
			}, "locale"),
			Handler: s.toolListFiles,
		},
		{
			Name: "get_translation_context",
			Description: "Everything needed to translate one file: the raw source, the existing " +
				"translation (null when never translated), the glossary of terms that must stay " +
				"untranslated, and - for stale files - drift: the source blocks that changed since the " +
				"file was last translated, so only those need updating. Translate the source yourself, " +
				"then submit the complete file with write_translation.",
			Schema: mcpObjectSchema(map[string]any{
				"locale": mcpProp("string", "Locale code, e.g. \"de\"."),
				"path":   mcpProp("string", "Source-relative file path, e.g. \"guides/routing.mdx\"."),
				"include": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string", "enum": []string{"source", "target", "drift", "glossary"}},
					"description": "Restrict which fields are returned (all by default); useful for very large files.",
				},
			}, "locale", "path"),
			Handler: s.toolGetTranslationContext,
		},
		{
			Name: "write_translation",
			Description: "Write a complete translated file (including frontmatter). Never translate code " +
				"blocks, inline code, import lines, JSX/MDX tags, glossary terms, or frontmatter keys " +
				"(only title and description values). Internal links are locale-prefixed server-side - " +
				"do not rewrite them yourself. The write is structure-validated, cache-stamped, and " +
				"queued as needs_review for human review. A code-fence count mismatch (usually truncated " +
				"output) is rejected: fix the content and retry, or pass force to write anyway.",
			Schema: mcpObjectSchema(map[string]any{
				"locale":  mcpProp("string", "Locale code, e.g. \"de\"."),
				"path":    mcpProp("string", "Source-relative file path, e.g. \"guides/routing.mdx\"."),
				"content": mcpProp("string", "The complete translated file, frontmatter included."),
				"model":   mcpProp("string", "Your model name, recorded as the translation's provenance."),
				"force":   mcpProp("boolean", "Write even when the code-fence count differs from the source."),
			}, "locale", "path", "content"),
			Handler: s.toolWriteTranslation,
		},
		{
			Name: "add_locale",
			Description: "Register a new locale in the docs site config so it can be translated. " +
				"Label and lang default to the code; dir defaults to ltr.",
			Schema: mcpObjectSchema(map[string]any{
				"code":  mcpProp("string", "Locale code, e.g. \"pt-br\"."),
				"label": mcpProp("string", "Human label, e.g. \"Português do Brasil\"."),
				"lang":  mcpProp("string", "BCP-47 tag; defaults to the code."),
				"dir":   mcpEnumProp("Text direction; defaults to ltr.", "ltr", "rtl"),
			}, "code"),
			Handler: s.toolAddLocale,
		},
	}
}

// toolSetupDocs resolves a docs root, cloning the Wails docs as a last resort.
// It never prompts (MCP mode has no TTY semantics) and sends clone progress to
// stderr. Safe to call when a session already exists.
func (s *mcpServer) toolSetupDocs(map[string]any) (any, error) {
	cloned := false
	if s.session == nil {
		sess, err := newHelperSession(s.dir)
		switch {
		case err == nil:
			// A docs site appeared since startup (e.g. the user cd'd the harness).
		case s.dir != "":
			// An explicit --dir that doesn't resolve is honoured, never cloned over.
			return nil, err
		default:
			root, _, didClone, cloneErr := cloneWailsDocs(s.logw)
			if cloneErr != nil {
				return nil, cloneErr
			}
			cloned = didClone
			if sess, err = newHelperSession(root); err != nil {
				return nil, err
			}
		}
		s.session = sess
	}
	locales, err := s.session.be.ListLocales()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"docsRoot": s.session.be.Root(),
		"backend":  s.session.be.Name(),
		"cloned":   cloned,
		"locales":  locales,
	}, nil
}

func (s *mcpServer) toolTranslationStatus(args map[string]any) (any, error) {
	sess, err := s.sess()
	if err != nil {
		return nil, err
	}
	code := mcpStringArg(args, "locale")
	if code == "" {
		return sess.cat.LocaleStatuses()
	}
	loc, err := sess.locale(code)
	if err != nil {
		return nil, err
	}
	return sess.cat.LocaleStatus(loc)
}

func (s *mcpServer) toolListFiles(args map[string]any) (any, error) {
	sess, err := s.sess()
	if err != nil {
		return nil, err
	}
	code, err := mcpRequiredString(args, "locale")
	if err != nil {
		return nil, err
	}
	loc, err := sess.locale(code)
	if err != nil {
		return nil, err
	}
	filter := mcpStringArg(args, "filter")
	if filter == "" {
		filter = "needs_work"
	}
	files, err := sess.cat.FileStatuses(loc)
	if err != nil {
		return nil, err
	}
	keep := func(f catalog.FileStatus) bool {
		switch filter {
		case "needs_work":
			return f.Freshness == catalog.Stale || f.Freshness == catalog.Missing
		case "all":
			return true
		default:
			return string(f.Freshness) == filter
		}
	}
	out := make([]catalog.FileStatus, 0, len(files))
	for _, f := range files {
		if keep(f) {
			out = append(out, f)
		}
	}
	return out, nil
}

// mcpDriftBlock is one source block that changed since the file was translated.
type mcpDriftBlock struct {
	Kind string `json:"kind"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

type mcpDrift struct {
	ChangedBlocks []mcpDriftBlock `json:"changedBlocks"`
	TotalBlocks   int             `json:"totalBlocks"`
	ChangedCount  int             `json:"changedCount"`
}

type mcpTranslationContext struct {
	Path         string            `json:"path"`
	Locale       backend.Locale    `json:"locale"`
	Freshness    catalog.Freshness `json:"freshness"`
	ReviewStatus string            `json:"reviewStatus,omitempty"`
	Source       *string           `json:"source,omitempty"`
	Target       *string           `json:"target"` // null when never translated
	Glossary     []string          `json:"glossary,omitempty"`
	Drift        *mcpDrift         `json:"drift"` // null unless stale with a snapshot
}

func (s *mcpServer) toolGetTranslationContext(args map[string]any) (any, error) {
	sess, err := s.sess()
	if err != nil {
		return nil, err
	}
	code, err := mcpRequiredString(args, "locale")
	if err != nil {
		return nil, err
	}
	path, err := mcpRequiredString(args, "path")
	if err != nil {
		return nil, err
	}
	loc, err := sess.locale(code)
	if err != nil {
		return nil, err
	}
	src, err := backend.ParseSourceRef(path)
	if err != nil {
		return nil, err
	}

	include := map[string]bool{}
	if rawInc, ok := args["include"].([]any); ok {
		for _, v := range rawInc {
			if fld, ok := v.(string); ok {
				include[fld] = true
			}
		}
	}
	want := func(field string) bool { return len(include) == 0 || include[field] }

	// Freshness + review state for this file (also validates the path is a
	// known source doc or an orphan).
	files, err := sess.cat.FileStatuses(loc)
	if err != nil {
		return nil, err
	}
	var status *catalog.FileStatus
	for i := range files {
		if files[i].Path == src.Path {
			status = &files[i]
			break
		}
	}
	if status == nil {
		return nil, fmt.Errorf("unknown source file %q (see list_files)", src.Path)
	}

	ctx := mcpTranslationContext{Path: src.Path, Locale: loc, Freshness: status.Freshness, ReviewStatus: status.ReviewStatus}

	if want("source") {
		srcDoc, err := sess.be.ReadDoc(backend.DocRef{Locale: backend.Locale{IsSource: true}, Source: src})
		if err != nil {
			return nil, fmt.Errorf("source page not found: %s: %w", src.Path, err)
		}
		ctx.Source = &srcDoc.Raw
	}
	if want("target") {
		if tgtDoc, err := sess.be.ReadDoc(backend.DocRef{Locale: loc, Source: src}); err == nil {
			ctx.Target = &tgtDoc.Raw
		}
	}
	if want("glossary") {
		ctx.Glossary = sess.eng.Glossary()
	}
	if want("drift") && status.Freshness == catalog.Stale && status.HasSnapshot {
		oldBody, curBody, err := sess.eng.SourceDrift(loc, src)
		if err == nil {
			diff := engine.AlignDiff(engine.Segment(oldBody), engine.Segment(curBody))
			drift := mcpDrift{TotalBlocks: len(diff.Rows)}
			for _, row := range diff.Rows {
				if row.Changed {
					drift.ChangedBlocks = append(drift.ChangedBlocks, mcpDriftBlock{Kind: row.Kind, Old: row.Source, New: row.Target})
				}
			}
			drift.ChangedCount = len(drift.ChangedBlocks)
			ctx.Drift = &drift
		}
	}
	return ctx, nil
}

func (s *mcpServer) toolWriteTranslation(args map[string]any) (any, error) {
	sess, err := s.sess()
	if err != nil {
		return nil, err
	}
	code, err := mcpRequiredString(args, "locale")
	if err != nil {
		return nil, err
	}
	path, err := mcpRequiredString(args, "path")
	if err != nil {
		return nil, err
	}
	content, ok := args["content"].(string)
	if !ok || content == "" {
		return nil, fmt.Errorf("content is required (the complete translated file)")
	}
	loc, err := sess.locale(code)
	if err != nil {
		return nil, err
	}
	src, err := backend.ParseSourceRef(path)
	if err != nil {
		return nil, err
	}
	srcDoc, err := sess.be.ReadDoc(backend.DocRef{Locale: backend.Locale{IsSource: true}, Source: src})
	if err != nil {
		return nil, fmt.Errorf("source page not found: %s: %w", src.Path, err)
	}

	// Mirror engine.TranslateFile semantics: split frontmatter before link
	// rewriting so it is never rewritten, validate, rewrite links server-side.
	doc := starlight.SplitDoc(content)
	warnings := engine.ValidateStructure(srcDoc.Body, doc.Body)

	// A fence-count mismatch almost always means truncated output; reject so
	// the agent can retry cleanly. Everything else is write-and-warn.
	if a, b := engine.CountCodeFences(srcDoc.Body), engine.CountCodeFences(doc.Body); a != b && !mcpBoolArg(args, "force") {
		return nil, fmt.Errorf(
			"code fence count differs (source %d, translation %d) - the content may be truncated; fix and resubmit the complete file, or pass force:true to write anyway", a, b)
	}

	outDoc := sess.be.RewriteInternalLinks(backend.Doc{Frontmatter: doc.Frontmatter, Body: doc.Body}, loc)
	if err := sess.be.WriteDoc(backend.DocRef{Locale: loc, Source: src}, outDoc); err != nil {
		return nil, err
	}
	if err := sess.eng.Stamp(loc, src, []byte(srcDoc.Raw)); err != nil {
		warnings = append(warnings, "cache stamp failed: "+err.Error())
	}
	if err := sess.rev.SetStatus(loc.Code, src.Path, "needs_review", review.Provenance{Method: "mcp", Model: mcpStringArg(args, "model")}); err != nil {
		warnings = append(warnings, "review state not recorded: "+err.Error())
	}
	if warnings == nil {
		warnings = []string{}
	}
	return map[string]any{
		"written":      sess.be.TargetPath(loc, src),
		"freshness":    catalog.UpToDate,
		"reviewStatus": "needs_review",
		"warnings":     warnings,
	}, nil
}

func (s *mcpServer) toolAddLocale(args map[string]any) (any, error) {
	sess, err := s.sess()
	if err != nil {
		return nil, err
	}
	code, err := mcpRequiredString(args, "code")
	if err != nil {
		return nil, err
	}
	opts := backend.AddLocaleOpts{
		Label: mcpStringArg(args, "label"),
		Lang:  mcpStringArg(args, "lang"),
		Dir:   mcpStringArg(args, "dir"),
	}
	if opts.Label == "" {
		opts.Label = code
	}
	if opts.Lang == "" {
		opts.Lang = code
	}
	if opts.Dir == "" {
		opts.Dir = "ltr"
	}
	if err := sess.be.AddLocale(code, opts); err != nil {
		return nil, err
	}
	return map[string]string{"status": "added", "code": code}, nil
}
