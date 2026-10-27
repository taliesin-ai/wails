package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

func TestNativeLinuxSigningOnlyForSignedDebAndRPM(t *testing.T) {
	signing := manifest.Signing{Linux: manifest.SigningPlatform{Enabled: true, Certificate: "signing.asc", Identity: "maint", Credential: "PGP_PASS"}}
	signed := Request{TargetOS: "linux", TargetArch: "amd64", sign: true}
	assert.Equal(t, signing.Linux, nativeLinuxSigning(signed, signing, "deb"))
	assert.Equal(t, signing.Linux, nativeLinuxSigning(signed, signing, "rpm"))

	assert.Equal(t, manifest.SigningPlatform{}, nativeLinuxSigning(signed, signing, "archlinux"), "archlinux has no native signer")
	assert.Equal(t, manifest.SigningPlatform{}, nativeLinuxSigning(Request{TargetOS: "linux", TargetArch: "amd64"}, signing, "deb"), "unsigned plans stay unsigned")
	assert.Equal(t, manifest.SigningPlatform{}, nativeLinuxSigning(Request{TargetOS: "darwin", TargetArch: "arm64", sign: true}, signing, "deb"))
	disabled := manifest.Signing{}
	assert.Equal(t, manifest.SigningPlatform{}, nativeLinuxSigning(signed, disabled, "deb"), "signing without configuration is not native")
}

func TestNativeLinuxSigningCredentialMustBeAvailableOnHost(t *testing.T) {
	config := testConfig(t)
	config.Signing.Linux.Enabled = true
	config.Signing.Linux.Certificate = "release@example.com"
	config.Signing.Linux.Credential = "PGP_PASS"
	request := Request{Verb: "sign", TargetOS: "linux", TargetArch: "amd64", Formats: []string{"deb"}}
	_, err := PlanBuildForHost(config, request, testHost("linux", "amd64"))
	require.Error(t, err)
	assert.ErrorContains(t, err, `"PGP_PASS"`)
	assert.ErrorContains(t, err, "unavailable")
}
