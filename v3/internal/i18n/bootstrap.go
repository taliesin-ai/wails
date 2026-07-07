package i18n

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/wailsapp/wails/v3/internal/git"
	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/backend/starlight"
	"github.com/wailsapp/wails/v3/internal/term"
)

// wailsRepoURL is the upstream repository whose docs/ tree we translate, and
// wailsDocsCacheDir resolves where it is cloned. Both are vars (not consts) so
// tests can point the bootstrap at a local remote and a temp cache dir.
var (
	wailsRepoURL      = "https://github.com/wailsapp/wails.git"
	wailsDocsCacheDir = defaultDocsCacheDir
)

// defaultDocsCacheDir returns the persistent location for a cloned Wails
// checkout used as a translation workspace. It lives in the user's cache dir
// (outside any repo) so the clone — and the .translation-cache/review state that
// accumulates inside it — survives across runs.
func defaultDocsCacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "wails", "translate", "wails-src"), nil
}

// offerCloneWailsDocs is the fallback when no docs site is found in or above the
// cwd. It reuses a previously cloned workspace as-is, or (interactively, or with
// --clone) offers to shallow-clone the Wails docs into the cache dir. It returns
// an empty root when the user declines or no TTY is available, letting the caller
// fall back to the existing fail-fast error.
//
// It deliberately never pulls/updates an existing clone: that clone is also the
// workspace (it holds .translation-cache/, review state and in-progress
// translations), so a silent pull could clobber work or hit merge conflicts.
func (a *App) offerCloneWailsDocs(opts Options) (string, backend.Confidence, error) {
	// Only offer on the no-override, nothing-found path. If the user passed an
	// explicit --dir that didn't resolve, honour that and keep erroring.
	if opts.Dir != "" {
		return "", backend.NoConfidence, nil
	}

	dest, err := wailsDocsCacheDir()
	if err != nil {
		return "", backend.NoConfidence, nil
	}
	docsRoot := filepath.Join(dest, "docs")

	// Reuse an existing clone exactly as it is — never auto-update it.
	if ok, conf := starlight.Detect(docsRoot); ok {
		fmt.Printf("Using the Wails docs cloned earlier at %s\n", docsRoot)
		return docsRoot, conf, nil
	}

	// Decide whether to clone: --clone forces it; otherwise ask, but only when
	// attached to a terminal (CI/headless must never block on a prompt).
	doClone := opts.Clone
	if !doClone {
		if !term.IsTerminal() {
			return "", backend.NoConfidence, nil
		}
		confirm := false
		form := huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("No docs site found here. Clone the Wails documentation?").
				Description(fmt.Sprintf("Shallow-clones %s into\n%s\nso you can start translating. Re-run with --clone to skip this prompt.", wailsRepoURL, dest)).
				Affirmative("Clone").
				Negative("Cancel").
				Value(&confirm),
		))
		if err := form.Run(); err != nil {
			return "", backend.NoConfidence, nil // treat an aborted prompt as a decline
		}
		doClone = confirm
	}
	if !doClone {
		return "", backend.NoConfidence, nil
	}

	root, conf, _, err := cloneWailsDocs(os.Stdout)
	return root, conf, err
}

// cloneWailsDocs is the non-interactive core of the bootstrap: resolve the
// cached Wails docs workspace, cloning it if needed. It never prompts - callers
// decide beforehand that a clone is wanted - so it is safe for MCP mode, where
// stdout carries the protocol and progress must go to stderr. Human-readable
// progress (including git's own output) is written to progress. cloned reports
// whether a fresh clone happened (false when an earlier clone was reused).
func cloneWailsDocs(progress io.Writer) (root string, conf backend.Confidence, cloned bool, err error) {
	dest, err := wailsDocsCacheDir()
	if err != nil {
		return "", backend.NoConfidence, false, err
	}
	docsRoot := filepath.Join(dest, "docs")

	// Reuse an existing clone exactly as it is - never auto-update it: the clone
	// doubles as the translation workspace (.translation-cache/, review state,
	// in-progress work), so a pull could clobber work or hit conflicts.
	if ok, conf := starlight.Detect(docsRoot); ok {
		fmt.Fprintf(progress, "Using the Wails docs cloned earlier at %s\n", docsRoot)
		return docsRoot, conf, false, nil
	}

	// A leftover directory that isn't a valid docs site is a partial/failed
	// clone; clear it so the fresh clone has a clean target.
	if _, statErr := os.Stat(dest); statErr == nil {
		if err := os.RemoveAll(dest); err != nil {
			return "", backend.NoConfidence, false, fmt.Errorf("clear stale clone at %s: %w", dest, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", backend.NoConfidence, false, fmt.Errorf("prepare cache dir: %w", err)
	}

	fmt.Fprintf(progress, "Cloning the Wails documentation into %s\n", dest)
	// Sparse + partial clone of just docs/ is a fraction of the full monorepo
	// (~46MB vs ~200MB). Fall back to a plain shallow clone if the git version or
	// server doesn't support partial/sparse clone.
	if err := git.CloneSparse(wailsRepoURL, dest, "", []string{"docs"}, progress); err != nil {
		fmt.Fprintf(progress, "sparse clone unavailable (%v); cloning the full docs tree instead\n", err)
		_ = os.RemoveAll(dest)
		if err := git.CloneShallow(wailsRepoURL, dest, "", progress); err != nil {
			return "", backend.NoConfidence, false, fmt.Errorf("clone Wails docs: %w", err)
		}
	}

	if ok, conf := starlight.Detect(docsRoot); ok {
		fmt.Fprintf(progress, "Cloned. Translating the docs at %s\n", docsRoot)
		return docsRoot, conf, true, nil
	}
	return "", backend.NoConfidence, false, fmt.Errorf("cloned %s but found no docs site at %s", wailsRepoURL, docsRoot)
}
