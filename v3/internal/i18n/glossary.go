package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// defaultGlossary lists do-not-translate terms when no project glossary is found.
var defaultGlossary = []string{
	"Wails", "Go", "Golang", "goroutine", "WebView2", "WebKit", "CLI",
	"JavaScript", "TypeScript", "Vite", "npm", "Taskfile", "macOS", "Windows",
	"Linux", "GTK", "frontend", "backend",
}

// translateConfig is the subset of docs/scripts/translate.config.json we read.
type translateConfig struct {
	Glossary       []string `json:"glossary"`
	DoNotTranslate []string `json:"doNotTranslate"`
}

// loadGlossary reads the project glossary if present, falling back to defaults.
// It looks for docs/scripts/translate.config.json relative to the docs root's
// parent (repo) and the docs root itself.
func loadGlossary(docsRoot string) []string {
	candidates := []string{
		filepath.Join(docsRoot, "scripts", "translate.config.json"),
		filepath.Join(filepath.Dir(docsRoot), "docs", "scripts", "translate.config.json"),
	}
	for _, c := range candidates {
		data, err := os.ReadFile(c)
		if err != nil {
			continue
		}
		var cfg translateConfig
		if json.Unmarshal(data, &cfg) != nil {
			continue
		}
		terms := append(cfg.Glossary, cfg.DoNotTranslate...)
		if len(terms) > 0 {
			return terms
		}
	}
	return defaultGlossary
}
