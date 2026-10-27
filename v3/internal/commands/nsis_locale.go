package commands

import (
	"os"
	"os/exec"
	"strings"
)

// nsisUTF8Candidates are the UTF-8 locales tried, in order, for makensis.
var nsisUTF8Candidates = []string{"C.UTF-8", "en_US.UTF-8"}

// nsisEnvironment returns the environment makensis should run with. makensis
// reads its scripts with the host's text encoding, so on a non-UTF-8 POSIX
// locale (LANG unset, C, POSIX) a stock script containing UTF-8 characters such
// as © fails with "Bad text encoding". When the effective locale is not UTF-8,
// LC_ALL is set to the first candidate the host supports. A locale that is
// already UTF-8 is never overridden, and Windows is left untouched.
//
// supports reports whether a locale name is usable on the host. It is injected
// so the decision stays a pure function of its inputs.
func nsisEnvironment(environment []string, goos string, supports func(locale string) bool) []string {
	if goos == "windows" {
		return environment
	}
	if isUTF8Locale(effectiveLocale(environment)) {
		return environment
	}
	for _, candidate := range nsisUTF8Candidates {
		if supports(candidate) {
			return mergeEnvironment(environment, []string{"LC_ALL=" + candidate})
		}
	}
	return environment
}

// effectiveLocale follows POSIX precedence: LC_ALL, then LC_CTYPE, then LANG.
// Empty values count as unset.
func effectiveLocale(environment []string) string {
	values := map[string]string{}
	for _, raw := range environment {
		key, value, ok := strings.Cut(raw, "=")
		if ok && value != "" {
			values[key] = value
		}
	}
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if value := values[key]; value != "" {
			return value
		}
	}
	return ""
}

func isUTF8Locale(locale string) bool {
	normalised := strings.ToLower(locale)
	return strings.Contains(normalised, "utf-8") || strings.Contains(normalised, "utf8")
}

// nsisLocaleSupported probes the host with `locale charmap` under the candidate
// locale. Any failure counts as unsupported.
func nsisLocaleSupported(locale string) bool {
	cmd := exec.Command("locale", "charmap")
	cmd.Env = mergeEnvironment(os.Environ(), []string{"LC_ALL=" + locale})
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "UTF-8"
}
