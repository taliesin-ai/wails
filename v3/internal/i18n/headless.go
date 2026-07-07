package i18n

import (
	"context"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
	"github.com/wailsapp/wails/v3/internal/i18n/catalog"
	"github.com/wailsapp/wails/v3/internal/i18n/review"
)

// runHeadless machine-translates the stale + missing files for a locale and
// exits, reusing the same engine/backend/llm as the GUI (SPEC.md section 13).
// The API key comes from ANTHROPIC_API_KEY / OPENAI_API_KEY in the environment;
// keys are never persisted (SPEC.md section 10).
func (a *App) runHeadless() error {
	if a.opts.Locale == "" {
		return fmt.Errorf("--headless requires --locale <code>")
	}
	loc, ok := a.locale(a.opts.Locale)
	if !ok {
		return fmt.Errorf("unknown locale %q", a.opts.Locale)
	}

	// Seed the key from the environment for non-interactive runs.
	cfg := a.llm.Config()
	switch cfg.Provider {
	case "anthropic":
		if k := os.Getenv("ANTHROPIC_API_KEY"); k != "" {
			cfg.APIKey = k
			a.llm.SetConfig(cfg)
		}
	case "openai":
		if k := os.Getenv("OPENAI_API_KEY"); k != "" {
			cfg.APIKey = k
			a.llm.SetConfig(cfg)
		}
	}

	prov, err := a.llm.Provider()
	if err != nil {
		return err
	}

	files, err := a.cat.FileStatuses(loc)
	if err != nil {
		return err
	}
	var todo []backend.SourceRef
	for _, f := range files {
		if f.Freshness == catalog.Stale || f.Freshness == catalog.Missing {
			todo = append(todo, backend.SourceRef{Path: f.Path})
		}
	}
	fmt.Printf("Translating %d file(s) for %s using %s/%s...\n", len(todo), loc.Code, cfg.Provider, cfg.Model)

	ctx := context.Background()
	model := a.llm.Config().Model
	var failed int
	for i, src := range todo {
		res := a.eng.TranslateFile(ctx, prov, loc, src, model)
		if res.Err != "" {
			failed++
			fmt.Printf("  [%d/%d] %s: ERROR %s\n", i+1, len(todo), src.Path, res.Err)
			continue
		}
		_ = a.rev.SetStatus(loc.Code, src.Path, "machine", review.Provenance{Method: "llm", Model: cfg.Model})
		warn := ""
		if len(res.Warnings) > 0 {
			warn = fmt.Sprintf(" (warnings: %v)", res.Warnings)
		}
		fmt.Printf("  [%d/%d] %s: ok%s\n", i+1, len(todo), src.Path, warn)
	}
	if failed > 0 {
		return fmt.Errorf("%d file(s) failed", failed)
	}
	return nil
}
