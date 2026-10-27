package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// AndroidBuildToolPath resolves an Android SDK build-tools executable such as
// zipalign or apksigner. It prefers PATH, then the newest build-tools directory
// under ANDROID_HOME or ANDROID_SDK_ROOT, because the SDK tools are rarely on PATH.
func AndroidBuildToolPath(name string) (string, bool) {
	return androidBuildToolPath(name, runtime.GOOS, os.Getenv, exec.LookPath, filepath.Glob)
}

func androidBuildToolPath(name, goos string, getenv func(string) string, lookPath func(string) (string, error), glob func(string) ([]string, error)) (string, bool) {
	if path, err := lookPath(name); err == nil {
		return path, true
	}
	candidates := []string{name}
	if goos == "windows" {
		// The SDK ships zipalign.exe and apksigner.bat / apksigner.jar style launchers.
		candidates = []string{name + ".exe", name + ".bat", name + ".cmd"}
	}
	for _, key := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		sdk := getenv(key)
		if sdk == "" {
			continue
		}
		best, found := "", false
		for _, candidate := range candidates {
			matches, _ := glob(filepath.Join(sdk, "build-tools", "*", candidate))
			for _, match := range matches {
				dir := filepath.Base(filepath.Dir(match))
				if _, ok := parseBuildToolVersion(dir); !ok {
					continue
				}
				// Only a strictly newer version replaces the current pick, so the
				// candidate order above breaks ties within one build-tools directory.
				if !found || compareBuildToolDirs(dir, filepath.Base(filepath.Dir(best))) > 0 {
					best, found = match, true
				}
			}
		}
		if found {
			return best, true
		}
	}
	return "", false
}

// buildToolVersion is a parsed build-tools directory name such as "36.0.0" or
// "36.0.0-rc1". Pre-release identifiers are kept in order for comparison.
type buildToolVersion struct {
	core       []int
	prerelease []string
}

func parseBuildToolVersion(dir string) (buildToolVersion, bool) {
	var version buildToolVersion
	core, pre, hasPre := strings.Cut(dir, "-")
	if core == "" {
		return version, false
	}
	for _, part := range strings.Split(core, ".") {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return buildToolVersion{}, false
		}
		version.core = append(version.core, number)
	}
	if hasPre {
		if pre == "" {
			return buildToolVersion{}, false
		}
		version.prerelease = strings.Split(pre, ".")
	}
	return version, true
}

// compareBuildToolDirs orders two build-tools directory names. Unparseable names
// rank below every parseable name. Returns >0 when a is newer than b.
func compareBuildToolDirs(a, b string) int {
	va, okA := parseBuildToolVersion(a)
	vb, okB := parseBuildToolVersion(b)
	switch {
	case okA && !okB:
		return 1
	case !okA && okB:
		return -1
	case !okA && !okB:
		return strings.Compare(a, b)
	}
	return compareBuildToolVersions(va, vb)
}

func compareBuildToolVersions(a, b buildToolVersion) int {
	for i := 0; i < len(a.core) || i < len(b.core); i++ {
		x, y := 0, 0
		if i < len(a.core) {
			x = a.core[i]
		}
		if i < len(b.core) {
			y = b.core[i]
		}
		if x != y {
			if x > y {
				return 1
			}
			return -1
		}
	}
	// A final release outranks any pre-release of the same core version.
	switch {
	case len(a.prerelease) == 0 && len(b.prerelease) == 0:
		return 0
	case len(a.prerelease) == 0:
		return 1
	case len(b.prerelease) == 0:
		return -1
	}
	for i := 0; i < len(a.prerelease) || i < len(b.prerelease); i++ {
		if i >= len(a.prerelease) {
			return -1
		}
		if i >= len(b.prerelease) {
			return 1
		}
		if c := comparePrereleaseIdentifier(a.prerelease[i], b.prerelease[i]); c != 0 {
			return c
		}
	}
	return 0
}

// comparePrereleaseIdentifier compares identifiers naturally, so rc2 < rc10,
// and compares purely numeric identifiers as numbers.
func comparePrereleaseIdentifier(a, b string) int {
	for a != "" && b != "" {
		chunkA, restA := splitIdentifierChunk(a)
		chunkB, restB := splitIdentifierChunk(b)
		numA, errA := strconv.Atoi(chunkA)
		numB, errB := strconv.Atoi(chunkB)
		var c int
		if errA == nil && errB == nil {
			switch {
			case numA > numB:
				c = 1
			case numA < numB:
				c = -1
			}
		} else {
			c = strings.Compare(chunkA, chunkB)
		}
		if c != 0 {
			return c
		}
		a, b = restA, restB
	}
	return strings.Compare(a, b)
}

// splitIdentifierChunk splits off the leading run of digits or non-digits.
func splitIdentifierChunk(s string) (string, string) {
	digit := s[0] >= '0' && s[0] <= '9'
	for i := 1; i < len(s); i++ {
		if (s[i] >= '0' && s[i] <= '9') != digit {
			return s[:i], s[i:]
		}
	}
	return s, ""
}
