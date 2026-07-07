package commands

import (
	"fmt"

	"github.com/wailsapp/wails/v3/internal/i18n"
	"github.com/wailsapp/wails/v3/internal/term"
)

// TranslateOptions are the flags for `wails3 translate`.
type TranslateOptions struct {
	Dir       string `description:"Path to the docs root (auto-detected if omitted)"`
	Locale    string `description:"Jump straight into a locale's workspace"`
	Headless  bool   `description:"Run engine-only (no window) for CI: machine-translate stale+missing files for --locale and exit"`
	Clone     bool   `description:"If no docs site is found, clone the Wails docs (no prompt) and translate those"`
	MCP       bool   `name:"mcp" description:"Serve MCP over stdio so an agent harness can translate the docs with its own model"`
	Status    bool   `description:"Print translation status and exit"`
	ListStale bool   `name:"list-stale" description:"Print stale/missing source-relative files and exit"`
	Record    bool   `description:"Stamp cache metadata for an already-written target file and exit"`
	All       bool   `description:"Apply helper mode to all non-source locales"`
	File      string `description:"Source-relative file path for --record"`
}

// Translate opens the translation workbench (or runs headless for CI, or
// serves MCP over stdio for agent harnesses).
func Translate(options *TranslateOptions) error {
	if options.MCP {
		if options.Status || options.ListStale || options.Record {
			return fmt.Errorf("--mcp cannot be combined with --status, --list-stale, or --record")
		}
		if options.Headless {
			return fmt.Errorf("--mcp cannot be combined with --headless")
		}
		// stdout carries the MCP protocol: no banner, no footer.
		DisableFooter = true
		return i18n.RunMCP(i18n.MCPOptions{Dir: options.Dir})
	}
	if mode, ok, err := translateHelperMode(options); ok || err != nil {
		if err != nil {
			return err
		}
		DisableFooter = true
		return i18n.RunHelper(i18n.HelperOptions{
			Mode:   mode,
			Dir:    options.Dir,
			Locale: options.Locale,
			File:   options.File,
			All:    options.All,
		})
	}
	if !options.Headless {
		DisableFooter = true
		term.Println("")
		term.Warning("The translation workbench is experimental.")
		term.Println("")
	}
	app := &i18n.App{}
	return app.Run(i18n.Options{
		Dir:      options.Dir,
		Locale:   options.Locale,
		Headless: options.Headless,
		Clone:    options.Clone,
	})
}

func translateHelperMode(options *TranslateOptions) (i18n.HelperMode, bool, error) {
	modes := 0
	var mode i18n.HelperMode
	if options.Status {
		modes++
		mode = i18n.HelperModeStatus
	}
	if options.ListStale {
		modes++
		mode = i18n.HelperModeListStale
	}
	if options.Record {
		modes++
		mode = i18n.HelperModeRecord
	}
	if modes > 1 {
		return "", false, fmt.Errorf("choose only one translate helper mode: --status, --list-stale, or --record")
	}
	return mode, modes == 1, nil
}
