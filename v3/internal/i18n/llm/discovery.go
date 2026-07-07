package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// discoveryClient has a short timeout: model listing / local scans must fail
// fast (e.g. when no local Ollama is running) rather than hang the UI.
var discoveryClient = &http.Client{Timeout: 4 * time.Second}

// AnthropicModels is the fixed set of Anthropic models offered in the dropdown.
// Anthropic's lineup is fairly stable, so we hard-code it (Fable is excluded -
// it is unavailable). claude-opus-4-8 is the default/recommended.
func AnthropicModels() []string {
	return []string{
		"claude-opus-4-8",
		"claude-opus-4-7",
		"claude-opus-4-6",
		"claude-sonnet-4-6",
		"claude-haiku-4-5",
	}
}

// suggestedSeeds are curated, translation-friendly picks per provider, in
// preference order. For anthropic the seeds are exact ids; for openrouter and
// ollama they are matched case-insensitively as substrings against the models a
// user actually has available (their ids carry vendor prefixes and version tags
// that drift over time, so an exact list would rot).
var suggestedSeeds = map[string][]string{
	"anthropic": {"claude-opus-4-8", "claude-sonnet-4-6"},
	"openrouter": {
		"anthropic/claude-opus",
		"anthropic/claude-sonnet",
		"google/gemini-2.5-pro",
		"openai/gpt-4o",
		"deepseek/deepseek-chat",
		"qwen/qwen-2.5-72b",
	},
	// Aya is purpose-built for multilingual work; the rest are strong general
	// local models that handle translation well.
	"ollama": {"aya", "qwen2.5", "llama3.1", "gemma2", "mistral-nemo"},
}

// SuggestedModels returns the curated picks for a provider that are actually
// present in available, preserving preference order and de-duplicating. It is
// meant to populate a "Suggested" group at the top of the model dropdown.
func SuggestedModels(provider string, available []string) []string {
	seeds := suggestedSeeds[provider]
	if len(seeds) == 0 || len(available) == 0 {
		return nil
	}
	var out []string
	seen := map[string]bool{}
	for _, seed := range seeds {
		s := strings.ToLower(seed)
		for _, m := range available {
			if seen[m] {
				continue
			}
			if strings.Contains(strings.ToLower(m), s) {
				out = append(out, m)
				seen[m] = true
				break // first (best-ranked available) realisation of this seed
			}
		}
	}
	return out
}

// OpenRouterModels fetches the live model catalogue from OpenRouter's public
// models endpoint (no key required for listing).
func OpenRouterModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return nil, err
	}
	resp, err := discoveryClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openrouter %d", resp.StatusCode)
	}
	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(out.Data))
	for _, m := range out.Data {
		if m.ID != "" {
			ids = append(ids, m.ID)
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// ollamaTags queries an Ollama server's /api/tags for its installed models.
func ollamaTags(ctx context.Context, baseURL string) ([]string, error) {
	url := strings.TrimRight(baseURL, "/") + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := discoveryClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(out.Models))
	for _, m := range out.Models {
		names = append(names, m.Name)
	}
	sort.Strings(names)
	return names, nil
}

// OllamaModels lists models served by the Ollama instance at baseURL.
func OllamaModels(ctx context.Context, baseURL string) ([]string, error) {
	return ollamaTags(ctx, baseURL)
}

// ollamaCandidates are the local endpoints probed when scanning for Ollama.
var ollamaCandidates = []string{
	"http://localhost:11434",
	"http://127.0.0.1:11434",
}

// OllamaScan probes the default local Ollama endpoint(s). It returns the first
// reachable URL and its models; found is false if none responded.
func OllamaScan(ctx context.Context) (url string, models []string, found bool) {
	for _, u := range ollamaCandidates {
		if names, err := ollamaTags(ctx, u); err == nil {
			return u, names, true
		}
	}
	return ollamaCandidates[0], nil, false
}
