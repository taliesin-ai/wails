package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 120 * time.Second}

// ---- Anthropic ----

type anthropic struct {
	apiKey string
	model  string
}

func (a *anthropic) Name() string { return "anthropic" }

func (a *anthropic) Translate(ctx context.Context, in TranslateInput) (string, error) {
	model := in.Model
	if model == "" {
		model = a.model
	}
	system, user := translationPrompt(in)
	// Stream so long translations never hit a request timeout, and set a high
	// output ceiling so a document is never silently truncated. The engine caps
	// per-call input (see engine.maxChunkChars), so this headroom is generous.
	body := map[string]any{
		"model":      model,
		"max_tokens": 32000,
		"system":     system,
		"messages":   []map[string]any{{"role": "user", "content": user}},
		"stream":     true,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(mustJSON(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	return readAnthropicStream(resp.Body)
}

// readAnthropicStream consumes the Messages API SSE stream, accumulating
// text_delta content and surfacing any mid-stream error event.
func readAnthropicStream(r io.Reader) (string, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var sb strings.Builder
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(line[len("data:"):])
		if payload == "" {
			continue
		}
		var ev struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue
		}
		switch ev.Type {
		case "content_block_delta":
			if ev.Delta.Type == "text_delta" {
				sb.WriteString(ev.Delta.Text)
			}
		case "error":
			return "", fmt.Errorf("anthropic stream error: %s", ev.Error.Message)
		}
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("anthropic stream read: %w", err)
	}
	return sb.String(), nil
}

// ---- OpenAI-compatible ----

type openAICompatible struct {
	apiKey  string
	model   string
	baseURL string
}

func (o *openAICompatible) Name() string { return "openai" }

func (o *openAICompatible) Translate(ctx context.Context, in TranslateInput) (string, error) {
	model := in.Model
	if model == "" {
		model = o.model
	}
	system, user := translationPrompt(in)
	body := map[string]any{
		"model": model,
		"messages": []map[string]any{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		o.baseURL+"/v1/chat/completions", bytes.NewReader(mustJSON(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "Bearer "+o.apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai: empty response")
	}
	return out.Choices[0].Message.Content, nil
}

// ---- Ollama / local ----

type ollama struct {
	baseURL string
	model   string
}

func (o *ollama) Name() string { return "ollama" }

func (o *ollama) Translate(ctx context.Context, in TranslateInput) (string, error) {
	model := in.Model
	if model == "" {
		model = o.model
	}
	system, user := translationPrompt(in)
	body := map[string]any{
		"model":  model,
		"system": system,
		"prompt": user,
		"stream": false,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		o.baseURL+"/api/generate", bytes.NewReader(mustJSON(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	var out struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	return out.Response, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
