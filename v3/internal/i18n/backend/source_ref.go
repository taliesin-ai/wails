package backend

import (
	"fmt"
	"path"
	"strings"
)

// ParseSourceRef parses untrusted CLI/API input into a source-relative docs path.
func ParseSourceRef(raw string) (SourceRef, error) {
	if raw == "" || strings.Contains(raw, `\`) || path.IsAbs(raw) || isWindowsDrivePath(raw) || path.Clean(raw) != raw {
		return SourceRef{}, fmt.Errorf("invalid source-relative path %q", raw)
	}
	for _, part := range strings.Split(raw, "/") {
		if part == ".." || part == "." || part == "" {
			return SourceRef{}, fmt.Errorf("invalid source-relative path %q", raw)
		}
	}
	return SourceRef{Path: raw}, nil
}

func isWindowsDrivePath(raw string) bool {
	return len(raw) >= 2 && raw[1] == ':'
}
