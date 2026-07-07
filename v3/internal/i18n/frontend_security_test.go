package i18n

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestFrontendSecurityUnsafePatterns(t *testing.T) {
	htmlPath := filepath.Join("frontend", "dist", "index.html")
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", htmlPath, err)
	}

	tests := []struct {
		name    string
		pattern string
		errMsg  string
	}{
		{
			name:    "Unsafe file path in tree",
			pattern: `(?m)<span class="path">\$\{f\.path\}<\/span>`,
			errMsg:  "Found unsafe interpolation of f.path. Use escapeHtml(f.path)",
		},
		{
			name:    "Unsafe file path in editor header",
			pattern: `(?m)<strong[^>]*>\$\{path\}<\/strong>`,
			errMsg:  "Found unsafe interpolation of path in editor header. Use escapeHtml(path)",
		},
		{
			name:    "Unsafe file-load error",
			pattern: `(?m)Failed to load file: \$\{e\.message\|\|e\}`,
			errMsg:  "Found unsafe interpolation of file load error. Use escapeHtml(e.message||e)",
		},
		{
			name:    "Unsafe translation event data (file_done)",
			pattern: `(?m)log\.innerHTML\+=` + "`" + `<div>✓ \$\{ev\.path\}`,
			errMsg:  "Found unsafe interpolation of ev.path in log.innerHTML",
		},
		{
			name:    "Unsafe translation event error",
			pattern: `(?m)log\.innerHTML\+=` + "`" + `<div[^>]*>✗ \$\{ev\.path\}: \$\{ev\.error\}<\/div>` + "`",
			errMsg:  "Found unsafe interpolation of ev.error/path in log.innerHTML",
		},
		{
			name:    "Unsafe model option",
			pattern: `(?m)<option value="\$\{m\}">\$\{m\}<\/option>`,
			errMsg:  "Found unsafe interpolation of model name in <option>",
		},
		{
			name:    "Unsafe recommendation rationale",
			pattern: `(?m)note\.innerHTML=.*— \$\{rec\.rationale\}` + "`",
			errMsg:  "Found unsafe interpolation of rec.rationale in note.innerHTML",
		},
		{
			name:    "Unsafe submit changedFiles",
			pattern: `(?m)changedFiles\|\|\[\]\)\.map\(f=>` + "`" + `<div>\$\{f\}<\/div>` + "`" + `\)`,
			errMsg:  "Found unsafe interpolation of changedFiles in submit",
		},
		{
			name:    "Unsafe submit commands",
			pattern: `(?m)commands\|\|\[\]\)\.map\(c=>` + "`" + `<div>\$ \$\{c\}<\/div>` + "`" + `\)`,
			errMsg:  "Found unsafe interpolation of commands in submit",
		},
		{
			name:    "Unsafe submit message",
			pattern: `(?m)<p class="muted">\$\{r\.message\|\|''\}<\/p>`,
			errMsg:  "Found unsafe interpolation of r.message in submit",
		},
		{
			name:    "Unsafe navItem label/mini",
			pattern: `(?m)<span class="lbl">\$\{label\}<\/span>`,
			errMsg:  "Found unsafe interpolation of label in navItem",
		},
		{
			name:    "Unsafe navItem mini",
			pattern: `(?m)<span class="mini">\$\{mini\}<\/span>`,
			errMsg:  "Found unsafe interpolation of mini in navItem",
		},
		{
			name:    "Unsafe backend root",
			pattern: `(?m)<code>\$\{backend\.root\}<\/code>`,
			errMsg:  "Found unsafe interpolation of backend.root in renderStatus",
		},
		{
			name:    "Unsafe dashboard load error",
			pattern: `(?m)Couldn't reach the translation server \(\$\{e\.message\|\|e\}\)`,
			errMsg:  "Found unsafe interpolation of dashboard load error",
		},
		{
			name:    "Unsafe localeRow label/code/lang",
			pattern: `(?m)<td class="lang"><div class="name">\$\{l\.label\}<\/div><div class="code">\$\{l\.code\} · \$\{l\.lang\}<\/div><\/td>`,
			errMsg:  "Found unsafe interpolation of locale fields in localeRow",
		},
		{
			name:    "Unsafe freshnessBar tooltip",
			pattern: `(?m)data-tip="\$\{s\.n\} \$\{s\.label\.toLowerCase\(\)\} · \$\{Math\.round\(pc\(s\.n\)\)\}%"`,
			errMsg:  "Found unsafe interpolation of s.label in freshnessBar",
		},
		{
			name:    "Unsafe confirmTranslate summary",
			pattern: `(?m)<p class="muted"[^>]*>\$\{t\('translate\.summary',\{n,label:l\.label,model:cfg\.model\}\)\}<\/p>`,
			errMsg:  "Found unsafe interpolation of translate.summary in confirmTranslate",
		},
		{
			name:    "Unsafe runJob error message",
			pattern: `(?m)<p style="color:var\(--err\);line-height:1\.5">\$\{msg\}<\/p>`,
			errMsg:  "Found unsafe interpolation of msg in runJob error modal",
		},
		{
			name:    "Unsafe runJob code title",
			pattern: `(?m)<h2>\$\{t\('llm\.translate'\)\} · \$\{code\}<\/h2>`,
			errMsg:  "Found unsafe interpolation of code in runJob title",
		},
		{
			name:    "Unsafe dynamic inline handler (showWorkspace retry)",
			pattern: `(?m)onclick="showWorkspace\(` + "`?\\$" + `\{`,
			errMsg:  "Found unsafe dynamic inline handler for showWorkspace. Use DOM events or static ID bindings.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			re := regexp.MustCompile(tc.pattern)
			if re.Match(content) {
				t.Errorf("%s:\nMatched pattern: %s", tc.errMsg, tc.pattern)
			}
		})
	}
}
