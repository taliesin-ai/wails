package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const appxFixture = `<?xml version="1.0" encoding="utf-8"?>
<Package xmlns="http://schemas.microsoft.com/appx/manifest/foundation/windows10">
  <Identity Name="com.example.app" Publisher="CN=My Company" Version="1.0.0.0" ProcessorArchitecture="x64"/>
</Package>
`

func TestMSIXManifestPublisherReportsValueAndLine(t *testing.T) {
	publisher, line, found, err := MSIXManifestPublisher([]byte(appxFixture))
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "CN=My Company", publisher)
	assert.Equal(t, 3, line)
}

func TestMSIXManifestPublisherWithoutAttribute(t *testing.T) {
	_, _, found, err := MSIXManifestPublisher([]byte(`<Package><Identity Name="x"/></Package>`))
	require.NoError(t, err)
	assert.False(t, found)
}

func TestMSIXManifestPublisherRejectsMalformedXML(t *testing.T) {
	_, _, _, err := MSIXManifestPublisher([]byte(`<Package><Identity`))
	assert.Error(t, err)
}

func TestMSIXPublisherMatchesSubject(t *testing.T) {
	assert.True(t, MSIXPublisherMatchesSubject("CN=Wails Wake Test", "CN=Wails Wake Test"))
	assert.True(t, MSIXPublisherMatchesSubject("CN=Example, O=Acme, C=AU", "C=AU,O=Acme,CN=Example"), "RDN order is ignored")
	assert.True(t, MSIXPublisherMatchesSubject("cn=Example", "CN=Example"), "attribute keys are case-insensitive")
	assert.False(t, MSIXPublisherMatchesSubject("CN=My Company", "CN=Wails Wake Test"))
	assert.False(t, MSIXPublisherMatchesSubject("CN=Example", "CN=Example, O=Acme"), "extra subject component")
	assert.False(t, MSIXPublisherMatchesSubject("", "CN=Example"))
}

func TestPackageValidationRejectsPublisherThatDisagreesWithUserManifest(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "app_manifest.xml"), []byte(appxFixture), 0o644))
	paths, err := newProjectPathValidator(root)
	require.NoError(t, err)

	err = validatePackageOptions(paths, "windows.msix", PackageFormat{Format: "msix", Publisher: "CN=Wails Wake Test", Manifest: "app_manifest.xml"})
	require.Error(t, err)
	assert.ErrorContains(t, err, `"CN=Wails Wake Test"`)
	assert.ErrorContains(t, err, `app_manifest.xml line 3`)
	assert.ErrorContains(t, err, `"CN=My Company"`)

	assert.NoError(t, validatePackageOptions(paths, "windows.msix", PackageFormat{Format: "msix", Publisher: "CN=My Company", Manifest: "app_manifest.xml"}))
	assert.NoError(t, validatePackageOptions(paths, "windows.msix", PackageFormat{Format: "msix", Manifest: "app_manifest.xml"}), "no publisher setting means nothing to compare")
}
