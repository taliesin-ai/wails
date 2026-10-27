package pipeline

import (
	"errors"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errNotOnPath = errors.New("not on PATH")

func notOnPath(string) (string, error) { return "", errNotOnPath }

func sdkEnv(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

// fakeBuildToolsGlob matches patterns against a fixed set of files using the
// host's filepath rules, so fixtures built with filepath.Join line up.
func fakeBuildToolsGlob(t *testing.T, files ...string) func(string) ([]string, error) {
	return func(pattern string) ([]string, error) {
		var matches []string
		for _, file := range files {
			ok, err := filepath.Match(pattern, file)
			assert.NoError(t, err)
			if ok {
				matches = append(matches, file)
			}
		}
		sort.Strings(matches)
		return matches, nil
	}
}

func buildTool(sdk, version, name string) string {
	return filepath.Join(sdk, "build-tools", version, name)
}

func TestAndroidBuildToolPathPicksNumericallyNewestVersion(t *testing.T) {
	const sdk = "/sdk"
	files := []string{
		buildTool(sdk, "9.0.0", "apksigner"),
		buildTool(sdk, "34.0.0", "apksigner"),
		buildTool(sdk, "100.0.0", "apksigner"),
		buildTool(sdk, "36.0.0", "apksigner"),
	}
	path, ok := androidBuildToolPath("apksigner", "linux", sdkEnv(map[string]string{"ANDROID_HOME": sdk}), notOnPath, fakeBuildToolsGlob(t, files...))
	assert.True(t, ok)
	assert.Equal(t, buildTool(sdk, "100.0.0", "apksigner"), path)
}

func TestAndroidBuildToolPathPrefersFinalReleaseAndNumericRC(t *testing.T) {
	const sdk = "/sdk"
	env := sdkEnv(map[string]string{"ANDROID_HOME": sdk})

	final := []string{buildTool(sdk, "36.0.0-rc1", "zipalign"), buildTool(sdk, "36.0.0", "zipalign")}
	path, ok := androidBuildToolPath("zipalign", "linux", env, notOnPath, fakeBuildToolsGlob(t, final...))
	assert.True(t, ok)
	assert.Equal(t, buildTool(sdk, "36.0.0", "zipalign"), path, "final release outranks rc of the same version")

	rcs := []string{buildTool(sdk, "36.0.0-rc2", "zipalign"), buildTool(sdk, "36.0.0-rc10", "zipalign")}
	path, ok = androidBuildToolPath("zipalign", "linux", env, notOnPath, fakeBuildToolsGlob(t, rcs...))
	assert.True(t, ok)
	assert.Equal(t, buildTool(sdk, "36.0.0-rc10", "zipalign"), path, "rc10 outranks rc2 numerically")

	mixed := []string{buildTool(sdk, "35.0.0", "zipalign"), buildTool(sdk, "36.0.0-rc1", "zipalign")}
	path, ok = androidBuildToolPath("zipalign", "linux", env, notOnPath, fakeBuildToolsGlob(t, mixed...))
	assert.True(t, ok)
	assert.Equal(t, buildTool(sdk, "36.0.0-rc1", "zipalign"), path, "a newer rc beats an older final release")
}

func TestAndroidBuildToolPathSkipsUnparseableDirectories(t *testing.T) {
	const sdk = "/sdk"
	files := []string{
		buildTool(sdk, "preview", "apksigner"),
		buildTool(sdk, "36..0", "apksigner"),
		buildTool(sdk, "36.0.0-", "apksigner"),
		buildTool(sdk, "34.0.0", "apksigner"),
	}
	path, ok := androidBuildToolPath("apksigner", "linux", sdkEnv(map[string]string{"ANDROID_HOME": sdk}), notOnPath, fakeBuildToolsGlob(t, files...))
	assert.True(t, ok)
	assert.Equal(t, buildTool(sdk, "34.0.0", "apksigner"), path)
}

func TestAndroidBuildToolPathOnWindowsResolvesExecutableExtensions(t *testing.T) {
	// Forward slashes keep the fixture valid for filepath.Match on every host.
	const sdk = "C:/Android/sdk"
	env := sdkEnv(map[string]string{"ANDROID_HOME": sdk})

	zipalign := []string{buildTool(sdk, "36.0.0", "zipalign.exe")}
	path, ok := androidBuildToolPath("zipalign", "windows", env, notOnPath, fakeBuildToolsGlob(t, zipalign...))
	assert.True(t, ok, "zipalign.exe must resolve on Windows")
	assert.Equal(t, buildTool(sdk, "36.0.0", "zipalign.exe"), path)

	launcher := []string{buildTool(sdk, "34.0.0", "apksigner.bat"), buildTool(sdk, "36.0.0", "apksigner.bat")}
	path, ok = androidBuildToolPath("apksigner", "windows", env, notOnPath, fakeBuildToolsGlob(t, launcher...))
	assert.True(t, ok, "apksigner.bat must resolve on Windows")
	assert.Equal(t, buildTool(sdk, "36.0.0", "apksigner.bat"), path)

	sameVersion := []string{buildTool(sdk, "36.0.0", "apksigner.cmd"), buildTool(sdk, "36.0.0", "apksigner.exe"), buildTool(sdk, "36.0.0", "apksigner.bat")}
	path, ok = androidBuildToolPath("apksigner", "windows", env, notOnPath, fakeBuildToolsGlob(t, sameVersion...))
	assert.True(t, ok)
	assert.Equal(t, buildTool(sdk, "36.0.0", "apksigner.exe"), path, ".exe wins within one build-tools directory")

	path, ok = androidBuildToolPath("apksigner", "windows", env, func(string) (string, error) { return `C:\bin\apksigner.exe`, nil }, fakeBuildToolsGlob(t))
	assert.True(t, ok)
	assert.Equal(t, `C:\bin\apksigner.exe`, path, "PATH wins over the SDK on Windows too")
}

func TestAndroidBuildToolPathOffWindowsIgnoresWindowsLaunchers(t *testing.T) {
	const sdk = "/sdk"
	files := []string{buildTool(sdk, "36.0.0", "apksigner.bat"), buildTool(sdk, "36.0.0", "zipalign.exe")}
	env := sdkEnv(map[string]string{"ANDROID_HOME": sdk})
	_, ok := androidBuildToolPath("apksigner", "linux", env, notOnPath, fakeBuildToolsGlob(t, files...))
	assert.False(t, ok)
	_, ok = androidBuildToolPath("zipalign", "darwin", env, notOnPath, fakeBuildToolsGlob(t, files...))
	assert.False(t, ok)
}

func TestAndroidBuildToolPathFallsBackToSDKRootVariable(t *testing.T) {
	const sdk = "/opt/sdk"
	files := []string{buildTool(sdk, "36.0.0", "zipalign")}
	path, ok := androidBuildToolPath("zipalign", "linux", sdkEnv(map[string]string{"ANDROID_SDK_ROOT": sdk}), notOnPath, fakeBuildToolsGlob(t, files...))
	assert.True(t, ok)
	assert.Equal(t, buildTool(sdk, "36.0.0", "zipalign"), path)
}
