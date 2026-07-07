package starlight

import (
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
)

const sampleConfig = `
export default defineConfig({
  integrations: [starlight({
      locales: {
        root:    { label: "English",  lang: "en",    dir: "ltr" },
        "zh-cn": { label: "简体中文",  lang: "zh-CN", dir: "ltr" },
        ja:      { label: "日本語",    lang: "ja",    dir: "ltr" },
      },
  })],
})`

func TestParseLocales(t *testing.T) {
	locs := parseLocales(sampleConfig)
	if len(locs) != 3 {
		t.Fatalf("expected 3 locales, got %d: %+v", len(locs), locs)
	}
	if !locs[0].IsSource || locs[0].Code != "en" {
		t.Errorf("root should map to source locale en, got %+v", locs[0])
	}
	if locs[1].Code != "zh-cn" || locs[1].Lang != "zh-CN" {
		t.Errorf("quoted locale parsed wrong: %+v", locs[1])
	}
	if locs[2].Code != "ja" {
		t.Errorf("unquoted locale parsed wrong: %+v", locs[2])
	}
}

func TestSplitJoinDoc(t *testing.T) {
	raw := "---\ntitle: Hi\n---\nBody line.\n"
	d := SplitDoc(raw)
	if d.Frontmatter != "title: Hi" {
		t.Errorf("frontmatter = %q", d.Frontmatter)
	}
	if d.Body != "Body line.\n" {
		t.Errorf("body = %q", d.Body)
	}
	if got := joinDoc(backend.Doc{Frontmatter: d.Frontmatter, Body: d.Body}); got != raw {
		t.Errorf("join mismatch: %q != %q", got, raw)
	}
}

func TestSplitDocNoFrontmatter(t *testing.T) {
	raw := "Just body.\n"
	d := SplitDoc(raw)
	if d.Frontmatter != "" || d.Body != raw {
		t.Errorf("unexpected split: fm=%q body=%q", d.Frontmatter, d.Body)
	}
}

func TestRewriteInternalLinks(t *testing.T) {
	s := New("/tmp/docs")
	loc := backend.Locale{Code: "id", Lang: "id"}
	in := backend.Doc{Body: "See [start](/quick-start/installation) and [ext](https://x.io) and [done](/id/already)."}
	out := s.RewriteInternalLinks(in, loc)
	if !strings.Contains(out.Body, "](/id/quick-start/installation)") {
		t.Errorf("internal link not prefixed: %q", out.Body)
	}
	if !strings.Contains(out.Body, "](https://x.io)") {
		t.Errorf("external link should be untouched: %q", out.Body)
	}
	if strings.Contains(out.Body, "/id/id/already") {
		t.Errorf("already-prefixed link double-prefixed: %q", out.Body)
	}
}

func TestRewriteSkipsSourceLocale(t *testing.T) {
	s := New("/tmp/docs")
	loc := backend.Locale{Code: "en", IsSource: true}
	in := backend.Doc{Body: "[x](/a)"}
	if got := s.RewriteInternalLinks(in, loc); got.Body != in.Body {
		t.Errorf("source locale should not be rewritten: %q", got.Body)
	}
}
