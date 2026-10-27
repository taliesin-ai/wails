package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

func apkPackageSpecs(plan Plan) []PackageSpec {
	var specs []PackageSpec
	for _, node := range plan.Nodes {
		if spec, ok := node.Spec.(PackageSpec); ok && spec.Format == "apk" {
			specs = append(specs, spec)
		}
	}
	return specs
}

func hasSignNode(plan Plan) bool {
	for _, node := range plan.Nodes {
		if node.Kind == SignArtifact {
			return true
		}
	}
	return false
}

func TestProductionAPKPlansGradleReleaseBuildForArm64AndUniversal(t *testing.T) {
	for _, arch := range []string{"arm64", "universal"} {
		t.Run(arch, func(t *testing.T) {
			plan, err := PlanBuild(testConfig(t), Request{Verb: "build", TargetOS: "android", TargetArch: arch, Formats: []string{"apk"}})
			require.NoError(t, err)
			specs := apkPackageSpecs(plan)
			require.Len(t, specs, 1)
			assert.False(t, specs[0].Development)
			assert.True(t, specs[0].Release, "production APK must use the Gradle release task")
			assert.True(t, specs[0].DebugSign, "no sign stage runs, so the release APK must be debug-signed")
			assert.False(t, hasSignNode(plan))
		})
	}
}

func TestDevelopmentAndRunAPKStayDebugBuilds(t *testing.T) {
	plan, err := PlanBuild(testConfig(t), Request{Verb: "build", TargetOS: "android", TargetArch: "arm64", Formats: []string{"apk"}, Development: true})
	require.NoError(t, err)
	for _, spec := range apkPackageSpecs(plan) {
		assert.False(t, spec.Release)
		assert.False(t, spec.DebugSign)
	}

	plan, err = PlanBuild(testConfig(t), Request{Verb: "run", TargetOS: "android", TargetArch: "arm64"})
	require.NoError(t, err)
	for _, spec := range apkPackageSpecs(plan) {
		assert.False(t, spec.Release)
		assert.False(t, spec.DebugSign)
	}
}

func TestConfiguredAndroidSigningSignsReleaseAPKInEveryVerb(t *testing.T) {
	config := testConfig(t)
	config.Signing.Android = manifest.SigningPlatform{Enabled: true, Certificate: "android.keystore", KeyAlias: "upload", Credential: "ANDROID_PASSWORD"}
	for _, verb := range []string{"build", "package", "sign"} {
		t.Run(verb, func(t *testing.T) {
			plan, err := PlanBuild(config, Request{Verb: verb, TargetOS: "android", TargetArch: "arm64", Formats: []string{"apk"}})
			require.NoError(t, err)
			specs := apkPackageSpecs(plan)
			require.Len(t, specs, 1)
			assert.True(t, specs[0].Release)
			assert.False(t, specs[0].DebugSign, "configured release signing must replace the debug fallback")
			assert.True(t, hasSignNode(plan), "configured release signing must plan a sign stage")
		})
	}
}

func TestAABKeepsItsReleaseBundleBehaviour(t *testing.T) {
	plan, err := PlanBuild(testConfig(t), Request{Verb: "build", TargetOS: "android", TargetArch: "arm64"})
	require.NoError(t, err)
	for _, node := range plan.Nodes {
		if spec, ok := node.Spec.(PackageSpec); ok && spec.Format == "aab" {
			assert.False(t, spec.Release)
			assert.False(t, spec.DebugSign)
		}
	}
}

func TestMobileBindingsCarryOSTagForEveryVerb(t *testing.T) {
	type verb struct {
		name    string
		request Request
	}
	verbs := []verb{
		{name: "build", request: Request{Verb: "build"}},
		{name: "package", request: Request{Verb: "package", Formats: []string{"aab"}}},
		{name: "sign", request: Request{Verb: "sign", Formats: []string{"aab"}}},
		{name: "dev", request: Request{Verb: "build", Development: true}},
		{name: "run", request: Request{Verb: "run"}},
	}
	for _, targetOS := range []string{"android", "ios"} {
		for _, v := range verbs {
			t.Run(targetOS+"/"+v.name, func(t *testing.T) {
				config := testConfig(t)
				config.Signing.IOS = manifest.SigningPlatform{Enabled: true, Identity: "Apple Distribution"}
				request := v.request
				request.TargetOS, request.TargetArch = targetOS, "arm64"
				if targetOS == "ios" && (v.name == "package" || v.name == "sign") {
					// IPA packaging is only valid for a device destination, which comes from a profile.
					config.Selected = manifest.Profile{Name: "store", Targets: []manifest.ProfileTarget{{Target: "ios/arm64", Destination: "device", Formats: []string{"ipa"}}}}
					request.TargetOS, request.TargetArch, request.Formats = "", "", nil
				}
				plan, err := PlanBuild(config, request)
				require.NoError(t, err)
				bindings, ok := plan.Nodes[NodeKey("frontend:bindings")]
				require.True(t, ok, "bindings node must be planned")
				assert.Contains(t, bindings.Spec.(BindingsSpec).Tags, targetOS)
			})
		}
	}
}

func TestSharedMobileTagRequiresEveryTargetToBeTheSameMobileOS(t *testing.T) {
	assert.Equal(t, "", sharedMobileTag(nil))
	assert.Equal(t, "android", sharedMobileTag([]buildOutcome{{target: Target{OS: "android", Arch: "arm64"}}, {target: Target{OS: "android", Arch: "amd64"}}}))
	assert.Equal(t, "ios", sharedMobileTag([]buildOutcome{{target: Target{OS: "ios", Arch: "arm64"}}}))
	assert.Equal(t, "", sharedMobileTag([]buildOutcome{{target: Target{OS: "android", Arch: "arm64"}}, {target: Target{OS: "linux", Arch: "amd64"}}}))
	assert.Equal(t, "", sharedMobileTag([]buildOutcome{{target: Target{OS: "linux", Arch: "amd64"}}}))
}

func TestAndroidBuildToolPathPrefersPATHThenNewestSDKBuildTools(t *testing.T) {
	env := map[string]string{"ANDROID_HOME": "/sdk"}
	getenv := func(key string) string { return env[key] }
	globbed := []string{"/sdk/build-tools/34.0.0/apksigner", "/sdk/build-tools/36.0.0/apksigner"}
	glob := func(pattern string) ([]string, error) {
		assert.Equal(t, "/sdk/build-tools/*/apksigner", pattern)
		return globbed, nil
	}
	notFound := func(string) (string, error) { return "", assert.AnError }

	path, ok := androidBuildToolPath("apksigner", getenv, func(string) (string, error) { return "/usr/bin/apksigner", nil }, glob)
	assert.True(t, ok)
	assert.Equal(t, "/usr/bin/apksigner", path)

	path, ok = androidBuildToolPath("apksigner", getenv, notFound, glob)
	assert.True(t, ok)
	assert.Equal(t, "/sdk/build-tools/36.0.0/apksigner", path)

	env = map[string]string{}
	_, ok = androidBuildToolPath("apksigner", getenv, notFound, glob)
	assert.False(t, ok)
}
