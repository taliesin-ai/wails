package commands

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func localeEnv(entries ...string) []string {
	return append([]string(nil), entries...)
}

// envValue returns the value of key and how many entries set it.
func envValue(environment []string, key string) (string, int) {
	value, count := "", 0
	for _, raw := range environment {
		k, v, ok := strings.Cut(raw, "=")
		if ok && k == key {
			value = v
			count++
		}
	}
	return value, count
}

func supportsAll(string) bool { return true }

func TestNSISEnvironmentOverridesNonUTF8Locale(t *testing.T) {
	for name, environment := range map[string][]string{
		"LANG unset":           localeEnv("PATH=/bin"),
		"LANG C":               localeEnv("PATH=/bin", "LANG=C"),
		"LC_ALL POSIX":         localeEnv("LC_ALL=POSIX", "LANG=en_US.UTF-8"),
		"LC_CTYPE non-UTF-8":   localeEnv("LC_CTYPE=C", "LANG=en_US.UTF-8"),
		"empty LC_ALL ignored": localeEnv("LC_ALL=", "LANG=C"),
	} {
		t.Run(name, func(t *testing.T) {
			got := nsisEnvironment(environment, "darwin", supportsAll)
			value, count := envValue(got, "LC_ALL")
			assert.Equal(t, "C.UTF-8", value)
			assert.Equal(t, 1, count)
		})
	}
}

func TestNSISEnvironmentFallsBackToEnUSUTF8(t *testing.T) {
	environment := localeEnv("LANG=C")
	got := nsisEnvironment(environment, "linux", func(locale string) bool { return locale == "en_US.UTF-8" })
	value, _ := envValue(got, "LC_ALL")
	assert.Equal(t, "en_US.UTF-8", value)
}

func TestNSISEnvironmentLeavesEnvironmentWhenNoUTF8Available(t *testing.T) {
	environment := localeEnv("LANG=C")
	got := nsisEnvironment(environment, "darwin", func(string) bool { return false })
	assert.Equal(t, environment, got)
}

func TestNSISEnvironmentKeepsExistingUTF8Locale(t *testing.T) {
	for name, environment := range map[string][]string{
		"LANG utf8":        localeEnv("LANG=en_US.UTF-8"),
		"LC_ALL overrides": localeEnv("LC_ALL=en_GB.utf8", "LANG=C"),
		"LC_CTYPE utf-8":   localeEnv("LC_CTYPE=C.UTF-8"),
	} {
		t.Run(name, func(t *testing.T) {
			got := nsisEnvironment(environment, "darwin", func(string) bool {
				t.Fatal("UTF-8 locale must not be probed or overridden")
				return false
			})
			assert.Equal(t, environment, got)
		})
	}
}

func TestNSISEnvironmentIgnoresWindows(t *testing.T) {
	environment := localeEnv("LANG=C")
	got := nsisEnvironment(environment, "windows", supportsAll)
	assert.Equal(t, environment, got)
}

func TestNSISEffectiveLocalePrecedence(t *testing.T) {
	assert.Equal(t, "a.UTF-8", effectiveLocale(localeEnv("LANG=c", "LC_CTYPE=b", "LC_ALL=a.UTF-8")))
	assert.Equal(t, "b", effectiveLocale(localeEnv("LANG=c", "LC_CTYPE=b")))
	assert.Equal(t, "c", effectiveLocale(localeEnv("LANG=c")))
	assert.Equal(t, "", effectiveLocale(localeEnv("PATH=/bin")))
}
