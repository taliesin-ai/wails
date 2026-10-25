package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wailsapp/wails/v3/internal/flags"
)

func bindingsTestOptions(out string) *flags.GenerateBindingsOptions {
	return &flags.GenerateBindingsOptions{
		OutputDirectory: out,
		ModelsFilename:  "models",
		IndexFilename:   "index",
		TimeType:        "string",
		TS:              true,
		Clean:           true,
		Silent:          true,
	}
}

// checkBindingsFailures reports every file in want that is missing from out or
// whose content changed.
func checkBindingsFailures(t *testing.T, out string, want map[string]string) {
	t.Helper()
	for name, content := range want {
		path := filepath.Join(out, filepath.FromSlash(name))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("existing binding %q was deleted: %v", name, err)
			continue
		}
		if string(data) != content {
			t.Errorf("existing binding %q was rewritten: got %q, want %q", name, data, content)
		}
	}
}

// TestGenerateBindingsPreservesOutputWhenPackagesFail verifies that a
// generation run which could not load its packages leaves the existing
// bindings in the output directory intact.
//
// #6215: package load and type-check errors are reported as warnings, not
// errors. The run still wrote the remaining scaffolding into a temp directory
// and then synced it over the output directory, whose delete phase removed
// every binding the current run did not produce. The command reported success
// while the frontend's imports were gone, so Vite failed with
// "Failed to resolve import ... Does the file exist?".
func TestGenerateBindingsPreservesOutputWhenPackagesFail(t *testing.T) {
	if testing.Short() {
		t.Skip("invokes the binding generator, which loads Go packages")
	}

	root := t.TempDir()
	out := filepath.Join(root, "bindings")
	existing := map[string]string{
		"changeme/internal/servicev2.ts": "// existing service bindings",
		"changeme/index.ts":              "// existing index",
	}
	writeFiles(t, out, existing)

	// A package that does not type-check, so the generator reports warnings
	// and emits no service bindings for it.
	pkgDir := filepath.Join(root, "broken")
	writeFiles(t, pkgDir, map[string]string{
		"broken.go": "package broken\n\nvar _ int = \"not an int\"\n",
	})

	// Warnings are reported, not returned as an error, so the command
	// succeeds either way; what matters is that the bindings survive.
	_ = GenerateBindings(bindingsTestOptions(out), []string{pkgDir})

	checkBindingsFailures(t, out, existing)
}

// TestGenerateBindingsPreservesOutputWithoutPackages verifies that a run whose
// patterns match no packages leaves the existing bindings intact.
func TestGenerateBindingsPreservesOutputWithoutPackages(t *testing.T) {
	if testing.Short() {
		t.Skip("invokes the binding generator, which loads Go packages")
	}

	root := t.TempDir()
	out := filepath.Join(root, "bindings")
	existing := map[string]string{
		"changeme/internal/servicev2.ts": "// existing service bindings",
	}
	writeFiles(t, out, existing)

	_ = GenerateBindings(bindingsTestOptions(out), []string{filepath.Join(root, "does-not-exist")})

	checkBindingsFailures(t, out, existing)
}

// TestGenerateBindingsInstallsCleanRun verifies that a successful run still
// replaces the output directory and prunes stale files, so the fix for #6215
// does not disable cleaning altogether.
func TestGenerateBindingsInstallsCleanRun(t *testing.T) {
	if testing.Short() {
		t.Skip("invokes the binding generator, which loads Go packages")
	}

	root := t.TempDir()
	out := filepath.Join(root, "bindings")
	writeFiles(t, out, map[string]string{
		"stale/stale.ts": "// stale binding from a removed service",
	})

	// A package inside this module that exposes a bound service.
	pattern := "github.com/wailsapp/wails/v3/internal/generator/testcases/function_single"
	if err := GenerateBindings(bindingsTestOptions(out), []string{pattern}); err != nil {
		t.Fatalf("clean run failed: %v", err)
	}

	if _, err := os.Lstat(filepath.Join(out, "stale")); !os.IsNotExist(err) {
		t.Errorf("stale binding should have been pruned, got err=%v", err)
	}

	// Bindings are generated under the package's import path. Check that real
	// output landed there rather than depending on a specific file name.
	pkgOut := filepath.Join(out, filepath.FromSlash(pattern))
	var generated []string
	_ = filepath.WalkDir(pkgOut, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			generated = append(generated, path)
		}
		return nil
	})
	if len(generated) == 0 {
		t.Errorf("expected generated bindings under %q, got none", pkgOut)
	}
}
