package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

const platformFormatsPipelineBase = `version: 3
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

func platformFormatsPlan(t *testing.T, contents, profile string, request Request) []string {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, manifest.Filename), []byte(contents), 0o644))
	loaded, err := manifest.Load(root, profile)
	require.NoError(t, err)
	plan, err := PlanBuild(loaded.Config, request)
	require.NoError(t, err)
	require.Len(t, plan.Intent.Targets, 1)
	return plan.Intent.Targets[0].Formats
}

func TestPlatformDefaultFormatsApplyToAnonymousPackageRequests(t *testing.T) {
	contents := platformFormatsPipelineBase + "linux:\n  formats: [deb, rpm]\n"
	formats := platformFormatsPlan(t, contents, "", Request{Verb: "package", TargetOS: "linux", TargetArch: "amd64"})
	assert.Equal(t, []string{"deb", "rpm"}, formats)
}

func TestBuiltInDefaultFormatsApplyWhenNothingIsConfigured(t *testing.T) {
	formats := platformFormatsPlan(t, platformFormatsPipelineBase, "", Request{Verb: "package", TargetOS: "linux", TargetArch: "amd64"})
	assert.Equal(t, []string{"appimage"}, formats)
}

func TestCLIFormatsOverridePlatformDefaults(t *testing.T) {
	contents := platformFormatsPipelineBase + "linux:\n  formats: [deb, rpm]\n"
	formats := platformFormatsPlan(t, contents, "", Request{Verb: "package", TargetOS: "linux", TargetArch: "amd64", Formats: []string{"archlinux"}})
	assert.Equal(t, []string{"archlinux"}, formats)
}

func TestProfileFormatsOverridePlatformDefaults(t *testing.T) {
	contents := platformFormatsPipelineBase + `linux:
  formats: [deb, rpm]
profiles:
  release:
    targets:
      linux/amd64:
        formats: [appimage]
`
	formats := platformFormatsPlan(t, contents, "release", Request{Verb: "package"})
	assert.Equal(t, []string{"appimage"}, formats)
}

func TestProfileTargetWithoutFormatsUsesPlatformDefaultOnlyForPackaging(t *testing.T) {
	contents := platformFormatsPipelineBase + `linux:
  formats: [deb, rpm]
profiles:
  release:
    targets:
      linux/amd64:
        sign: true
`
	assert.Equal(t, []string{"deb", "rpm"}, platformFormatsPlan(t, contents, "release", Request{Verb: "package"}))
	assert.Empty(t, platformFormatsPlan(t, contents, "release", Request{Verb: "build"}), "plain builds must not start producing packages")
}
