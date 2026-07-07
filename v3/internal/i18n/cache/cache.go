// Package cache reads and writes the committed translation cache
// (docs/.translation-cache/<locale>.json). It owns the hash scheme used to
// detect staleness and the on-disk JSON format.
//
// Hash recipe (documented deviation): the original docs-translation helper that
// produced the historical cache values lives on an unmerged branch and its
// recipe could not be reproduced from the source tree, so this package defines
// its own clear, reproducible recipe: the lowercase hex of sha256(sourceBytes)
// truncated to 12 characters. The tool stamps and reads with this single recipe,
// so everything it produces is internally consistent. Pre-existing cache entries
// whose hashes were produced by the old recipe (or whose English source has
// since changed) simply read as "stale", which is the correct, honest state.
//
// See SPEC.md section 4.
package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Entry is one source-path -> translation record in a committed cache file.
// Unknown keys are preserved by older readers; we only add keys that legitimately
// travel with the repo (hash, snapshot, timestamp).
type Entry struct {
	Hash           string `json:"hash"`                      // sha256[:12] of the source at translation time
	SourceSnapshot string `json:"source_snapshot,omitempty"` // path (relative to cache dir) of the snapshot of the source as translated
	TranslatedAt   string `json:"translated_at,omitempty"`   // RFC3339 timestamp
}

// Cache maps a source-relative path (e.g. "quick-start/why-wails.mdx") to its
// translation record for a single locale.
type Cache map[string]Entry

// Hash returns the staleness hash for the given source bytes: sha256[:12].
func Hash(content []byte) string {
	sum := sha256.Sum256(content)
	return fmt.Sprintf("%x", sum)[:12]
}

// FilePath returns the committed cache file path for a locale within cacheDir.
func FilePath(cacheDir, locale string) string {
	return filepath.Join(cacheDir, locale+".json")
}

// Load reads the cache file for a locale. A missing file is not an error; it
// returns an empty cache so a brand-new locale behaves as "everything missing".
func Load(cacheDir, locale string) (Cache, error) {
	path := FilePath(cacheDir, locale)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Cache{}, nil
		}
		return nil, err
	}
	c := Cache{}
	if len(data) == 0 {
		return c, nil
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse cache %s: %w", path, err)
	}
	return c, nil
}

// Save writes the cache file for a locale with stable key ordering so diffs stay
// small. It creates the cache directory if needed.
func (c Cache) Save(cacheDir, locale string) error {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}
	// Marshal with sorted keys for deterministic output.
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf []byte
	buf = append(buf, '{', '\n')
	for i, k := range keys {
		entryJSON, err := json.MarshalIndent(c[k], "  ", "  ")
		if err != nil {
			return err
		}
		keyJSON, _ := json.Marshal(k)
		buf = append(buf, ' ', ' ')
		buf = append(buf, keyJSON...)
		buf = append(buf, ':', ' ')
		buf = append(buf, entryJSON...)
		if i < len(keys)-1 {
			buf = append(buf, ',')
		}
		buf = append(buf, '\n')
	}
	buf = append(buf, '}', '\n')

	return os.WriteFile(FilePath(cacheDir, locale), buf, 0o644)
}

// SnapshotDir is the subdirectory under the cache dir holding source snapshots.
const SnapshotDir = ".snapshots"

// SnapshotPath returns the cache-dir-relative snapshot path for a source path.
func SnapshotPath(srcRelPath string) string {
	return filepath.Join(SnapshotDir, srcRelPath)
}

// WriteSnapshot stores a copy of the source content so a later source-vs-source
// diff (SPEC.md section 8) can show exactly which English blocks changed.
func WriteSnapshot(cacheDir, srcRelPath string, content []byte) error {
	full := filepath.Join(cacheDir, SnapshotPath(srcRelPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, content, 0o644)
}

// ReadSnapshot loads a previously written source snapshot, if present.
func ReadSnapshot(cacheDir, snapshotRelPath string) ([]byte, error) {
	return os.ReadFile(filepath.Join(cacheDir, snapshotRelPath))
}
