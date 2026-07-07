# `wails3 translate` — implementation

A Crowdin-style documentation translation workbench, implemented from
[`SPEC.md`](./SPEC.md). Run `wails3 translate` from anywhere inside a project
with an Astro Starlight docs site; it auto-detects the docs root, starts a local
HTTP server, and opens your browser on the Digital Wails–themed dashboard.

## Running

```bash
# From the repo (uses the local build, not a released wails3):
go run ./v3/cmd/wails3 translate
go run ./v3/cmd/wails3 translate --dir docs
go run ./v3/cmd/wails3 translate --locale de

# Headless (CI bridge): machine-translate stale+missing files for a locale.
ANTHROPIC_API_KEY=... go run ./v3/cmd/wails3 translate --headless --locale de

# Script-friendly helper modes.
go run ./v3/cmd/wails3 translate --dir docs --status --all
go run ./v3/cmd/wails3 translate --dir docs --list-stale --all
go run ./v3/cmd/wails3 translate --dir docs --record --locale de --file quick-start/why-wails.mdx
```

## Contributor workflow and cache acceptance

Public-facing contributor docs live in the docs site, but the operational model
is intentionally simple:

1. Run `wails3 translate --dir docs --locale <code>` from a Wails checkout, or
   `wails3 translate --clone` to create a persistent docs workspace under the
   user's cache directory.
2. Use the dashboard to find stale and missing files. `--status --all` provides a
   terminal summary, and `--list-stale --all` prints script-safe
   `locale<TAB>path` rows.
3. Translate or update the target MDX. For stale files with snapshots, the
   **Source changes** view shows old English vs current English so translators
   update the changed parts instead of rereading the whole page.
4. Save in the workbench, or run `--record --locale <code> --file <path>` after a
   manual edit. `<path>` is a source-relative path under `docs/src/content/docs/`;
   absolute paths, `..`, backslashes, drive names, and cleaned rewrites are
   rejected at the CLI and HTTP API boundaries.
5. Commit translated MDX, `docs/.translation-cache/<locale>.json`, and any
   `docs/.translation-cache/.snapshots/` files written for source-diff review.
   Do not commit `docs/.translation-cache/.review/`; it is local review state.

Historical cache entries whose hash recipe cannot be reproduced are not silently
migrated. They remain stale until a human reviews the current English and target
translation, then saves or records that file. That explicit acceptance is the
migration path and produces a cache entry with the current reproducible hash and,
when available, a source snapshot for future diffs.

## Architecture (matches SPEC.md §3)

```
commands.Translate ──► i18n.App (app.go: HTTP server, mirrors setupwizard)
  backend/            DocsBackend seam (the portability boundary)
    starlight/        the only concrete adapter
  cache/              .translation-cache read/write + hash recipe
  catalog/            freshness derivation (up_to_date/stale/missing/orphaned)
  engine/             MDX-safe protect/restore, blocks, diffs, translate pipeline
  llm/                provider abstraction + model recommendation map
  review/             personal, gitignored review state
  frontend/dist/      embedded UI (go:embed)
  frontend/locales/   UI message catalogs (en, id) served over localhost
```

Everything except `backend/` is backend-agnostic. Adding a second docs backend is
one new adapter implementing `backend.DocsBackend` (proven by the fake-backend
seam test in `catalog/catalog_test.go`).

## What's implemented

All seven SPEC build phases are present and the tool runs end to end:

- **Dashboard** — locale cards with freshness bars, counts, freshness %, the
  "English has moved on" banner, and quick actions (translate missing / update
  stale). Data from real `docs/.translation-cache`.
- **Workspace** — file tree with freshness badges + viewed/approved marks, a
  "Needs attention" filter, an editor with Raw / Side-by-side / Preview modes,
  manual edit + save (re-stamps the hash, sets review status).
- **Engine** — frontmatter split (whitelisted keys), code/JSX/import masking and
  restore, glossary pinning, per-locale link rewriting, round-trip structure
  validation, source snapshots for the source-vs-source stale diff. Unit-tested.
- **LLM** — Anthropic (default `claude-opus-4-8`), OpenAI-compatible, Ollama; a
  config-driven per-language recommendation map (`llm/models.json`). Keys are
  session-only.
- **Review** — viewed/approve toggles, machine/needs_review/approved status,
  provenance; state is personal and gitignored.
- **Live preview** — drives `npm run dev`, captures the dev-server URL, streams
  status over SSE, deep-links the iframe to the locale page.
- **Submit** — manual git/gh commands always shown; automated fork→PR via `gh`.
- **Secrets** — export/load credentials with the hard in-repo save guard.
- **Headless** — `--headless` reuses the same engine/backend/llm with no window.
- **UI i18n** — every string from a served catalog; `en` + `id`; language
  switcher; RTL-ready.

## Documented deviations from SPEC.md

These are intentional, pragmatic choices; the Go seams are unaffected so they can
be revised later without a rewrite.

1. **HTTP server + browser, not an embedded Wails window** (§3/§15 say "Wails
   desktop window mirroring setupwizard"). `setupwizard` is itself an HTTP-server
   /browser app, and the user explicitly wanted a browser. Live preview is an
   `<iframe>` rather than a webview as a result.
2. **Vanilla no-build frontend** instead of Vite + React + Tailwind (§15) and a
   **`<textarea>` editor** instead of CodeMirror 6 (§11). This keeps the tool
   buildable and runnable with no npm step; the `frontend/dist` is committed and
   embedded directly. Swapping in the real toolchain later only touches
   `frontend/` — no Go change.
3. **Hash recipe** (§4). The original docs-translate helper's recipe could not be
   reproduced from the tree (it lives on an unmerged branch), so `cache/` defines
   its own reproducible recipe: `sha256(sourceBytes)[:12]`. The tool reads and
   stamps with this single recipe, so everything it produces is internally
   consistent; pre-existing entries simply read as "stale" (truthful, since the
   English sources have changed since they were written).
4. **API surface is REST/SSE** rather than Wails-bound methods (follows from
   deviation 1). The method contract in §6 maps one-to-one onto the `/api/*`
   routes.
5. **Source-diff review screen.** The Go endpoint `/api/sourcediff` (old-source vs
   current source for stale files, §8) is wired in the UI so translators can see
   exactly what changed in the original text when updating a stale translation.
```
