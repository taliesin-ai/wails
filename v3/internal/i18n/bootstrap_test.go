package i18n

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// makeFakeWailsRemote builds a tiny local git repo with a Starlight docs/ tree,
// usable as a clone source so the bootstrap test never touches the network.
func makeFakeWailsRemote(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(p, body string) {
		if err := os.WriteFile(filepath.Join(docs, p), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("astro.config.mjs", `import starlight from '@astrojs/starlight';
export default defineConfig({
  integrations: [starlight({
    locales: {
      root: { label: "English", lang: "en", dir: "ltr" },
      de: { label: "Deutsch", lang: "de", dir: "ltr" },
    },
  })],
})
`)
	write("package.json", `{"dependencies":{"@astrojs/starlight":"^0"}}`)
	if err := os.MkdirAll(filepath.Join(docs, "src", "content", "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(filepath.Join("src", "content", "docs", "index.mdx"), "welcome")

	// A non-docs directory the sparse-checkout must leave out of the working tree.
	if err := os.MkdirAll(filepath.Join(dir, "framework"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "framework", "big.txt"), []byte("not docs"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{
		{"init", "--quiet"},
		{"-c", "user.email=t@t", "-c", "user.name=t", "add", "."},
		{"-c", "user.email=t@t", "-c", "user.name=t", "commit", "--quiet", "-m", "docs"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func TestOfferCloneClonesWhenForced(t *testing.T) {
	remote := makeFakeWailsRemote(t)
	cache := filepath.Join(t.TempDir(), "wails-src")

	origURL, origDir := wailsRepoURL, wailsDocsCacheDir
	wailsRepoURL = remote
	wailsDocsCacheDir = func() (string, error) { return cache, nil }
	t.Cleanup(func() { wailsRepoURL, wailsDocsCacheDir = origURL, origDir })

	a := &App{}
	root, _, err := a.offerCloneWailsDocs(Options{Clone: true})
	if err != nil {
		t.Fatalf("offerCloneWailsDocs: %v", err)
	}
	want := filepath.Join(cache, "docs")
	if root != want {
		t.Fatalf("root = %q, want %q", root, want)
	}
	if _, err := os.Stat(filepath.Join(root, "astro.config.mjs")); err != nil {
		t.Fatalf("cloned docs missing config: %v", err)
	}
	// sparse-checkout must have excluded the non-docs tree.
	if _, err := os.Stat(filepath.Join(cache, "framework")); !os.IsNotExist(err) {
		t.Fatalf("sparse-checkout did not exclude non-docs dir (stat err=%v)", err)
	}
}

func TestOfferCloneReusesExistingClone(t *testing.T) {
	remote := makeFakeWailsRemote(t)
	cache := filepath.Join(t.TempDir(), "wails-src")

	origURL, origDir := wailsRepoURL, wailsDocsCacheDir
	wailsRepoURL = remote
	wailsDocsCacheDir = func() (string, error) { return cache, nil }
	t.Cleanup(func() { wailsRepoURL, wailsDocsCacheDir = origURL, origDir })

	a := &App{}
	// First call clones.
	if _, _, err := a.offerCloneWailsDocs(Options{Clone: true}); err != nil {
		t.Fatal(err)
	}
	// Drop a marker file inside the clone; a reuse must leave it untouched
	// (proving we did not re-clone or pull over the existing workspace).
	marker := filepath.Join(cache, "docs", ".translation-cache", "marker")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Second call with no force and no TTY must reuse, not re-clone.
	root, _, err := a.offerCloneWailsDocs(Options{})
	if err != nil {
		t.Fatal(err)
	}
	if root != filepath.Join(cache, "docs") {
		t.Fatalf("reuse root = %q", root)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("reuse clobbered the workspace marker: %v", err)
	}
}

func TestOfferCloneSkipsWithExplicitDir(t *testing.T) {
	a := &App{}
	root, _, err := a.offerCloneWailsDocs(Options{Dir: "/some/explicit/path"})
	if err != nil {
		t.Fatal(err)
	}
	if root != "" {
		t.Fatalf("expected no clone offer with explicit --dir, got %q", root)
	}
}
