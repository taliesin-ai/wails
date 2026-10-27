package manifest

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevPortAndBuildTidyDecodeAndRoundTrip(t *testing.T) {
	root := t.TempDir()
	writeManifest(t, root, minimalYAML+"\n  tidy: true\ndev:\n  port: 9333\n")
	loaded, err := Load(root, "")
	require.NoError(t, err)
	assert.Equal(t, 9333, loaded.Config.Dev.Port)
	assert.True(t, loaded.Config.Build.Tidy)

	encoded, err := EncodeConfig(loaded.Config)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), "port: 9333")
	assert.Contains(t, string(encoded), "tidy: true")

	second := t.TempDir()
	writeManifest(t, second, string(encoded))
	roundTrip, err := Load(second, "")
	require.NoError(t, err)
	assert.Equal(t, loaded.Config.Dev.Port, roundTrip.Config.Dev.Port)
	assert.Equal(t, loaded.Config.Build.Tidy, roundTrip.Config.Build.Tidy)
}

func TestDevPortAndBuildTidyDefaultsAreOmittedWhenNotExplicit(t *testing.T) {
	root := t.TempDir()
	writeManifest(t, root, minimalYAML)
	loaded, err := Load(root, "")
	require.NoError(t, err)
	assert.Equal(t, 9245, loaded.Config.Dev.Port)
	assert.False(t, loaded.Config.Build.Tidy)

	encoded, err := EncodeConfig(loaded.Config)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "tidy:")
	assert.NotContains(t, string(encoded), "port:")
}

func TestDevPortAndBuildTidyValidation(t *testing.T) {
	tests := []struct {
		name    string
		suffix  string
		message string
	}{
		{"port zero", "\ndev:\n  port: 0\n", "dev.port"},
		{"port above range", "\ndev:\n  port: 65536\n", "dev.port"},
		{"port negative", "\ndev:\n  port: -1\n", "dev.port"},
		{"port not an integer", "\ndev:\n  port: nine\n", "dev.port"},
		{"tidy not a boolean", "\n  tidy: maybe\n", "build.tidy"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeManifest(t, root, minimalYAML+test.suffix)
			_, err := Load(root, "")
			require.Error(t, err)
			assert.ErrorContains(t, err, test.message)
		})
	}
}

func TestDevPortAcceptsRangeBoundaries(t *testing.T) {
	for _, port := range []int{1, 65535} {
		root := t.TempDir()
		writeManifest(t, root, minimalYAML+"\ndev:\n  port: "+strconv.Itoa(port)+"\n")
		loaded, err := Load(root, "")
		require.NoError(t, err, port)
		assert.Equal(t, port, loaded.Config.Dev.Port)
	}
}
