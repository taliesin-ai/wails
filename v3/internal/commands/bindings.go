package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/internal/flags"
	"github.com/wailsapp/wails/v3/internal/generator"
	"github.com/wailsapp/wails/v3/internal/generator/config"
	"github.com/wailsapp/wails/v3/internal/term"
)

func GenerateBindings(options *flags.GenerateBindingsOptions, patterns []string) error {
	DisableFooter = true

	if options.Silent {
		term.DisableOutput()
		defer term.EnableOutput()
	} else if options.Verbose {
		term.EnableDebug()
		defer term.DisableDebug()
	}

	if !underWake() {
		term.Header("Generate Bindings")
	}

	if len(patterns) == 0 {
		// No input pattern, load package from current directory.
		patterns = []string{"."}
	}

	// Compute absolute path of output directory.
	absPath, err := filepath.Abs(options.OutputDirectory)
	if err != nil {
		return err
	}

	// When clean mode is active and we're writing real files, generate bindings
	// into a dot-prefixed sibling temp directory first, then sync the result
	// into the output directory file by file. This prevents chokidar (used by
	// Vite) from entering a rename-event loop caused by rapid directory
	// delete+recreate, which would otherwise cause the node process to leak
	// memory at ~2-6 MB/s.
	// Dot-prefixed directories are ignored by chokidar's default glob pattern,
	// so no spurious HMR events fire during file generation.
	generationDir := absPath
	swapped := false
	if options.Clean && !options.DryRun {
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return fmt.Errorf("failed to create bindings parent directory: %w", err)
		}
		tmpDir, err := os.MkdirTemp(filepath.Dir(absPath), ".bindings-tmp-")
		if err != nil {
			return fmt.Errorf("failed to create temp directory for bindings: %w", err)
		}
		generationDir = tmpDir
		defer func() {
			if !swapped {
				_ = os.RemoveAll(tmpDir)
			}
		}()
	} else if options.Clean {
		if err := os.RemoveAll(absPath); err != nil {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}

	// Initialise file creator.
	var creator config.FileCreator
	if !options.DryRun {
		creator = config.DirCreator(generationDir)
	}

	// Under a wake build, forward progress to the build UI as wire events rather
	// than drawing a competing spinner; otherwise use the interactive spinner.
	var logger config.Logger
	var spinner term.Spinner
	if underWake() {
		logger = newWakeLogger()
	} else {
		spinner = term.StartSpinner("Initialising...")
		logger = spinner.Logger()
	}

	// Initialise and run generator.
	stats, err := generator.NewGenerator(
		options,
		creator,
		logger,
	).Generate(patterns...)

	summary := fmt.Sprintf(
		"Processed: %s, %s, %s, %s, %s, %s in %s.",
		pluralise(stats.NumPackages, "Package"),
		pluralise(stats.NumServices, "Service"),
		pluralise(stats.NumMethods, "Method"),
		pluralise(stats.NumEnums, "Enum"),
		pluralise(stats.NumModels, "Model"),
		pluralise(stats.NumEvents, "Event"),
		stats.Elapsed().String(),
	)

	if underWake() {
		logger.Infof("%s", summary)
		logger.Infof("Output directory: %s", absPath)
	} else {
		term.StopSpinner(spinner)
		term.Infof("%s", summary)
		term.Infof("Output directory: %s", absPath)
	}

	// Process generator error.
	//
	// Warnings mean the generation run did not complete normally: the package
	// patterns matched nothing, or some packages could not be loaded or
	// type-checked. In that case the generated output is incomplete, so it must
	// not replace the existing bindings (#6215).
	incomplete := false
	if err != nil {
		var report *generator.ErrorReport
		switch {
		case errors.Is(err, generator.ErrNoPackages):
			// Convert to warning message.
			term.Warning(err)
			incomplete = true
		case errors.As(err, &report):
			if report.HasErrors() {
				// Report error count.
				return err
			} else if report.HasWarnings() {
				// Report warning count.
				term.Warning(report)
				incomplete = true
			}
		default:
			// Report error.
			return err
		}
	}

	// Sync the generated output into place. Files are updated individually
	// rather than swapping the whole directory: on Windows, deleting or
	// renaming over the output directory fails with "Access is denied" while
	// a file watcher (e.g. Vite's dev server) holds it open (#5515).
	if !swapped && generationDir != absPath {
		if synced, err := installBindings(generationDir, absPath, incomplete); err != nil {
			return fmt.Errorf("failed to install new bindings at %q: %w\nThe bindings directory may be partially updated. Re-run the command to retry.", absPath, err)
		} else if !synced {
			term.Warningf("Generation did not complete; the existing bindings in %s were left unchanged. Fix the reported problems and re-run the command.", absPath)
		}
		swapped = true
	}

	return nil
}

// installBindings makes dst's contents match src's, unless incomplete reports
// that the generation run did not complete normally.
//
// syncDirs deletes every file in dst that src did not produce, so installing
// the output of an incomplete run would delete bindings that are still valid.
// When a package fails to load or type-check, no service bindings are emitted
// for it, yet the remaining scaffolding is still written; syncing that would
// leave the developer with a bindings directory that no longer contains what
// their frontend imports (#6215). The existing bindings are preserved and src
// is discarded instead.
//
// It reports whether the bindings were installed.
func installBindings(src, dst string, incomplete bool) (bool, error) {
	if incomplete {
		return false, withRetry(func() error { return os.RemoveAll(src) })
	}
	return true, syncDirs(src, dst)
}

func pluralise(number int, word string) string {
	if number == 1 {
		return fmt.Sprintf("%d %s", number, word)
	}
	return fmt.Sprintf("%d %ss", number, word)
}
