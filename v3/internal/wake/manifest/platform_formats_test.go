package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const platformFormatsBase = `version: 3
project:
  name: formats
  product_name: Formats
  identifier: com.example.formats
  version: 1.0.0
frontend:
  directory: frontend
  install: [npm, install]
  build: [npm, run, build]
  dev: [npm, run, dev]
  output: frontend/dist
`

func TestPlatformFormatsDecodeAndEncodeOnlyWhenOverridden(t *testing.T) {
	root := t.TempDir()
	writeManifest(t, root, platformFormatsBase+"linux:\n  formats: [deb, rpm]\n")
	loaded, err := Load(root, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"deb", "rpm"}, loaded.Config.Package.Linux.Formats)
	assert.Equal(t, []string{"dmg"}, loaded.Config.Package.Darwin.Formats, "unset platforms keep their defaults")

	encoded, err := EncodeConfig(loaded.Config)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), "formats: [deb, rpm]")

	defaults := t.TempDir()
	writeManifest(t, defaults, platformFormatsBase)
	loaded, err = Load(defaults, "")
	require.NoError(t, err)
	encoded, err = EncodeConfig(loaded.Config)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "formats:", "default platform formats must not be written")
}

func TestPlatformFormatsRejectFormatsOutsideThePlatform(t *testing.T) {
	for _, test := range []struct {
		name     string
		section  string
		contents string
		field    string
		allowed  string
	}{
		{name: "linux rejects nsis", section: "linux", contents: "[nsis]", field: "linux.formats", allowed: "appimage, deb, rpm, archlinux"},
		{name: "windows rejects deb", section: "windows", contents: "[deb]", field: "windows.formats", allowed: "nsis, msix"},
		{name: "linux rejects duplicates", section: "linux", contents: "[deb, deb]", field: "linux.formats"},
		{name: "unknown format", section: "linux", contents: "[zip]", field: "linux.formats", allowed: "appimage, deb, rpm, archlinux"},
		{name: "android rejects deb", section: "android", contents: "[deb]", field: "android.formats", allowed: "aab, apk"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeManifest(t, root, platformFormatsBase+test.section+":\n  formats: "+test.contents+"\n")
			_, err := Load(root, "")
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.field)
			if test.allowed != "" {
				assert.Contains(t, err.Error(), "allowed values: "+test.allowed)
			}
		})
	}
}

func TestPlatformFormatsAcceptAndroidProductionFormats(t *testing.T) {
	root := t.TempDir()
	writeManifest(t, root, platformFormatsBase+"android:\n  formats: [apk, aab]\n")
	loaded, err := Load(root, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"apk", "aab"}, loaded.Config.Package.Android.Formats)
}
