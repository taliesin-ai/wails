// Package llm provides the translation-provider abstraction (Anthropic, generic
// OpenAI-compatible, and Ollama/local), session-only key handling, and the
// config-driven per-language model recommendation map.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// TranslateInput is the structure-protected text plus translation metadata
// handed to a provider. Text has code/JSX already masked by the engine.
type TranslateInput struct {
	Text        string
	SourceLang  string
	TargetLang  string
	TargetLabel string
	Glossary    []string
	Model       string
}

// Provider is a translation backend.
type Provider interface {
	Name() string
	Translate(ctx context.Context, in TranslateInput) (string, error)
}

// Config is the session-only LLM configuration. The API key is held in memory
// for the session and never returned to the frontend or persisted by default.
type Config struct {
	Provider string `json:"provider"` // anthropic | openai | ollama
	Model    string `json:"model"`
	BaseURL  string `json:"baseURL"`
	APIKey   string `json:"-"` // never serialised to the frontend
	Preset   string `json:"preset,omitempty"`
	MaxConc  int    `json:"maxConcurrency,omitempty"`
}

// ConfigStatus is the safe, key-free view sent to the frontend.
type ConfigStatus struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"baseURL"`
	Preset   string `json:"preset,omitempty"`
	HasKey   bool   `json:"hasKey"`
	NeedsKey bool   `json:"needsKey"`
	MaxConc  int    `json:"maxConcurrency"`
}

// Preset is a built-in provider preset.
type Preset struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Provider    string   `json:"provider"`
	BaseURL     string   `json:"baseURL"`
	Models      []string `json:"models"`
	RequiresKey bool     `json:"requiresKey"`
	Reachable   bool     `json:"reachable"`
	Note        string   `json:"note,omitempty"`
}

// Manager holds the live session config and exposes presets/recommendations.
type Manager struct {
	mu     sync.RWMutex
	cfg    Config
	models ModelMap
}

// NewManager creates a manager with sane defaults (Anthropic + Opus 4.8),
// overlaid with any previously persisted non-secret choices (provider/model/URL).
func NewManager() *Manager {
	m := &Manager{
		cfg: Config{
			Provider: "anthropic",
			Model:    "claude-opus-4-8",
			MaxConc:  2,
		},
		models: loadModelMap(),
	}
	m.loadPersisted()
	return m
}

// SetConfig updates the session config. An empty APIKey leaves the existing key
// in place (so a config edit doesn't wipe a key the user already entered). The
// non-secret fields are persisted so the choice survives a restart; the key is
// never written (it stays session-only unless explicitly exported).
func (m *Manager) SetConfig(c Config) {
	m.mu.Lock()
	if c.APIKey == "" {
		c.APIKey = m.cfg.APIKey
	}
	if c.MaxConc == 0 {
		c.MaxConc = 2
	}
	m.cfg = c
	pc := persistedConfig{Provider: c.Provider, Model: c.Model, BaseURL: c.BaseURL, MaxConc: c.MaxConc}
	m.mu.Unlock()
	pc.save() // best-effort; never holds the lock for file I/O
}

// persistedConfig is the non-secret subset of Config saved across sessions. API
// keys are NEVER written here: they remain session-only unless the user
// explicitly exports credentials.
type persistedConfig struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"baseURL,omitempty"`
	MaxConc  int    `json:"maxConcurrency,omitempty"`
}

// llmConfigPath returns the path to the persisted non-secret config, in the
// user's standard config dir (outside any repo).
func llmConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "wails", "translate", "llm.json"), nil
}

func (pc persistedConfig) save() {
	p, err := llmConfigPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(pc, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(p, data, 0o644)
}

// loadPersisted overlays previously saved non-secret choices onto the defaults.
func (m *Manager) loadPersisted() {
	p, err := llmConfigPath()
	if err != nil {
		return
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}
	var pc persistedConfig
	if json.Unmarshal(data, &pc) != nil {
		return
	}
	if pc.Provider != "" {
		m.cfg.Provider = pc.Provider
	}
	if pc.Model != "" {
		m.cfg.Model = pc.Model
	}
	m.cfg.BaseURL = pc.BaseURL
	if pc.MaxConc > 0 {
		m.cfg.MaxConc = pc.MaxConc
	}
}

// Config returns a copy of the live config (including the key) for internal use.
func (m *Manager) Config() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

// Status returns the key-free view for the frontend.
func (m *Manager) Status() ConfigStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	needsKey := m.cfg.Provider == "anthropic" || m.cfg.Provider == "openai" || m.cfg.Provider == "openrouter"
	return ConfigStatus{
		Provider: m.cfg.Provider, Model: m.cfg.Model, BaseURL: m.cfg.BaseURL,
		Preset: m.cfg.Preset, HasKey: m.cfg.APIKey != "", NeedsKey: needsKey,
		MaxConc: m.cfg.MaxConc,
	}
}

// Presets returns the built-in provider presets. There are none: the provider is
// configured explicitly (Anthropic / OpenAI-compatible / Ollama).
func (m *Manager) Presets() []Preset { return nil }

// ApplyPreset is retained for API shape; there are no built-in presets.
func (m *Manager) ApplyPreset(id string) error {
	return fmt.Errorf("unknown preset %q", id)
}

// Recommend returns the ranked model suggestions for a target language.
func (m *Manager) Recommend(targetLang string) Recommendation {
	return m.models.Recommend(targetLang)
}

// Provider builds a Provider from the live config.
func (m *Manager) Provider() (Provider, error) {
	c := m.Config()
	switch c.Provider {
	case "anthropic":
		if c.APIKey == "" {
			return nil, fmt.Errorf("Anthropic API key not set for this session")
		}
		return &anthropic{apiKey: c.APIKey, model: c.Model}, nil
	case "openrouter":
		if c.APIKey == "" {
			return nil, fmt.Errorf("OpenRouter API key not set for this session")
		}
		return &openAICompatible{apiKey: c.APIKey, model: c.Model, baseURL: "https://openrouter.ai/api"}, nil
	case "openai":
		if c.APIKey == "" {
			return nil, fmt.Errorf("API key not set for this session")
		}
		base := c.BaseURL
		if base == "" {
			base = "https://api.openai.com"
		}
		return &openAICompatible{apiKey: c.APIKey, model: c.Model, baseURL: base}, nil
	case "ollama":
		base := c.BaseURL
		if base == "" {
			base = "http://localhost:11434"
		}
		return &ollama{baseURL: base, model: c.Model}, nil
	default:
		return nil, fmt.Errorf("unknown provider %q", c.Provider)
	}
}

// translationPrompt builds the system+user prompt shared across providers.
func translationPrompt(in TranslateInput) (system, user string) {
	system = "You are a professional documentation translator. Translate the user's " +
		"Markdown/MDX text into " + in.TargetLabel + " (" + in.TargetLang + "). " +
		"Preserve all Markdown structure, links, and any placeholder tokens of the form " +
		"[[PH:N]] EXACTLY as they appear - never translate, reorder, or alter them. " +
		"Do not translate code, URLs, or frontmatter keys. Return only the translated text, no commentary."
	if len(in.Glossary) > 0 {
		system += "\nKeep these terms untranslated: "
		for i, g := range in.Glossary {
			if i > 0 {
				system += ", "
			}
			system += g
		}
		system += "."
	}
	user = in.Text
	return
}

// marshal helper used by provider adapters.
func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
