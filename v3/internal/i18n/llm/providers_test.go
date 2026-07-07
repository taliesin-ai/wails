package llm

import (
	"strings"
	"testing"
)

func TestReadAnthropicStream(t *testing.T) {
	// A minimal but representative Messages API SSE transcript.
	stream := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"x"}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hola"}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" mundo"}}`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		``,
	}, "\n")

	got, err := readAnthropicStream(strings.NewReader(stream))
	if err != nil {
		t.Fatal(err)
	}
	if got != "Hola mundo" {
		t.Errorf("got %q, want %q", got, "Hola mundo")
	}
}

func TestSuggestedModels(t *testing.T) {
	// anthropic: exact ids, only those present are kept, in preference order.
	if got := SuggestedModels("anthropic", AnthropicModels()); len(got) != 2 || got[0] != "claude-opus-4-8" || got[1] != "claude-sonnet-4-6" {
		t.Errorf("anthropic suggestions = %v", got)
	}

	// openrouter: substring match against versioned/prefixed ids, preference order.
	or := []string{
		"openai/gpt-4o-2024-11-20",
		"qwen/qwen-2.5-72b-instruct",
		"anthropic/claude-sonnet-4.6",
		"anthropic/claude-opus-4.8",
		"some/unrelated-model",
	}
	got := SuggestedModels("openrouter", or)
	want := []string{"anthropic/claude-opus-4.8", "anthropic/claude-sonnet-4.6", "openai/gpt-4o-2024-11-20", "qwen/qwen-2.5-72b-instruct"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("openrouter suggestions = %v, want %v", got, want)
	}

	// ollama: matches installed tags by base name; nothing matched -> nil.
	if got := SuggestedModels("ollama", []string{"qwen2.5:7b", "phi3:latest"}); len(got) != 1 || got[0] != "qwen2.5:7b" {
		t.Errorf("ollama suggestions = %v", got)
	}
	if got := SuggestedModels("ollama", []string{"phi3:latest"}); got != nil {
		t.Errorf("expected no ollama suggestions, got %v", got)
	}
}

func TestReadAnthropicStreamError(t *testing.T) {
	stream := strings.Join([]string{
		`event: error`,
		`data: {"type":"error","error":{"type":"overloaded_error","message":"overloaded"}}`,
		``,
	}, "\n")

	_, err := readAnthropicStream(strings.NewReader(stream))
	if err == nil || !strings.Contains(err.Error(), "overloaded") {
		t.Fatalf("expected an overloaded error, got %v", err)
	}
}
