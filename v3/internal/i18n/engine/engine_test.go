package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/internal/i18n/llm"
)

// fakeProvider records each call's Text and echoes it back uppercased so we can
// assert on chunk boundaries.
type fakeProvider struct{ calls []string }

func (f *fakeProvider) Name() string { return "fake" }
func (f *fakeProvider) Translate(_ context.Context, in llm.TranslateInput) (string, error) {
	f.calls = append(f.calls, in.Text)
	return strings.ToUpper(in.Text), nil
}

func TestTranslateMaskedWhole(t *testing.T) {
	fp := &fakeProvider{}
	out, err := translateMasked(context.Background(), fp, "small text", llm.TranslateInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(fp.calls) != 1 {
		t.Fatalf("expected 1 call for small input, got %d", len(fp.calls))
	}
	if out != "SMALL TEXT" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestTranslateMaskedChunks(t *testing.T) {
	// Three paragraphs each near the budget force splitting onto separate calls.
	para := strings.Repeat("word ", maxChunkChars/5)
	body := para + "\n\n" + para + "\n\n" + para
	fp := &fakeProvider{}
	out, err := translateMasked(context.Background(), fp, body, llm.TranslateInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(fp.calls) < 2 {
		t.Fatalf("expected the oversized body to be chunked, got %d call(s)", len(fp.calls))
	}
	if out != strings.ToUpper(body) {
		t.Errorf("rejoined output does not match: len got %d want %d", len(out), len(strings.ToUpper(body)))
	}
}

func TestSplitChunksRoundTrips(t *testing.T) {
	body := "aaaa\n\nbbbb\n\ncccc"
	chunks := splitChunks(body, 6) // each 4-char para fits; pairs exceed 6, so no merging
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d: %q", len(chunks), chunks)
	}
	if rejoined := strings.Join(chunks, "\n\n"); rejoined != body {
		t.Errorf("rejoined %q != original %q", rejoined, body)
	}
}

func TestProtectRestoreRoundTrip(t *testing.T) {
	src := "Install with `npm install`.\n\n" +
		"```go\nfmt.Println(\"hi\")\n```\n\n" +
		"import Foo from './Foo'\n\n" +
		"Use the <Button label=\"Go\"/> component to continue."

	masked, tokens := Protect(src)
	if strings.Contains(masked, "npm install") {
		t.Errorf("inline code not masked: %q", masked)
	}
	if strings.Contains(masked, "fmt.Println") {
		t.Errorf("fenced code not masked: %q", masked)
	}
	if strings.Contains(masked, "import Foo") {
		t.Errorf("import not masked: %q", masked)
	}
	if !strings.Contains(masked, "<") == false && strings.Contains(masked, "Button label") {
		t.Errorf("JSX tag not masked: %q", masked)
	}
	if got := Restore(masked, tokens); got != src {
		t.Errorf("round trip mismatch:\n got: %q\nwant: %q", got, src)
	}
}

func TestProtectMaskCount(t *testing.T) {
	masked, tokens := Protect("a `b` c `d` e")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d (%q)", len(tokens), masked)
	}
	if Restore(masked, tokens) != "a `b` c `d` e" {
		t.Errorf("restore failed: %q", Restore(masked, tokens))
	}
}

func TestSegment(t *testing.T) {
	body := "# Title\n\nA paragraph.\n\n```go\ncode\n```\n\n- item one\n- item two"
	blocks := Segment(body)
	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d: %+v", len(blocks), blocks)
	}
	want := []string{"heading", "paragraph", "code", "list"}
	for i, w := range want {
		if blocks[i].Kind != w {
			t.Errorf("block %d kind = %q, want %q", i, blocks[i].Kind, w)
		}
	}
}

func TestValidateStructure(t *testing.T) {
	src := "# H\n\n```go\nx\n```\n"
	ok := "# Überschrift\n\n```go\nx\n```\n"
	if w := ValidateStructure(src, ok); len(w) != 0 {
		t.Errorf("expected no warnings, got %v", w)
	}
	broken := "# Überschrift\n\nlost the code fence\n"
	if w := ValidateStructure(src, broken); len(w) == 0 {
		t.Errorf("expected a warning for a dropped code fence")
	}
}

func TestValidateStructureWarnsWhenStepsOrderedListIndentationIsLost(t *testing.T) {
	// Given
	src := strings.Join([]string{
		"<Steps>",
		"",
		"1. **Install dependencies**",
		"",
		"   Run the install command:",
		"",
		"   ```bash",
		"   npm install",
		"   ```",
		"",
		"2. **Start the app**",
		"",
		"   Run the dev server.",
		"",
		"</Steps>",
	}, "\n")
	preserved := strings.ReplaceAll(src, "Run", "Execute")
	broken := strings.Join([]string{
		"<Steps>",
		"",
		"1. **Install dependencies**",
		"",
		"Run the install command:",
		"",
		"```bash",
		"npm install",
		"```",
		"",
		"2. **Start the app**",
		"",
		"   Run the dev server.",
		"",
		"</Steps>",
	}, "\n")

	// When
	preservedWarnings := ValidateStructure(src, preserved)
	brokenWarnings := ValidateStructure(src, broken)

	// Then
	if len(preservedWarnings) != 0 {
		t.Fatalf("expected preserved indentation to pass, got %v", preservedWarnings)
	}
	if !hasWarning(brokenWarnings, "Steps ordered-list indentation") {
		t.Fatalf("expected Steps indentation warning, got %v", brokenWarnings)
	}
}

func hasWarning(warnings []string, needle string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, needle) {
			return true
		}
	}
	return false
}

func TestAlignDiff(t *testing.T) {
	left := []Block{{Kind: "heading", Text: "# A"}, {Kind: "paragraph", Text: "same"}}
	right := []Block{{Kind: "heading", Text: "# B"}, {Kind: "paragraph", Text: "same"}}
	d := AlignDiff(left, right)
	if len(d.Rows) != 2 {
		t.Fatalf("rows = %d", len(d.Rows))
	}
	if !d.Rows[0].Changed {
		t.Errorf("row 0 should be changed")
	}
	if d.Rows[1].Changed {
		t.Errorf("row 1 should be unchanged")
	}
}

func TestTranslateFrontmatter(t *testing.T) {
	fm := "title: Hello\ndescription: A page\nslug: keep-me\nsidebar:\n  order: 3"
	out := translateFrontmatter(fm, func(s string) string { return "[" + s + "]" })
	if !strings.Contains(out, "title: [Hello]") {
		t.Errorf("title not translated: %q", out)
	}
	if !strings.Contains(out, "description: [A page]") {
		t.Errorf("description not translated: %q", out)
	}
	if !strings.Contains(out, "slug: keep-me") {
		t.Errorf("slug should be untouched: %q", out)
	}
}
