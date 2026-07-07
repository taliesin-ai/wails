package llm

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed models.json
var modelsJSON []byte

// rank is a ranked recommendation with rationale for one language.
type rank struct {
	Rank      []string `json:"rank"`
	Rationale string   `json:"rationale"`
}

// ModelMap is the config-driven per-language recommendation map (SPEC.md 9).
type ModelMap struct {
	Default           rank            `json:"default"`
	ByLanguage        map[string]rank `json:"byLanguage"`
	NativeReviewLangs []string        `json:"nativeReviewLangs"`
	LocalPrivacyNote  string          `json:"localPrivacyNote"`
}

// ModelSuggestion is one ranked model returned to the UI.
type ModelSuggestion struct {
	Model string `json:"model"`
	Rank  int    `json:"rank"`
}

// Recommendation is the full per-language recommendation payload.
type Recommendation struct {
	Suggestions       []ModelSuggestion `json:"suggestions"`
	Rationale         string            `json:"rationale"`
	NeedsNativeReview bool              `json:"needsNativeReview"`
	PrivacyNote       string            `json:"privacyNote"`
}

func loadModelMap() ModelMap {
	var m ModelMap
	_ = json.Unmarshal(modelsJSON, &m)
	return m
}

// Recommend resolves the ranking for a target language, falling back to the
// default ranking. Matching is case-insensitive and tolerant of region
// subtags (e.g. "zh-CN" matches "zh-cn", "pt-BR" matches "pt").
func (m ModelMap) Recommend(targetLang string) Recommendation {
	key := strings.ToLower(targetLang)
	r, ok := m.ByLanguage[key]
	if !ok {
		// try base subtag
		if base := strings.SplitN(key, "-", 2)[0]; base != key {
			r, ok = m.ByLanguage[base]
		}
	}
	if !ok {
		r = m.Default
	}
	out := Recommendation{Rationale: r.Rationale, PrivacyNote: m.LocalPrivacyNote}
	for i, model := range r.Rank {
		out.Suggestions = append(out.Suggestions, ModelSuggestion{Model: model, Rank: i + 1})
	}
	for _, l := range m.NativeReviewLangs {
		if strings.EqualFold(l, key) || strings.EqualFold(l, strings.SplitN(key, "-", 2)[0]) {
			out.NeedsNativeReview = true
		}
	}
	return out
}
