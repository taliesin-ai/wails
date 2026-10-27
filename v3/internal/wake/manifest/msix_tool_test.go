package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMSIXToolDecodesValidatesAndEncodesRoundTrip(t *testing.T) {
	project := Project{Name: "msix-tool", ProductName: "MSIX Tool", Identifier: "com.example.msix-tool", Version: "1.0.0"}

	for _, tool := range []string{MSIXToolMakeAppx, MSIXToolPackagingTool} {
		t.Run(tool, func(t *testing.T) {
			dir := t.TempDir()
			doc := defaults(project)
			require.NoError(t, applyPackage(&doc.Package, PackageFormat{Format: "msix", Publisher: "CN=Publisher", Tool: tool}))
			encoded, err := EncodeConfig(configFromDocument(dir, "", doc))
			require.NoError(t, err)
			assert.Contains(t, string(encoded), "tool: "+tool)

			require.NoError(t, os.WriteFile(filepath.Join(dir, Filename), encoded, 0o644))
			loaded, err := Load(dir, "")
			require.NoError(t, err)
			assert.Equal(t, tool, loaded.Config.Package.Windows.MSIX.Tool)
			resolved, err := ResolvePackageFormat(loaded.Config.Package, "windows", "msix")
			require.NoError(t, err)
			assert.Equal(t, tool, resolved.Tool)
		})
	}

	t.Run("omitted tool stays unset", func(t *testing.T) {
		doc := defaults(project)
		require.NoError(t, applyPackage(&doc.Package, PackageFormat{Format: "msix", Publisher: "CN=Publisher"}))
		encoded, err := EncodeConfig(configFromDocument(t.TempDir(), "", doc))
		require.NoError(t, err)
		assert.NotContains(t, string(encoded), "tool:")
	})
}

func TestMSIXToolRejectsUnknownValues(t *testing.T) {
	paths, err := newProjectPathValidator(t.TempDir())
	require.NoError(t, err)

	for _, tool := range []string{MSIXToolMakeAppx, MSIXToolPackagingTool, ""} {
		assert.NoError(t, validatePackageOptions(paths, "windows.msix", PackageFormat{Format: "msix", Tool: tool}), tool)
	}
	assert.Error(t, validatePackageOptions(paths, "windows.msix", PackageFormat{Format: "msix", Tool: "signtool"}))
	assert.Error(t, validatePackageOptions(paths, "darwin.dmg", PackageFormat{Format: "dmg", Tool: MSIXToolMakeAppx}))
}
