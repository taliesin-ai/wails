package commands

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

// msixPackageNode is a PackageArtifact node for an MSIX package using tool.
func msixPackageNode(tool string) pipeline.Node {
	return pipeline.Node{Kind: pipeline.PackageArtifact, Spec: pipeline.PackageSpec{
		Format: "msix", Config: manifest.PackageFormat{Tool: tool},
	}}
}

// putFakeExecutable places an executable named name on a PATH directory.
func putFakeExecutable(t *testing.T, directory, name string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(directory, name), []byte("#!/bin/sh\n"), 0o755))
}

func TestMSIXIdentityLooksUpTheSelectedPackager(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake executables are POSIX scripts")
	}
	handler := &manifestHandler{root: t.TempDir()}
	for _, test := range []struct {
		name      string
		tool      string
		available string
		wantErr   string
	}{
		{name: "default selects MakeAppx and fails without it", tool: "", available: "MsixPackagingTool.exe", wantErr: "MakeAppx.exe"},
		{name: "makeappx selects MakeAppx", tool: manifest.MSIXToolMakeAppx, available: "MsixPackagingTool.exe", wantErr: "MakeAppx.exe"},
		{name: "packaging tool selects MsixPackagingTool and fails without it", tool: manifest.MSIXToolPackagingTool, available: "MakeAppx.exe", wantErr: "MsixPackagingTool.exe"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir()
			putFakeExecutable(t, path, test.available)
			t.Setenv("PATH", path)
			_, err := handler.Identity(context.Background(), msixPackageNode(test.tool))
			require.ErrorContains(t, err, test.wantErr)
		})
	}

	t.Run("packaging tool identity resolves when MsixPackagingTool is installed", func(t *testing.T) {
		path := t.TempDir()
		putFakeExecutable(t, path, "MsixPackagingTool.exe")
		t.Setenv("PATH", path)
		_, err := handler.Identity(context.Background(), msixPackageNode(manifest.MSIXToolPackagingTool))
		assert.NoError(t, err)
	})

	t.Run("makeappx identity resolves when MakeAppx is installed", func(t *testing.T) {
		path := t.TempDir()
		putFakeExecutable(t, path, "MakeAppx.exe")
		t.Setenv("PATH", path)
		_, err := handler.Identity(context.Background(), msixPackageNode(manifest.MSIXToolMakeAppx))
		assert.NoError(t, err)
	})
}
