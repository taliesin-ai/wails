# `wails3 translate --mcp` — MCP Server Specification

An MCP (Model Context Protocol) server mode for the translation workbench, so that **any MCP client harness — Claude Code, a local-LLM harness, anything that speaks MCP — can translate the docs using its own model**. The server provides the deterministic parts we already own: what's stale, per-file translation context (source, existing translation, glossary, what-changed-in-English drift), and a write tool that validates structure, stamps the cache, and queues the result for human review in the workbench.

This inverts the LLM relationship of [SPEC.md](SPEC.md) §9: instead of the tool calling out to a provider with a user-supplied key, the harness's own model does the translating and calls back into us for context and persistence. No API key ever touches this path.

> **Strictly additive.** The built-in key-based LLM providers (SPEC.md §9), session-only key handling (§10), `--headless` CI mode, and the helper flags (`--status` / `--list-stale` / `--record`) all remain exactly as they are — first-class and fully supported. MCP is a **third machine interface** onto the same engine, not a replacement for any existing path. Nothing in this spec deprecates or alters the workbench's own translation capability.

---

## 1. Goals and non-goals

### Goals
- One flag (`wails3 translate --mcp`) turns the binary into a stdio MCP server any harness can register in one line.
- The harness's model translates; our server supplies context (source, existing target, glossary, drift) and owns persistence (structure validation, cache stamping, review queuing).
- Zero-key operation: this path needs no LLM credentials at all.
- Self-describing: the server teaches the workflow through `instructions` on `initialize` and through tool descriptions — no external skill or doc required to drive it correctly.
- From-nothing bootstrap: a `setup_docs` tool reuses the existing sparse-clone bootstrap so a harness with no checkout can still get to work.
- Agent writes land as `needs_review` with provenance, so the human workbench remains the review surface. The two compose: agent translates over MCP, human reviews in `wails3 translate`.

### Non-goals
- **Not a replacement for the key-based path.** The workbench's built-in providers and `--headless` stay untouched (see the additive note above).
- No MCP **sampling** (server-requested completions): client support is patchy and the inversion makes it unnecessary — the harness already owns the model.
- No **review/approve tools**: review state is personal and human (SPEC.md §17.1). Agents produce `needs_review`; only a person promotes it.
- No **submit tool** in v1: harnesses have their own git tooling, and PR creation is an outward-facing action better left to the human or the harness (future-work candidate, §12).
- No `prompts` capability in v1 (§12); no `resources`; no segment-level persistence (SPEC.md §2 still governs — file-level is the unit of work).

---

## 2. How it fits the existing architecture

The tool now has three machine interfaces, all thin layers over the same core:

| Interface | Invocation | Consumer | LLM |
|---|---|---|---|
| Helper flags | `--status` / `--list-stale` / `--record` | scripts, CI, agents parsing text | none |
| Headless | `--headless --locale <code>` | CI machine-translation runs | ours (key-based, §9) |
| **MCP server** | `--mcp` | any MCP client harness | **the harness's own** |

The MCP server reuses the exact session wiring the helper modes use — `newHelperSession(dir)` in `cli.go`: `starlight.New(root)` → `cacheDir` → `review.NewStore` → `catalog.New` → `engine.New(be, cacheDir, loadGlossary(root))`. No new core logic; the six tools (§5) are adapters over `catalog`, `engine`, `review`, `backend`, and the bootstrap.

Everything a tool returns or persists is identical to what the workbench would produce for the same action — same cache entries, same snapshots, same review records — so the workbench, the helper flags, and MCP interoperate on one source of truth.

---

## 3. Transport and lifecycle

**Transport: stdio.** stdout carries newline-delimited JSON-RPC 2.0 messages and nothing else; stdin carries requests; **all logging, git progress, and diagnostics go to stderr**. This is the standard MCP local-server transport and what `claude mcp add` expects.

- **Read side:** `json.NewDecoder(os.Stdin)` pulling one `json.RawMessage` per iteration. Not `bufio.Scanner` — its default 64KB token limit breaks on a large `write_translation` request (a whole MDX file JSON-escaped on one line easily exceeds it). The decoder has no such limit and exits cleanly on `io.EOF` (client closed stdin → server shuts down).
- **Write side:** `json.NewEncoder(os.Stdout)` — compact single-line output, exactly one trailing `\n` per message, and JSON string escaping guarantees no embedded newlines (the MCP stdio framing requirement).
- **Batch tolerance:** if the first non-space byte of a message is `[`, treat it as a JSON-RPC batch (2025-03-26 clients may send them; 2025-06-18 removed batching). Empty batch `[]` is an invalid request, mirroring the in-repo precedent.
- **Sequential, single-goroutine request loop.** Every tool is a fast filesystem operation; serializing requests costs nothing and eliminates all in-process races by construction.
- **Lazy session.** The server starts even when no docs root is resolvable (that is the `setup_docs` use case). Until a session exists, every tool except `setup_docs` returns the tool-level error `no docs site found; call setup_docs first`. Startup itself never fails on a missing docs root and never prompts.

**stdout hygiene.** The following existing prints must not fire in MCP mode (they would corrupt the protocol stream): the workbench banners in `app.go` (not on this code path), the bootstrap prints in `bootstrap.go` (moved to an `io.Writer` — §7), and the CLI's experimental-warning banner and pterm footer in `commands/translate.go` / `main.go` (the MCP route sets `DisableFooter = true` and skips the banner; the footer only prints after the server exits on EOF, and is disabled anyway).

---

## 4. Protocol subset

Hand-rolled JSON-RPC 2.0, copy-adapted (~200 lines) from the in-repo precedent `v3/pkg/application/mcp_protocol_enabled.go` (#5682): message types, error codes, dispatch switch, result/error helpers, and the tool-schema shape all lift nearly verbatim; the HTTP/CORS/origin/session machinery does not apply. **No external MCP SDK dependency** (repo precedent), and — unlike the app-control server — **no build-tag gating**: this ships in every `wails3` binary because it's a CLI feature, not an embedded-app feature.

Methods implemented:

| Method | Behavior |
|---|---|
| `initialize` | Capability handshake; returns `instructions` (below). Echo the client's protocol version if it's `2024-11-05` or `2025-03-26`; otherwise answer `2025-06-18`. |
| `notifications/initialized` | Accepted, no response (notification). |
| `ping` | `{}` result. |
| `tools/list` | The six tools with JSON schemas. `listChanged: false` — the tool set is static. |
| `tools/call` | Dispatch to §5. Tool failures are returned as tool results with `isError: true` (agent-visible, retryable), not JSON-RPC errors; JSON-RPC errors are reserved for protocol problems (malformed params, unknown method). |

`initialize` result:

```jsonc
{
  "protocolVersion": "2025-06-18",
  "capabilities": { "tools": { "listChanged": false } },
  "serverInfo": { "name": "wails-translate", "title": "Wails Docs Translation", "version": "1.0.0" },
  "instructions": "<the text below>"
}
```

**Instructions text** (the server-shipped workflow — this is how a connecting agent learns the loop with no external skill):

> Translation server for the Wails documentation (Astro Starlight). Workflow: call `translation_status` to pick a locale, `list_files` (filter `needs_work`) for the work queue, then for each file: `get_translation_context`, translate the source into the target language yourself, and submit the complete translated file with `write_translation`. Rules: translate prose only — never translate code blocks, inline code, import lines, JSX/MDX tags, or glossary terms; keep frontmatter keys and translate only the `title` and `description` values; preserve document structure (same headings, fences, components). For stale files, use the `drift` blocks to update only what changed in the source. Writes are validated, cache-stamped, and queued as `needs_review` for human review in the `wails3 translate` workbench. If no docs site is found, call `setup_docs` to clone the Wails docs.

Tool descriptions repeat the load-bearing rules (some clients ignore `instructions`; every client sees tool descriptions).

---

## 5. Tools

Six tools. All results are a single `{"type": "text"}` content block containing JSON (precedent: the app-control server's `mcpToolResult`). Path arguments are source-relative and validated with `backend.ParseSourceRef` (rejects absolute paths, `..`, backslashes, drive letters), so a tool call can never escape the docs tree.

### 5.1 `setup_docs`
No inputs. Resolves a docs root, in order: `findDocsRoot("")` (cwd and parents) → reuse the existing cache clone at `os.UserCacheDir()/wails/translate/wails-src/docs` as-is (never pulled/re-cloned — it doubles as the translation workspace) → forced sparse clone via the existing bootstrap (§7), git progress to stderr. **Never prompts** — MCP mode has no TTY semantics. (Re)initializes the session; safe to call when a session already exists (returns the current root).

```jsonc
// result
{ "docsRoot": "/Users/x/Library/Caches/wails/translate/wails-src/docs",
  "backend": "starlight", "cloned": true,
  "locales": [ { "code": "en", "label": "English", "isSource": true }, { "code": "de", "label": "Deutsch" } ] }
```

### 5.2 `translation_status`
`{ "locale"?: string }`. One `catalog.LocaleStatus` (counts: total / up-to-date / stale / missing / orphaned / approved / machine, freshness %), or the array for all non-source locales via `LocaleStatuses()`. The structs are already JSON-tagged; returned verbatim.

### 5.3 `list_files`
`{ "locale": string, "filter"?: "needs_work" | "stale" | "missing" | "orphaned" | "up_to_date" | "all" }`, default `needs_work` (= stale + missing: the agent's work queue). Returns `catalog.FileStatuses` entries: `{ path, freshness, reviewStatus, hasSnapshot, translatedAt }`.

### 5.4 `get_translation_context`
`{ "locale": string, "path": string, "include"?: string[] }` where `include` ⊆ `["source", "target", "drift", "glossary"]`, default all. One call returns everything needed to translate one file:

```jsonc
{ "path": "guides/routing.mdx",
  "locale": { "code": "de", "label": "Deutsch", "lang": "de", "dir": "ltr" },
  "freshness": "stale", "reviewStatus": "machine",
  "source": "<full raw source file>",
  "target": "<full raw existing translation>",   // null when never translated
  "glossary": ["Wails", "Go", "WebView2", "..."],
  "drift": {                                     // only when stale AND a snapshot exists; else null
    "changedBlocks": [ { "kind": "paragraph", "old": "<block as translated>", "new": "<block now>" } ],
    "totalBlocks": 42, "changedCount": 3 } }
```

`drift` is the source-drift feature (SPEC.md §8) exposed to agents: `engine.SourceDrift` → `engine.Segment` on both bodies → `engine.AlignDiff`, emitting only changed rows — so an agent updating a stale 300-line page sees exactly the three paragraphs that changed in English, in a payload far smaller than two full bodies.

**Sizing: whole files, no pagination.** Measured against the real corpus, docs pages are small (most < 33KB; single outlier `changelog.mdx` at 112KB). `include` is the escape valve — fetch `source` and `target` in separate calls for the outlier. Determinism over cursor machinery.

### 5.5 `write_translation`
`{ "locale": string, "path": string, "content": string, "model"?: string, "force"?: bool }`. `content` is the complete translated file including frontmatter. `model` is the harness's self-reported model name, recorded as provenance. The pipeline mirrors `engine.TranslateFile` semantics (not the workbench's raw-save handler):

1. `ParseSourceRef(path)`; locale must exist and be non-source; the source doc must exist.
2. Split `content` into frontmatter + body (`starlight.SplitDoc`, exported in §7) — link rewriting must never touch frontmatter.
3. `engine.ValidateStructure(sourceBody, contentBody)`.
4. **Reject — `isError`, nothing written — only on a code-fence-count mismatch, unless `force: true`.** A fence mismatch almost always means truncated agent output, and a hard reject enables one clean retry. All other findings (heading count, JSX/component, steps structure) are **write-and-warn**: legitimate variation exists there, and hard-rejecting creates non-converging retry loops. Warnings surface both in the result and in the workbench review screen.
5. Server applies `RewriteInternalLinks` (idempotent — already-prefixed links are skipped). Mechanical rewrites are our job, not the model's.
6. `WriteDoc` → `engine.Stamp(loc, src, sourceRaw)` (hash + source snapshot + timestamp — the file is now up-to-date and drift-diffable next time) → `review.SetStatus(locale, path, "needs_review", Provenance{Method: "mcp", Model: model})`.

```jsonc
// result
{ "written": "docs/src/content/docs/de/guides/routing.mdx",
  "freshness": "up_to_date", "reviewStatus": "needs_review",
  "warnings": [ "heading count differs: source 12, translation 11" ] }
```

### 5.6 `add_locale`
`{ "code": string, "label"?: string, "lang"?: string, "dir"?: string }` (defaults: label = code, lang = code, dir = `ltr`). Calls `backend.AddLocale`, wiring the locale into `astro.config.mjs` exactly as the workbench's add-locale action does.

---

## 6. Client setup

One line for Claude Code:

```sh
claude mcp add wails-translate -- wails3 translate --mcp
```

Or in any client's JSON config:

```jsonc
{ "mcpServers": {
    "wails-translate": { "command": "wails3", "args": ["translate", "--mcp", "--dir", "/path/to/docs"] } } }
```

`--dir` is optional; without it the server auto-detects from cwd, and a harness starting from nothing calls `setup_docs`.

---

## 7. Required internal refactors (small, spec'd here, shipped with the implementation)

1. **`bootstrap.go`: extract the non-interactive clone core.** `cloneWailsDocs(progress io.Writer) (root string, conf backend.Confidence, err error)` — cache-dir resolution, reuse-existing-clone check, stale-dir cleanup, sparse-then-shallow clone, post-clone detect. The three current stdout `fmt.Printf`s move onto `progress`. `offerCloneWailsDocs` keeps its TTY/huh prompt logic and calls the core with `os.Stdout`; `setup_docs` calls it with `os.Stderr`. The existing test seams (`wailsRepoURL`, `wailsDocsCacheDir`) are untouched.
2. **`starlight`: export `splitDoc` as `SplitDoc`.** `write_translation` needs frontmatter/body splitting outside the adapter; the workbench save handler's whole-raw `Doc` construction is not link-rewrite-safe for this path.
3. **`review`: add `"mcp"` to the provenance method enum** (`manual | llm | mcp`).
4. **`review`: fix cross-process staleness in `Store.load`.** The store memoizes each locale's records for the process lifetime, so a long-running workbench would clobber MCP-written `needs_review` records with its stale in-memory copy on its next persist. Fix: re-read the locale file per operation (files are tiny; the existing mutex already serializes). This is a real bug today for any two concurrent processes; MCP just makes it likely.

---

## 8. Concurrency policy

The MCP server and the workbench may run simultaneously against the same docs root — that's a feature (watch files flip to *needs review* live as an agent works).

- **In-process:** the sequential request loop (§3) means no races inside the server.
- **Committed cache (`<locale>.json`):** writers do a fresh load → mutate → save; the cross-process race window is milliseconds and a lost stamp self-heals — freshness is derived per read, so the file simply shows *stale* again and gets re-stamped. Documented, not locked.
- **Review store:** after the §7.4 fix, cross-process behavior is last-writer-wins per record write. Acceptable for personal, advisory state.
- **No lock files / flock.** Windows portability pain, no repo precedent, and the worst case is lost *metadata*, never content corruption (`WriteDoc` is a whole-file write).

---

## 9. CLI integration

`--mcp` is a new bool on `TranslateOptions` (clir derives the flag from the field). It joins `--status` / `--list-stale` / `--record` as a **fourth mutually exclusive mode** in the existing mode-count check in `commands/translate.go`, and additionally errors when combined with `--headless`. The MCP route sets `DisableFooter = true` and skips the experimental-warning banner before handing off to `i18n.RunMCP(i18n.MCPOptions{Dir: options.Dir})`.

A `wails3 translate mcp` subcommand was considered and rejected: clir registers `translate` via `Cli.NewSubCommandFunction`, and hanging a child off it would force restructuring the registration in `main.go` for zero functional gain. The flag matches the established helper-mode precedent. (Revisitable later without breaking anything — the flag can become an alias.)

New files: `v3/internal/i18n/mcp.go` (protocol core, stdio loop, lazy session), `mcp_tools.go` (schemas + the six handlers), `mcp_test.go`. Docs: a new section in `docs/src/content/docs/reference/cli.mdx` and a short guide page.

---

## 10. Testing

The loop is injectable — `serve(r io.Reader, w io.Writer) error` — so tests feed JSON-RPC lines and decode responses, modeled on the app-control server's protocol tests. Fixtures reuse `makeHelperDocs` (cli_test.go: en/de/fr tree with one stale, one missing, one up-to-date file) and `makeFakeWailsRemote` + the bootstrap seams (bootstrap_test.go) for `setup_docs`.

Key cases:
- **Handshake:** initialize (version echo, `serverInfo.name`, non-empty `instructions`), `notifications/initialized` produces no output, ping, `tools/list` exact name set, unknown method → `-32601`.
- **Read tools:** status and list_files against the fixture; `get_translation_context` on the stale file asserts `drift.changedBlocks`.
- **Write happy path:** target bytes on disk, cache entry hash equals the current source hash, snapshot written, review record `{needs_review, {mcp, <model>}}`, and a follow-up `list_files` shows the file `up_to_date`.
- **Write guards:** fence-mismatch → rejected, nothing written; same call with `force: true` succeeds. Path traversal rejected.
- **Framing regression:** one > 64KB single-line `write_translation` request (the `bufio.Scanner` failure class).
- **Lazy session:** tools before setup on an empty dir → `call setup_docs first`; `setup_docs` against the fake remote returns `cloned: true` and unblocks the rest.

Live smoke (documented, run manually — the real proof):

```sh
claude -p "List the stale files for locale de, then translate guides/routing.mdx" \
  --mcp-config '{"mcpServers":{"wails-translate":{"command":"wails3","args":["translate","--mcp","--dir","./docs"]}}}' \
  --allowedTools "mcp__wails-translate__*"
```

Then open `wails3 translate` and confirm the file sits in the review queue as *needs review* with `mcp` provenance.

---

## 11. Build order

1. `bootstrap.go`: extract `cloneWailsDocs(progress io.Writer)`; existing bootstrap tests stay green.
2. `review`: per-operation re-read fix + `mcp` provenance; add a two-store concurrency test.
3. `starlight`: export `SplitDoc`.
4. `mcp.go`: protocol core + stdio loop + lazy session; handshake tests.
5. `mcp_tools.go`: read-only tools first, then `write_translation`, `add_locale`, `setup_docs`; full test suite.
6. `commands/translate.go` + CLI docs: `--mcp` flag, mode exclusion, config snippets.
7. Verify: `go test ./internal/i18n/... ./internal/commands/`, build, piped-initialize smoke (`printf '...' | wails3 translate --mcp`), then the live Claude Code smoke test (§10).

---

## 12. Future work (explicitly out of v1)

- **`prompts` capability** — a `translate-locale` prompt surfaced as a slash command in Claude Code; the `instructions` field already carries the workflow, so this is sugar.
- **`submit` tool** — branch/commit/PR via `gh` for fully autonomous runs; today the harness's own git tooling or the workbench's Submit covers it.
- **Progress notifications** — `notifications/progress` during `setup_docs` clones for clients that render them.
- **Resources** — exposing the glossary or the docs tree as MCP resources; tools cover the need for now.

---

## 13. Resolved decisions

1. **Additive, never a replacement.** The key-based provider path, `--headless`, and helper flags are untouched; MCP is a third interface onto the same core.
2. **stdio + hand-rolled protocol subset, no SDK dependency, no build tags.** Copy-adapt the #5682 precedent; ships in every binary.
3. **The harness's model translates; the server never calls an LLM in this mode.** No sampling.
4. **Review/approve and submit stay human.** Agent output is capped at `needs_review` with `mcp` provenance.
5. **Fence mismatches reject (unless `force`), everything else writes with warnings.** Matches how truncation actually presents, keeps agent retry loops convergent.
6. **`--mcp` flag, not a subcommand** — helper-flag precedent, zero clir churn.
7. **Whole-file tool payloads with an `include` escape valve; no pagination.**
8. **No cross-process locking** — documented last-writer-wins; metadata self-heals, content can't corrupt.
