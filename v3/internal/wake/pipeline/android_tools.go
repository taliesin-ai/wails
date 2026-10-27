package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
)

// AndroidBuildToolPath resolves an Android SDK build-tools executable such as
// zipalign or apksigner. It prefers PATH, then the newest build-tools directory
// under ANDROID_HOME or ANDROID_SDK_ROOT, because the SDK tools are rarely on PATH.
func AndroidBuildToolPath(name string) (string, bool) {
	return androidBuildToolPath(name, os.Getenv, exec.LookPath, filepath.Glob)
}

func androidBuildToolPath(name string, getenv func(string) string, lookPath func(string) (string, error), glob func(string) ([]string, error)) (string, bool) {
	if path, err := lookPath(name); err == nil {
		return path, true
	}
	for _, key := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		sdk := getenv(key)
		if sdk == "" {
			continue
		}
		matches, _ := glob(filepath.Join(sdk, "build-tools", "*", name))
		if len(matches) > 0 {
			// Glob returns matches in lexical order; the last one is the highest
			// version for the current SDK layout (e.g. 34.x, 36.x).
			return matches[len(matches)-1], true
		}
	}
	return "", false
}
