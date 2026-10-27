package pipeline

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// fakeWindowsHost models a Windows host: PATH lookups fail, environment
// variables are a map, and the files that exist are a set.
type fakeWindowsHost struct {
	env   map[string]string
	files map[string]bool
}

func (h fakeWindowsHost) getenv(key string) string { return h.env[key] }

func (h fakeWindowsHost) lookPath(string) (string, error) {
	return "", errors.New("not on PATH")
}

// fakeFileInfo is a regular file for stat results.
type fakeFileInfo struct{ fs.FileInfo }

func (fakeFileInfo) IsDir() bool { return false }

func (h fakeWindowsHost) stat(path string) (fs.FileInfo, error) {
	if h.files[path] {
		return fakeFileInfo{}, nil
	}
	return nil, fs.ErrNotExist
}

func (h fakeWindowsHost) glob(pattern string) ([]string, error) {
	var matches []string
	for file := range h.files {
		if ok, _ := filepath.Match(pattern, file); ok {
			matches = append(matches, file)
		}
	}
	return matches, nil
}

func TestWindowsToolPathFindsNSISInstallLocations(t *testing.T) {
	x86 := filepath.Join("C:", "Program Files (x86)")
	native := filepath.Join("C:", "Program Files")
	nsisX86 := filepath.Join(x86, "NSIS", "makensis.exe")
	nsisNative := filepath.Join(native, "NSIS", "makensis.exe")

	host := fakeWindowsHost{env: map[string]string{"ProgramFiles(x86)": x86, "ProgramFiles": native}, files: map[string]bool{nsisNative: true}}
	path, ok := windowsToolPath("makensis", "windows", "amd64", host.getenv, host.lookPath, host.stat, host.glob)
	assert.True(t, ok)
	assert.Equal(t, nsisNative, path, "falls back to ProgramFiles when x86 is absent")

	host.files[nsisX86] = true
	path, ok = windowsToolPath("makensis", "windows", "amd64", host.getenv, host.lookPath, host.stat, host.glob)
	assert.True(t, ok)
	assert.Equal(t, nsisX86, path, "prefers ProgramFiles(x86)")
}

func TestWindowsToolPathSelectsHighestSDKVersionForHostArchitecture(t *testing.T) {
	x86 := filepath.Join("C:", "Program Files (x86)")
	bin := filepath.Join(x86, "Windows Kits", "10", "bin")
	files := map[string]bool{
		filepath.Join(bin, "10.0.19041.0", "x64", "signtool.exe"):   true,
		filepath.Join(bin, "10.0.9.0", "x64", "signtool.exe"):       true,
		filepath.Join(bin, "10.0.26100.0", "x64", "signtool.exe"):   true,
		filepath.Join(bin, "10.0.26100.0", "arm64", "signtool.exe"): true,
		filepath.Join(bin, "10.0.26100.0", "x64", "makeappx.exe"):   true,
		filepath.Join(bin, "x64", "signtool.exe"):                   true,
	}
	host := fakeWindowsHost{env: map[string]string{"ProgramFiles(x86)": x86}, files: files}

	path, ok := windowsToolPath("signtool.exe", "windows", "amd64", host.getenv, host.lookPath, host.stat, host.glob)
	assert.True(t, ok)
	assert.Equal(t, filepath.Join(bin, "10.0.26100.0", "x64", "signtool.exe"), path, "numeric version sort, host arch x64")

	path, ok = windowsToolPath("signtool.exe", "windows", "arm64", host.getenv, host.lookPath, host.stat, host.glob)
	assert.True(t, ok)
	assert.Equal(t, filepath.Join(bin, "10.0.26100.0", "arm64", "signtool.exe"), path)

	path, ok = windowsToolPath("MakeAppx.exe", "windows", "amd64", host.getenv, host.lookPath, host.stat, host.glob)
	assert.True(t, ok)
	assert.Equal(t, filepath.Join(bin, "10.0.26100.0", "x64", "makeappx.exe"), path, "MakeAppx maps to makeappx.exe")
}

func TestWindowsToolPathReportsMissingAndNonWindowsHosts(t *testing.T) {
	host := fakeWindowsHost{env: map[string]string{"ProgramFiles": `C:\Program Files`}, files: map[string]bool{}}

	_, ok := windowsToolPath("makensis", "windows", "amd64", host.getenv, host.lookPath, host.stat, host.glob)
	assert.False(t, ok, "no NSIS installed")

	_, ok = windowsToolPath("signtool.exe", "windows", "riscv64", host.getenv, host.lookPath, host.stat, host.glob)
	assert.False(t, ok, "no SDK architecture for riscv64")

	_, ok = windowsToolPath("makensis", "linux", "amd64", host.getenv, host.lookPath, host.stat, host.glob)
	assert.False(t, ok, "install locations are Windows-only")
}

func TestHostProbeAgreesWithWindowsToolResolution(t *testing.T) {
	programFiles := filepath.Join("C:", "Program Files")
	host := fakeWindowsHost{env: map[string]string{"ProgramFiles": programFiles}, files: map[string]bool{
		filepath.Join(programFiles, "NSIS", "makensis.exe"): true,
	}}
	operations := hostProbeOperations{
		hostOS: "windows", hostArch: "amd64", lookPath: host.lookPath,
		lookupEnv: func(string) (string, bool) { return "", false },
		getenv:    host.getenv, stat: host.stat, glob: host.glob,
	}
	tools, _ := probeHostToolsAndCredentials(nil, operations)
	assert.Contains(t, tools, "makensis", "NSIS in ProgramFiles is a usable makensis")
	assert.NotContains(t, tools, "signtool.exe", "no SDK installed")
}

func TestWindowsToolPathPrefersPATH(t *testing.T) {
	onPath := func(string) (string, error) { return `C:\tools\makensis.exe`, nil }
	path, ok := windowsToolPath("makensis", "windows", "amd64", func(string) string { return "" }, onPath, nil, nil)
	assert.True(t, ok)
	assert.Equal(t, `C:\tools\makensis.exe`, path)
}
