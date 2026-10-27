package pipeline

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// WindowsToolPath resolves a Windows packaging or signing executable on the
// current host. It prefers PATH and then the standard install locations, since
// NSIS and the Windows SDK are usually not on PATH by default.
func WindowsToolPath(name string) (string, bool) {
	return windowsToolPath(name, runtime.GOOS, runtime.GOARCH, os.Getenv, exec.LookPath, os.Stat, filepath.Glob)
}

// windowsToolPath is the pure resolution used by the host probe, the signing
// identity and every execution site, so all three agree on one path.
//
//   - makensis.exe: %ProgramFiles(x86)%\NSIS, then %ProgramFiles%\NSIS.
//   - signtool.exe and makeappx.exe: the highest numeric version directory under
//     %ProgramFiles(x86)%\Windows Kits\10\bin, then %ProgramFiles%, for the host
//     architecture (x64, arm64 or x86).
func windowsToolPath(name, goos, goarch string, getenv func(string) string, lookPath func(string) (string, error), stat func(string) (fs.FileInfo, error), glob func(string) ([]string, error)) (string, bool) {
	if path, err := lookPath(name); err == nil {
		return path, true
	}
	if goos != "windows" {
		return "", false
	}
	programFiles := []string{"ProgramFiles(x86)", "ProgramFiles"}
	switch strings.ToLower(name) {
	case "makensis", "makensis.exe":
		for _, key := range programFiles {
			root := getenv(key)
			if root == "" {
				continue
			}
			candidate := filepath.Join(root, "NSIS", "makensis.exe")
			if info, err := stat(candidate); err == nil && !info.IsDir() {
				return candidate, true
			}
		}
	case "signtool.exe", "makeappx.exe", "makeappx":
		file := "signtool.exe"
		if !strings.HasPrefix(strings.ToLower(name), "signtool") {
			file = "makeappx.exe"
		}
		arch, ok := windowsSDKArchitecture(goarch)
		if !ok {
			return "", false
		}
		for _, key := range programFiles {
			root := getenv(key)
			if root == "" {
				continue
			}
			matches, _ := glob(filepath.Join(root, "Windows Kits", "10", "bin", "*", arch, file))
			best, found := "", false
			for _, match := range matches {
				version := filepath.Base(filepath.Dir(filepath.Dir(match)))
				if _, ok := parseBuildToolVersion(version); !ok {
					continue
				}
				if !found || compareBuildToolDirs(version, filepath.Base(filepath.Dir(filepath.Dir(best)))) > 0 {
					best, found = match, true
				}
			}
			if found {
				return best, true
			}
		}
	}
	return "", false
}

// windowsSDKArchitecture maps a Go architecture to the directory name the
// Windows SDK uses under bin\<version>\.
func windowsSDKArchitecture(goarch string) (string, bool) {
	switch goarch {
	case "amd64":
		return "x64", true
	case "arm64":
		return "arm64", true
	case "386":
		return "x86", true
	}
	return "", false
}
