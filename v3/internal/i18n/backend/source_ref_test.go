package backend

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSourceRefAcceptsSourceRelativePath(t *testing.T) {
	// Given
	raw := "guide/a.mdx"

	// When
	ref, err := ParseSourceRef(raw)

	// Then
	require.NoError(t, err)
	require.Equal(t, SourceRef{Path: raw}, ref)
}

func TestParseSourceRefRejectsUnsafePaths(t *testing.T) {
	tests := []string{
		"",
		"/abs.mdx",
		"../secret.mdx",
		"guide/../secret.mdx",
		`guide\secret.mdx`,
		"guide//a.mdx",
		"./guide/a.mdx",
		"guide/./a.mdx",
		"C:/secret.mdx",
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			// When
			_, err := ParseSourceRef(raw)

			// Then
			require.ErrorContains(t, err, "invalid source-relative path")
		})
	}
}
