# `wails3 translate` — Specification

A Crowdin-style translation workbench, shipped inside the `wails3` binary. Run `wails3 translate`, get a desktop window with the Digital Wails branding, and manage documentation translations: see what's translated, edit existing locales, create new ones manually or with an LLM, review LLM output diff-by-diff, and approve files GitHub-style before they land.

Internal tool first, but built to be opened to the community. Built modular — the docs site sits behind one interface — so it can be re-pointed to a different docs generator in the future without a rewrite. (Keeping that door open is a design consideration, not a v1 feature: we don't target or detect any specific second generator today.)

> **Naming.** The subsystem lives at **`v3/internal/i18n/`** — it's the home for all internationalization (docs translation now; the tool's own UI strings per §14; and, later, app runtime strings). The CLI verb stays **`wails3 translate`**, the entry point into the i18n subsystem. The bound Go↔JS layer is the **`api/`** package (this replaces the earlier `service/` name).

---

## 1. Goals and non-goals

### Goals
- One command (`wails3 translate`) opens a branded local UI.
- Locale-aware: knows which languages exist, how complete each is, and what's stale.
- Two creation paths: **manual** (Crowdin-style side-by-side editing) and **LLM-assisted** (configure a provider, machine-translate, then review).
- Recommend a good translation model per target language; let the user override.
- A review screen with an embedded editor and GitHub-style per-file "viewed / approved" checkboxes for LLM output.
- Editing an existing locale: open the files you need in the embedded editor and save.
- **Portable** across documentation backends. Starlight today, behind a docs-backend interface so a different generator can be added later as a single new adapter — no rewrite.
- Community-friendly: bring-your-own LLM key (never committed), and a clear path to submit a translation back (export / PR).

### Non-goals (v1)
- Not a general-purpose CAT tool with translation memory servers or vendor integration.
- Not a translator for the **user's application** runtime strings — v1 is **docs only** (the `.md` / `.mdx` corpus). App-string translation is explicitly out of scope; the modular seam would allow it as a future adapter, but it is not a planned mode. (The tool's *own* UI strings are still localized — see §14.)
- **Single-user / personal.** No team-collaboration features. Review/approval state is private to the user (§10); collaboration, if any, happens through git/PRs, not the tool.
- No hosted/multi-user collaboration server. Review state is personal and local (§10). Collaboration, if any, happens through git/PRs.
- No automatic glossary mining. The glossary is a hand-maintained config (it already exists).

---

## 2. The central design decision: "exactly like Crowdin" vs. the existing pipeline

Crowdin's core model is **segment-level**: every source string/paragraph has a target, each segment carries its own status, and translation memory reuses segments. The Wails docs pipeline that already exists is **whole-file**: hash the entire `.mdx`, translate the whole file, stamp the hash in a cache (see §4).

These conflict. The resolution this spec adopts:

> **Hybrid. File-level is the unit of work, staleness, caching, and approval. Block-level is a presentation layer in the review/edit screen only.**

- The **engine** translates and tracks whole files. This matches the existing `.translation-cache/<locale>.json` hash scheme and the existing agent-driven flow, so we stay backward-compatible and don't have to solve segment reassembly.
- The **review/edit UI** parses each file into blocks (headings, paragraphs, code fences, JSX/MDX components, list items) and shows a **side-by-side source↔target** view per block — the Crowdin feel — without making the block the persisted translation unit. Editing a block edits the underlying file region; saving writes the whole file and re-stamps the file hash.

Rejected alternative: true segment-level translation + caching. It gives perfect Crowdin parity and finer reuse, but segment extraction/reassembly over MDX (JSX components spanning blocks, frontmatter, nested fences) is fragile, and it throws away the existing cache format. Revisit only if file-level reuse proves too coarse in practice.

**Implication for the data model:** block boundaries are computed on the fly from the Markdown AST for display; they are never the source of truth. The source of truth is the file on disk plus its cache entry.

---

## 3. Architecture — modular by construction

The whole point of the modular split is portability across docs backends (§7). Everything that knows about *where docs live and how they're shaped* sits behind one interface; everything else is backend-agnostic.

```
wails3 translate (clir subcommand)
        │
        ▼
  cmd wiring ──► opens a Wails desktop window (mirrors v3/internal/setupwizard)
        │
        ▼
┌────────────────────────────── Go core (v3/internal/i18n) ────────────────────────────────┐
│                                                                                          │
│  api/            Bound API surface exposed to the frontend (the Go↔JS contract, §6)      │
│       │                                                                                   │
│       ├── backend/     DocsBackend interface  ◄── the portability seam (§7)              │
│       │     └── starlight/   adapter for the current Astro Starlight docs                 │
│       │         (future docs backends drop in here behind the same interface)             │
│       │                                                                                   │
│       ├── catalog/     locale registry + coverage/status computation                     │
│       ├── cache/       .translation-cache read/write, hash scheme, status+approval schema │
│       ├── engine/      orchestration: manual + LLM translate, MDX-safe transforms (§8)    │
│       │     └── (link rewriting, code-fence protection, frontmatter handling)            │
│       ├── llm/         provider abstraction + per-language model recommendation (§9)      │
│       └── review/      approval state (viewed/approved), provenance, submission/export    │
│                                                                                          │
└──────────────────────────────────────────────────────────────────────────────────────────┘
        │
        ▼
  frontend/  (Vite + Tailwind, Digital Wails theme — mirrors setupwizard/frontend)
```

### Why this shape
- **`backend/` is the only package that imports knowledge of the docs layout.** Adding a future docs backend means writing one new adapter implementing `DocsBackend` and registering it. Nothing in `catalog/`, `engine/`, `llm/`, `review/`, or the frontend changes.
- **`engine/` is backend-agnostic** — it operates on the abstract `SourceDoc` / `TargetDoc` values the backend hands it, never on raw paths it constructs itself.
- **`api/` is the single bound surface.** The frontend never reaches into a sub-package. This is what lets the same frontend run unchanged against either backend, and (later) lets the whole thing be driven headless/CLI for CI.

### Precedent to mirror
`v3/internal/setupwizard/` is already a Wails-window app with the Digital Wails theme (glass morphism, Inter, `digital_wales_master.webp` background, Tailwind config). The translate tool copies that structure: `frontend/` built with Vite, assets embedded via `go:embed`, window opened via `application.New(...)` + `BundledAssetFileServer`. Reuse its `index.css` tokens and Tailwind config verbatim where possible.

---

## 4. Existing pipeline this builds on (and a dependency to absorb)

The current docs translation setup (discovered in-repo):

- **Docs site:** Astro Starlight at `docs/`, content under `docs/src/content/docs/`, per-locale subdirs (`docs/src/content/docs/<lang>/...`). Locales configured in `docs/astro.config.mjs`.
- **Cache:** `docs/.translation-cache/<locale>.json`, keyed by source path → `{ "hash": <sha256[:12]>, "translated_at": <iso8601> }`. **These files exist on the current branch.**
- **Staleness:** a file is stale if it's absent from the cache or its cached hash ≠ the current source hash.
- **Helper + config + glossary:** a Go helper (`docs/scripts/translate/` with `list-stale` / `record` / `status`), `docs/scripts/translate.config.json` (sourceRoot, cacheDir, ignore globs, locale list), and glossary/do-not-translate rules. **These currently live only in a separate worktree branch (`docs-translate-skill`), not on this branch.**

> **Dependency, not existing infra.** `wails3 translate` should **absorb** that helper's responsibilities into the `cache/` and `catalog/` packages rather than shelling out to a script that isn't merged. Treat `translate.config.json` and the glossary as inputs to load. Plan: either (a) land the helper branch first and have the tool call the same Go functions, or (b) reimplement `list-stale`/`record`/`status` inside `cache/` (small, well-specified) and keep the JSON formats identical so the agent-driven flow and the GUI stay interoperable. Recommend (b) — fewer cross-branch ordering constraints, and the logic is trivial.

### Two files: committed cache + personal review state
Review state is **personal, never committed** (§10), so it's kept out of the committed cache. There are two files:

**1. Committed cache** — the existing `docs/.translation-cache/<locale>.json`, extended only with drift-tracking data that legitimately travels with the repo (so staleness and source-drift work on any clone). Older readers ignore unknown keys:

```jsonc
// docs/.translation-cache/<locale>.json   (committed)
{
  "quick-start/why-wails.mdx": {
    "hash": "053b3578f336",          // sha256[:12] of the SOURCE file at translation time (unchanged)
    "source_snapshot": ".snapshots/quick-start/why-wails.mdx",  // new: copy of the SOURCE as translated (see §8 source-drift)
    "translated_at": "2026-06-29T10:00:00Z"
  }
}
```

**2. Personal review state** — a local, **gitignored** file holding the review progress the UI binds to. Never committed; it's your private "where was I / what have I approved" tracker:

```jsonc
// docs/.translation-cache/.review/<locale>.json   (gitignored — never committed)
{
  "quick-start/why-wails.mdx": {
    "status": "approved",            // untranslated | machine | needs_review | approved
    "viewed": true,                  // "I've looked at this file"
    "provenance": { "method": "llm", "model": "claude-opus-4-8" }  // how it was produced
  }
}
```

The freshness states (up-to-date / stale / missing / orphaned, §5.6) are derived from the committed cache; the review state above is layered on top, locally.

---

## 5. User flows

### 5.1 Launch & dashboard
`wails3 translate` →
1. Detect the docs backend (§7). If none found, show a one-screen "point me at your docs" picker (root dir) and offer to confirm the detected backend.
2. Open the branded window on the **Dashboard**: a grid of locale cards. Each card shows the language (label + native name + flag-free chip), a **freshness bar** broken into Up-to-date / Stale / Missing (and Orphaned if any), the approved count, and how far behind it is (§5.6). An "Add a language" card sits at the end.

A **summary banner** sits above the grid when English has moved on: e.g. "English has 3 new files and 12 changed files since locales were last updated." Coverage/freshness numbers come from `catalog/` cross-referencing the backend's current source list with the cache (see §5.6 for how each state is derived).

### 5.2 Edit an existing locale
From a locale card → **Workspace** for that locale:
- Left: file tree (mirrors the source tree), each file badged with status (untranslated / stale / machine / needs-review / approved) and a viewed checkmark.
- Center: the file in the **embedded editor** (§11) — either the raw target file, or the side-by-side block view (toggle). Edit, save (writes file, re-stamps hash, sets `status` appropriately).
- "Edit which files you need to" is exactly this: pick files in the tree, edit, save. No LLM required.

### 5.3 Create a new language
"Add a language" → pick a target language (from the configured locale registry, or add a new locale code which also offers to wire it into the backend's locale config, e.g. append to `astro.config.mjs`). Then choose:

- **Manual** → opens the Workspace with all files untranslated; user translates block-by-block, Crowdin-style.
- **LLM-assisted** → LLM config screen (§9) → pick scope (whole site / a section / stale-only) → run → land on the **Review** screen.

### 5.4 Review (LLM output)
The heart of the LLM path, modeled on a GitHub PR review:
- File list on the left, each with a **"Viewed"** checkbox and an **"Approve"** toggle.
- Center: per-file **diff/side-by-side** — source on the left, generated translation on the right, block-aligned. Editable in place (the embedded editor), so the user can fix machine output before approving.
- Approving a file sets `status: approved` + `viewed: true`. Files default to `needs_review` after machine translation.
- A header progress indicator: "12 / 40 files reviewed, 8 approved."
- Nothing is considered done until reviewed; the dashboard coverage distinguishes `machine` from `approved`.

### 5.5 Submit
When the user is happy: **Submit** offers two paths — automated via `gh`, or manual. The files are already written to the working tree either way.

- **Automated (`gh`)** — the convenient path for a fork-based contribution. If the GitHub CLI is available and authenticated, the tool: creates a branch, commits the changed translation files, pushes to the contributor's **fork**, and opens a PR against upstream with a templated body (locale, files changed, model used). `gh` already understands the fork→upstream flow (it'll create the fork if needed and target the parent repo), so we lean on it rather than inventing GitHub tooling. This is the default automated approach.
- **Manual** — for users who'd rather drive git/GitHub themselves (or don't have `gh`): show the changed-file list and the exact commands to run (branch, commit, push to their fork, open the PR), so they can do it by hand. Always available as a fallback; never required.

We don't reimplement git/PR plumbing in Go — `gh` for automation, copyable git commands for manual. (No reliance on `go get` or bespoke GitHub clients.)

### 5.6 When English has moved on (new + changed source files)

This is the everyday case: someone updated the English docs, and now you run `wails3 translate`. Here's exactly what happens.

**On launch, the tool re-scans and derives a state per source file, per locale** (`catalog/` comparing the backend's current source list + hashes against each locale's cache):

| Derived state | How it's detected | Meaning |
|---|---|---|
| **Up-to-date** | in cache, cached source-hash == current source-hash | translation matches current English |
| **Stale** | in cache, cached hash != current hash | English changed since this file was translated |
| **Missing** | source exists, no cache/target entry | never translated — includes **brand-new English files** |
| **Orphaned** | target/cache entry exists, source is gone | English file was deleted or renamed |

These freshness states are **derived** (recomputed every launch); they're separate from the stored review `status` (machine / needs_review / approved, §4). A file can be e.g. "Stale + previously-approved."

**How far behind a translation is** — quantified, not just a flag:
- Always available (no git needed): counts per locale — "8 stale, 3 missing, 1 orphaned" and a freshness % = up-to-date / total.
- Richer when the repo is a git checkout (it is): for each stale/missing file, show **how long and how many commits** since the English last changed (e.g. "English edited 3 weeks ago, 5 commits ahead of this translation"). The per-locale card surfaces the worst offender ("most behind: 6 weeks"). This answers "how far behind certain translations are" concretely.

**What you do about it — actioning, from the dashboard and the workspace:**
- Each locale card with work to do shows quick actions:
  - **Translate missing (N)** — runs the chosen path (manual workspace, or LLM job) over just the never-translated files (new English files included).
  - **Update stale (N)** — targets only changed files. For each, the review screen opens with the **source-vs-source diff** (old snapshot vs current English, via `GetSourceDiff` / §8 `source_snapshot`) so you see *exactly which blocks changed in English* and update only those in the target — you don't re-read the whole page. With the LLM path you can re-translate just the changed regions or regenerate the whole file; either way it lands in the normal review/approve flow.
  - **Review orphans (N)** — lists target files whose source is gone; offer to delete (source removed) or keep+rename (source moved). Never auto-deleted.
- In the **Workspace**, the file tree badges every file with its freshness state, and a "Needs attention" filter collapses the tree to Stale + Missing + Orphaned so a maintainer can work straight down the list.
- Bulk option: **"Bring this locale up to date"** = translate-missing + update-stale in one job, then drop you into the review screen with everything queued.

Nothing here changes the file-level model — staleness, the source-snapshot diff, and these actions all sit on top of the existing hash cache (§4) and engine (§8).

---

## 6. The Go↔frontend API (the backbone)

This is the bound surface in `api/`. It ties every screen to the core and is identical across backends. Method names are illustrative; shape is the contract.

| Method | Purpose |
|---|---|
| `DetectBackend() BackendInfo` | What docs system is in use, root path, confidence, whether confirmation is needed. |
| `ListLocales() []LocaleStatus` | Dashboard data: each locale's label/native name, total/translated/stale/approved counts, lastUpdated. |
| `GetTree(locale) []FileNode` | File tree for a locale with per-file status + `viewed`. |
| `GetFile(locale, path) FileView` | Source text + target text (if any) + parsed blocks for side-by-side. |
| `SaveFile(locale, path, content)` | Write target file, re-stamp hash, recompute status. |
| `GetBlocks(locale, path) []Block` | On-the-fly block segmentation for the review/edit view (presentation only). |
| `RecommendModels(targetLang) []ModelSuggestion` | Per-language model ranking + rationale (§9). |
| `GetLLMConfig() / SetLLMConfig(cfg)` | Session-only provider/model/baseURL + key status (never returns the raw key — §10). Includes built-in presets like `ai-master`. |
| `ExportCredentials(path) / LoadCredentials(path)` | Opt-in save/load of LLM config to a user-chosen file; `ExportCredentials` rejects paths inside the repo (§10). |
| `Translate(req) TranslateJob` | Start a machine-translation job (locale, scope, model). Returns a job id. |
| `JobEvents(jobId)` (stream/events) | Progress: per-file start/done/error, token usage, ETA. |
| `GetDiff(locale, path) Diff` | Source↔generated block-aligned diff for the review screen. |
| `GetSourceDiff(locale, path) Diff` | Old-source-snapshot↔current-source diff for **stale** files (what changed in English — §5.6). |
| `SetViewed(locale, path, bool)` / `SetApproved(locale, path, bool)` | Review state writes. |
| `AddLocale(code, opts)` | Register a new locale (and optionally wire it into the backend's locale config). |
| `StartPreview(locale) / StopPreview()` | Launch/stop the backend's live preview server; returns the local URL (§12). |
| `PreviewEvents()` (stream/events) | Preview server status: starting / ready+URL / rebuilt / error. |
| `Submit(locale, mode)` | `export` (return changed files + git hints) or `pr` (branch/commit/push/open via `gh`). |

Streaming (`JobEvents`) follows the Wails events pattern; the rest are request/response bound methods. The frontend holds no docs-layout knowledge — it only knows this surface.

---

## 7. Portability: the `DocsBackend` seam

Portability across docs backends is a **design consideration**, not a v1 feature: there is one backend today (Starlight), and the tool does not target or detect any specific second generator. What the seam buys us is that swapping or adding a docs backend later is an *additive* change — write one adapter — rather than a rewrite. We pin down the contract: the set of things that differ between docs generators.

```go
// backend/backend.go
type DocsBackend interface {
    // Identity & detection
    Name() string                                  // "starlight"
    Detect(root string) (bool, Confidence)         // marker-file probe

    // Content layout
    SourceLocale() string                          // e.g. "en" / root
    ListLocales() ([]Locale, error)                // declared locales + labels/native names
    ListSourceDocs() ([]SourceRef, error)          // translatable files (honors ignore globs)
    TargetPath(loc Locale, src SourceRef) string   // where a translation for src lives
    ReadDoc(ref DocRef) (Doc, error)               // load source/target, with frontmatter split out
    WriteDoc(ref DocRef, doc Doc) error            // persist a translation

    // Conventions the engine must respect
    Format() DocFormat                             // md | mdx | other; what's protectable
    RewriteInternalLinks(doc Doc, loc Locale) Doc  // per-locale link prefixing (§8)
    AddLocale(code string, opts AddLocaleOpts) error // wire a new locale into the site config

    // Build / preview (for an optional "preview my translation" affordance)
    PreviewCommand(loc Locale) (Command, bool)
}
```

These are the concerns any adapter must answer; the Starlight adapter is the concrete (and only) implementation:

| Concern | Starlight adapter |
|---|---|
| Locate the docs | `docs/astro.config.mjs` present + `@astrojs/starlight` dependency |
| Content path layout | `docs/src/content/docs/<locale>/...` |
| Locale declaration | `locales: {...}` in `astro.config.mjs` |
| Frontmatter/format | MDX with YAML frontmatter, JSX components |
| Internal-link rule | prefix internal links with `/<locale>/` (see the existing `i18n(id)` link-prefix commit) |
| Preview command | `npm run dev` in `docs/` |

**Locating the docs:** `Detect(root)` finds the Starlight site (its marker files) so the tool knows where the docs live; if it can't, the UI asks the user to point at the docs root. There is no multi-backend contest — that machinery only becomes relevant if a second backend is ever added.

**Build order:** ship the Starlight adapter concrete and complete. Because everything else depends only on `DocsBackend` (not on Starlight directly), adding another docs backend later is "write one adapter," not a rewrite — which is the entire point of the seam.

---

## 8. Engine: MDX-safe translation (the known pain point)

The repo already shows where translation goes wrong here — the `i18n(id)` work needed internal links prefixed with `/id/`, and there's a do-not-translate glossary. So the engine treats structure preservation as a **first-class concern**, not an afterthought. For every file, machine or manual:

1. **Split frontmatter** from body; translate only whitelisted frontmatter keys (`title`, `description`, …), never keys/IDs/slugs.
2. **Protect code** — fenced code blocks and inline code are masked before sending to an LLM and restored after. Never translated.
3. **Protect MDX/JSX** — component names, props, and import statements are preserved; only human-readable text nodes are translated.
4. **Glossary / do-not-translate** — terms from the config (Wails, Go, goroutine, WebView2, CLI, etc.) are pinned. Passed to the LLM as constraints and verified post-hoc.
5. **Rewrite internal links per-locale** — delegated to `backend.RewriteInternalLinks`, because the rule is backend-specific (Starlight wants `/<locale>/...`).
6. **Validate round-trip** — after translation, assert the block structure matches the source (same number/order of code fences, headings, components); flag mismatches in the review screen rather than writing silently-broken MDX.

This logic lives in `engine/` and runs identically for both creation paths (manual editing reuses the same protect/restore so a user can't accidentally translate a code fence in the block view).

### Source-drift: the canonical stale-translation problem
File-level staleness tells you *that* a source changed, not *what* changed — and a translator updating a 300-line page shouldn't have to re-read all of it to find the three edited paragraphs. Mitigation that stays inside the file-level model (no move to segment-level): when a translation is written, also snapshot the **source content as it was at translation time** (`source_snapshot` in the cache, §4). When that file later goes stale, the review screen shows a **source-vs-source diff** (old snapshot vs. current source) so the translator sees exactly which blocks changed and updates only those in the target. This delivers the practical benefit of segment-level change tracking — the thing users actually feel — without segment-level caching. (If the snapshot is absent — e.g. translations created before this field existed — fall back to whole-file "stale" with no diff.)

---

## 9. LLM configuration & per-language model recommendation

### Provider abstraction (`llm/`)
A small interface so the tool isn't tied to one vendor:

```go
type Provider interface {
    Translate(ctx, TranslateInput) (TranslateOutput, error) // input: protected source, target lang, glossary
    Name() string
}
```

Adapters: **Anthropic** (default), generic **OpenAI-compatible**, and **Ollama / local**. Config: provider, model, base URL, max-concurrency, and (for cloud providers) an API key that is **not persisted by default** (§10).

### Built-in `ai-master` preset (hard-coded)
The tool ships a hard-coded provider preset for the internal **`ai-master`** Ollama box, so maintainers get a zero-config local option:

```jsonc
// built-in, compiled into the binary — not user config
{
  "id": "ai-master",
  "label": "ai-master (internal Ollama)",
  "provider": "ollama",
  "baseURL": "http://100.73.175.21:11434",   // Tailscale IP, not localhost
  "models": ["qwen3.6:35b", "qwen3.5-9b-distill"],
  "requiresKey": false
}
```

- No API key needed (Ollama, internal network).
- It's **reachability-gated**: on startup the tool probes the endpoint; if unreachable (e.g. a community user not on the Tailnet), the preset is shown disabled with a "internal-only" note rather than erroring. Internal users on the Tailnet get it preselected as the default local option.
- The endpoint is hard-coded but the model list is just a default — the dropdown still lets you type any model the box serves.

### Default to the latest Claude models
When the Anthropic provider is selected, default the model to **`claude-opus-4-8`** (Claude Opus 4.8 — current most capable Opus tier). Other current IDs available for the model dropdown: `claude-fable-5` (most capable overall), `claude-sonnet-4-6` (balanced/cheaper for high-volume), `claude-haiku-4-5` (fast/cheap). These IDs are current as of this spec; the dropdown should be data-driven (see below) so they don't rot.

### Per-language recommendation — an updatable map, not hardcoded truth
"Suggest good LLMs for translations based on the language" is implemented as a **config-driven ranking with rationale**, so it can be updated as models change without touching code:

```jsonc
// translate.models.json  (loaded at runtime; editable by maintainers)
{
  "default": { "rank": ["claude-opus-4-8", "claude-fable-5", "claude-sonnet-4-6"],
               "rationale": "Strong general multilingual quality." },
  "byLanguage": {
    "ja":   { "rank": ["claude-fable-5", "claude-opus-4-8"],
              "rationale": "CJK: prefer the most capable model for nuance and honorifics." },
    "zh-cn":{ "rank": ["claude-fable-5", "claude-opus-4-8"], "rationale": "CJK nuance." },
    "ko":   { "rank": ["claude-fable-5", "claude-opus-4-8"], "rationale": "CJK + honorific register." },
    "cy":   { "rank": ["claude-opus-4-8", "claude-fable-5"],
              "rationale": "Lower-resource language: prefer the most capable, verify with a native reviewer." }
  },
  "localPrivacyNote": "For sensitive/offline work, an Ollama model trades quality for privacy; always human-review."
}
```

`RecommendModels(targetLang)` returns this ranking + rationale; the UI shows the top suggestion preselected with a "why" tooltip and the full list as alternatives. Guidance encoded:
- **High-resource (ja, zh, ko, de, fr, es, pt, ru):** any current Claude model does well; prefer the most capable for CJK and for register-sensitive languages.
- **Lower-resource (cy/Welsh, id, etc.):** prefer the most capable model and **strongly recommend native review** — surfaced as a banner on the review screen for those locales.
- **Privacy/offline:** local model option, with an explicit quality caveat.

This keeps recommendations honest and maintainable: when a new model ships, edit one JSON file.

---

## 10. State, secrets, and where things live

- **LLM API keys — not persisted by default.** A key the user enters lives **in memory for the session only** and is gone when the window closes. The tool never writes it anywhere unless the user explicitly asks. (The `ai-master` preset needs no key at all, so the common internal path stores nothing.)
  - **Explicit export / save:** an action that writes the current LLM config (provider, model, base URL, key) to a credentials file the user chooses. This is opt-in, never automatic.
  - **Explicit load:** point the tool at a previously-exported file to repopulate the session config.
  - **Default location is outside the Wails repo.** The save dialog defaults to a per-user config dir (e.g. `~/.config/wails3/translate-credentials.json` on Linux/macOS, the platform equivalent elsewhere). We do **not** dictate where it lives — the user can pick any path — we only default away from the checkout.
  - **Hard guard: refuse to save inside the Wails checkout.** Before writing, resolve the chosen path to its canonical form and check it against the repo root (walk up for a `.git` / `go.work`, or compare to the detected repo root). If the target is inside the checkout, **block the write and show an error** ("Don't save credentials inside the repository — they could be committed. Choose a location outside `<repo root>`."). Same guard applies to the default, so we never silently drop a key file in-tree. This also covers the docs checkout and any nested worktree.
  - The bound API passes a key *handle/status* to the frontend (e.g. "a key is set for this session"), never the raw key value.
- **Review state (`viewed` / `status` / provenance) is personal and never committed.** It lives in a local, **gitignored** file — `docs/.translation-cache/.review/<locale>.json` (§4) — separate from the committed cache. It's your private progress tracker (what you've approved, where you left off), not a shared artifact. There is no "share with the team" mode: the tool adds the review-state path to `.gitignore` (and won't write it anywhere that would land in a commit). Only the staleness essentials (hash, snapshot) travel with the repo.
- **Resume:** closing and reopening mid-translation restores job progress and review checkboxes from the local review file. Jobs are resumable (file-level granularity makes this trivial — re-run skips files whose hash already matches).

---

## 11. Embedded editor

- Embed **CodeMirror 6** (lightweight, MDX/Markdown modes, good enough syntax highlighting; smaller and simpler to bundle than Monaco). Markdown/MDX highlighting only — full language-completion is explicitly out of scope (the user left this to discretion; skip it).
- Two modes, toggle per file:
  - **Raw**: edit the target file directly.
  - **Side-by-side blocks**: source block (read-only) next to target block (editable), Crowdin-style. Block edits map back to file regions; save writes the whole file.
- Protected regions (code fences, JSX) are visually marked and read-only in block mode to prevent accidental edits.

---

## 12. Live preview

Translators need to *see* their pages rendered, not just edited as Markdown — especially for MDX where a broken component or a mangled link only shows up in the built site. The docs backend already runs a dev server (Starlight = `npm run dev` in `docs/`, which is Astro/Vite with hot-module reload), so the tool drives that rather than reinventing rendering.

### How it works
- The backend exposes the command via `DocsBackend.PreviewCommand(locale)` (§7) — for Starlight that's `npm run dev` in `docs/`. The `api/` layer starts it as a managed child process (`StartPreview`), captures its local URL (e.g. `http://localhost:4321`), and streams status over `PreviewEvents` (starting → ready+URL → rebuilt → error). `StopPreview` (and window close) tears the process down.
- The rendered page is shown **inside the workbench** next to the editor — an embedded webview/iframe pointed at the dev server's URL for the file currently open (deep-linked to the locale path, e.g. `/ja/quick-start/why-wails`). A "open in browser" affordance is there too for a full-width look.

### Real-time preview
This is the better experience the request asks for, and it falls out of the dev server's existing hot-reload:
- The tool writes edits to the target file (debounced, ~300ms after typing stops, or on save depending on a toggle), Astro/Vite detects the change and hot-reloads, and the embedded preview updates within a second — no manual refresh.
- A small status chip mirrors `PreviewEvents`: "live", "rebuilding…", or an error (so a broken MDX component surfaces immediately instead of silently failing).
- For the side-by-side review screen, the layout is three-up where space allows: **source · translation editor · live rendered preview.**

### Practicalities
- **First-run cost:** the dev server needs `node_modules`. If `docs/node_modules` is absent, prompt to run the install (`npm install`) once, with progress; don't block the rest of the tool on it — preview is optional, editing/translating works without it.
- **Lifecycle:** one preview server at a time, owned by the tool, killed on exit so we don't leak node processes. If the user already has `npm run dev` running, detect the port and offer to attach instead of starting a second one.
- **Backend-agnostic:** because preview goes through `PreviewCommand`, a future docs backend just returns its own preview/serve command and live preview works there too — no preview code changes when the backend swaps.
- **Headless/community:** if Node isn't installed or the command isn't available, preview is disabled gracefully (the editor and translation flow are unaffected).

---

## 13. CLI integration

Register under the existing `clir`-based CLI in `v3/cmd/wails3/main.go`:

```go
app.NewSubCommandFunction("translate", "Open the translation workbench", commands.Translate)
```

`commands.Translate` (in `v3/internal/commands/translate.go`) loads config, instantiates the detected backend, constructs the `api` layer, and opens the Wails window (mirroring how `setupwizard` is launched). Flags (optional, all have sensible defaults):

| Flag | Default | Purpose |
|---|---|---|
| `--dir` | auto-detect | Path to the docs root, if not auto-detected. |
| `--locale` | — | Jump straight into a locale's workspace. |
| `--headless` | false | Run engine-only (no window) for CI: machine-translate stale files for a locale and exit. Reuses the same `engine`/`backend`/`llm` — the modular split makes this nearly free. |

The `--headless` mode is also the bridge back to the existing agent-driven flow: CI (or an agent) can drive the same engine without the GUI.

> **MCP server mode (`--mcp`).** A third machine interface is specified separately in [MCP_SPEC.md](MCP_SPEC.md): a stdio MCP server that lets any harness (Claude Code, local-LLM harnesses) translate the docs **using its own model**, with this tool providing context (source, existing translation, glossary, source-drift) and persistence (structure validation, cache stamping, `needs_review` queuing). It is strictly additive — the LLM providers in §9, the key handling in §10, and `--headless` are unchanged.

---

## 14. UI internationalization (the tool localizes itself)

The translation workbench must itself be internationalized — a tool whose whole purpose is translation should not be English-only. Every user-facing string in the frontend (and any Go-originated message surfaced to the UI) comes from a message catalog, never a hardcoded literal.

- **Message catalogs:** one JSON file per UI locale under `frontend/locales/<locale>.json`, flat key→string (`{ "dashboard.addLanguage": "Add a language", ... }`). `en` is the source/fallback. Missing keys fall back to `en` rather than showing a blank.
- **Frontend wiring:** a lightweight i18n layer (e.g. a small `t(key, vars)` helper backed by the catalogs; no heavy framework dependency). Interpolation for counts/placeholders (`"{translated} / {total} files"`). Pluralization via the standard `Intl.PluralRules`.
- **Locale selection:** detect the OS/browser locale on launch; expose a language switcher in the header so the user can override. The chosen UI locale persists in local config (separate from the *content* locales being translated — don't conflate "what language is the app shown in" with "what language am I translating docs into").
- **RTL readiness:** drive text direction from the UI locale (`dir="rtl"` for ar/he/fa) so the layout flips when an RTL UI catalog is added. Tailwind logical properties where practical.
- **Go-side strings:** any message the Go core sends to the frontend (job errors, validation, status labels) is sent as a **catalog key + params**, not a pre-rendered English sentence, so the frontend localizes it. Keep CLI-only output (terminal logs from `clir`) English for now; in scope is the *windowed UI*.
- **Bundling:** catalogs are embedded via `go:embed` with the rest of the frontend assets, so localized UI ships in the binary with no runtime fetch.

**Dogfooding (elegant, low-cost):** the tool's own `frontend/locales/*.json` catalogs are themselves a translatable corpus. A future `flat-json` `DocsBackend` adapter (the same interface from §7, pointed at the locales dir) would let `wails3 translate` translate *its own UI* — and let the community contribute UI translations through the exact workflow they use for docs. Not required for v1, but the architecture already permits it; note it so the catalog format stays adapter-friendly (flat keys, no logic in values).

---

## 15. Branding

Reuse the Digital Wails identity already established in `v3/internal/setupwizard/frontend`:
- Background: `digital_wales_master.webp` mountain hero, dark base (`#05070b`), glass-morphism cards (`--glass-bg: rgba(255,255,255,0.06)`, soft borders, deep shadows).
- Accent: Wails red→pink gradient (`linear-gradient(135deg, #ff5a4d 0%, #ff2d72 100%)`); gradient text for headings; red focus rings.
- Type: **Inter** with system fallbacks.
- Logo: the white-text Wails wordmark in the header.

Copy the setupwizard's `index.css` tokens and `tailwind.config.js` so the two tools are visually consistent and we don't re-derive the theme. Keep copy positive and benefit-led (no em dashes, no competitor comparisons).

---

## 16. Build phases

1. **Backend seam + Starlight adapter + catalog/cache.** Locate docs, locale/freshness on the dashboard, read/write, hash+status cache, the derived staleness states (§5.6). (No LLM yet — manual editing works end to end.)
2. **Window + theme + Workspace + embedded editor + UI i18n scaffolding.** Mirror setupwizard. Manual create/edit path fully working. Stand up the message-catalog layer (§14) from the start — retrofitting i18n after hardcoding strings is the expensive path; the `en` catalog and `t()` helper go in here.
3. **Engine MDX-safety + LLM providers + recommendation map.** Machine translation with protect/restore + link rewriting; the `ai-master` preset and session-only key handling (§9, §10).
4. **Review screen + source-drift.** Diff/side-by-side, viewed/approve, provenance, progress, and the source-vs-source diff for stale files (§5.6, §8).
5. **Live preview (§12).** Drive the backend's dev server, embedded webview, hot-reload on edit.
6. **Submit (export + `gh`/fork PR) + community polish.** Credential export/load with the in-repo save guard (§10), contributor docs.
7. **Headless mode + seam validation.** A `--headless` run that reuses the entire core, plus a thin fake backend in tests to prove the `DocsBackend` seam holds without a second real backend.

Phases 1–2 deliver a usable manual tool; 3–4 add the Crowdin-with-LLM workflow; 5 adds live preview; 6 opens it to the community; 7 hardens portability.

---

## 17. Resolved decisions

These were open during design and are now settled:

1. **Review/approval state is personal, never shared.** No team. The `viewed` / `status` / provenance marks live in a local, gitignored file (§4, §10); only staleness essentials are committed. No "share with the team" mode.
2. **No second docs backend is named or detected.** Portability is kept purely as a design consideration via the `DocsBackend` seam (§7) — Starlight is the only implementation; adding another later is one adapter, but the tool does not target or detect any specific future generator.
3. **Submission: `gh` automated, manual fallback (§5.5).** The automated path uses the GitHub CLI for the fork→upstream PR flow; users without `gh` (or who prefer it) get copyable git commands. No bespoke GitHub tooling.
4. **Scope is docs v1 only.** App-string (runtime-UI) translation is out of scope (§1); the seam would allow it later, but it is not a planned mode.
