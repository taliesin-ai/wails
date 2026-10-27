package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

func TestAndroidReleaseAPKDebugSignDecision(t *testing.T) {
	tests := []struct {
		name string
		spec pipeline.PackageSpec
		want bool
	}{
		{name: "unsigned release apk falls back to debug keystore", spec: pipeline.PackageSpec{Format: "apk", Release: true, DebugSign: true}, want: true},
		{name: "release apk signed by sign stage is not debug-signed", spec: pipeline.PackageSpec{Format: "apk", Release: true}, want: false},
		{name: "development apk stays a Gradle debug build", spec: pipeline.PackageSpec{Format: "apk", DebugSign: true}, want: false},
		{name: "aab is never debug-signed by the apk path", spec: pipeline.PackageSpec{Format: "aab", Release: true, DebugSign: true}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, needsAndroidDebugSign(test.spec))
		})
	}
}

func TestDebugKeystoreSignArgsUseAndroidDefaults(t *testing.T) {
	args := debugKeystoreSignArgs("/home/dev/.android/debug.keystore")
	assert.Equal(t, []string{
		"--ks", "/home/dev/.android/debug.keystore",
		"--ks-key-alias", "androiddebugkey",
		"--ks-pass", "pass:android",
		"--key-pass", "pass:android",
	}, args)
}

func TestAndroidDebugSignWarningSaysNotForDistribution(t *testing.T) {
	assert.Contains(t, androidDebugSignWarning, "debug keystore")
	assert.Contains(t, androidDebugSignWarning, "NOT suitable for distribution")
}
