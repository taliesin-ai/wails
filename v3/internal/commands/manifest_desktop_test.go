package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinuxDesktopTemplateRendersWithoutWhitespaceOnlyLines(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, UpdateBuildAssets(&UpdateBuildAssetsOptions{Dir: dir, Name: "app", BinaryName: "app", ProductName: "App", ProductDescription: "Example", ProductVersion: "1.0.0", ProductIdentifier: "com.example.app", Silent: true}))
	data, err := os.ReadFile(filepath.Join(dir, "linux", "desktop"))
	require.NoError(t, err)
	rendered := string(data)
	assert.True(t, strings.HasSuffix(rendered, "\n"), "desktop entry must end with a newline")
	assert.False(t, strings.HasSuffix(rendered, "\n\n"), "desktop entry must end with exactly one newline")
	for index, line := range strings.Split(rendered, "\n") {
		if line == "" {
			continue
		}
		assert.Equal(t, strings.TrimRight(line, " \t\r"), line, "line %d has trailing whitespace", index+1)
		assert.NotEmpty(t, strings.TrimSpace(line), "line %d is whitespace-only", index+1)
	}
}

func TestNormalizeDesktopEntryDropsWhitespaceOnlyLines(t *testing.T) {
	input := "[Desktop Entry]  \n   \nName=App\t\n\t\nExec=app %u\n  \n"
	assert.Equal(t, "[Desktop Entry]\nName=App\nExec=app %u\n", string(normalizeDesktopEntry([]byte(input))))
	assert.Nil(t, normalizeDesktopEntry([]byte(" \n\t\n")))
}

func TestAppImageDesktopEntryCopyIsNormalizedAndUserFileUntouched(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".wails", "assets"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".wails", "assets", "appicon.png"), []byte("png"), 0o644))
	userDesktop := "[Desktop Entry]\nName=App\nExec=app\n \n"
	require.NoError(t, os.MkdirAll(filepath.Join(root, "packaging"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "packaging", "app.desktop"), []byte(userDesktop), 0o644))
	spec := packageTestSpec("linux", "amd64", "appimage")
	spec.Assets = ".wails/assets"
	spec.Output = "bin/app.AppImage"
	spec.Config.DesktopEntry = "packaging/app.desktop"
	handler := &manifestHandler{root: root}
	desktop, _, _, err := handler.prepareAppImageInputs(spec)
	require.NoError(t, err)
	copied, err := os.ReadFile(desktop)
	require.NoError(t, err)
	assert.Equal(t, "[Desktop Entry]\nName=App\nExec=app\n", string(copied))
	original, err := os.ReadFile(filepath.Join(root, "packaging", "app.desktop"))
	require.NoError(t, err)
	assert.Equal(t, userDesktop, string(original), "the user's desktop entry must not be modified")
}
