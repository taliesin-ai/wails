package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
	"gopkg.in/yaml.v3"
)

func TestNFPMDebSignatureTypeMapsDpkgRoles(t *testing.T) {
	for role, want := range map[string]string{"": "origin", "builder": "origin", "origin": "origin", "maint": "maint", "archive": "archive"} {
		got, err := nfpmDebSignatureType(role)
		require.NoError(t, err, role)
		assert.Equal(t, want, got, role)
	}
	_, err := nfpmDebSignatureType("release")
	assert.ErrorContains(t, err, "not a deb signature role")
}

func TestApplyNFPMSignatureWritesKeyFileAndRoleWithoutPassphrase(t *testing.T) {
	config := map[string]any{}
	require.NoError(t, applyNFPMSignature(config, "deb", "/project/signing.asc", "maint"))
	require.NoError(t, applyNFPMSignature(config, "rpm", "/project/signing.asc", ""))
	data, err := yaml.Marshal(config)
	require.NoError(t, err)
	text := string(data)
	assert.Contains(t, text, "key_file: /project/signing.asc")
	assert.Contains(t, text, "type: maint")
	assert.NotContains(t, text, "passphrase", "passphrases must never be written to the nfpm configuration")

	assert.ErrorContains(t, applyNFPMSignature(map[string]any{}, "archlinux", "k", ""), "does not support")
	assert.ErrorContains(t, applyNFPMSignature(map[string]any{}, "deb", "k", "release"), "not a deb signature role")
}

func TestLinuxSigningKeyDigestChangesWithKeyContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "signing.asc")
	require.NoError(t, os.WriteFile(path, []byte("key-one"), 0o600))
	first, err := linuxSigningKeyDigest(path)
	require.NoError(t, err)
	again, err := linuxSigningKeyDigest(path)
	require.NoError(t, err)
	assert.Equal(t, first, again)
	require.NoError(t, os.WriteFile(path, []byte("key-two"), 0o600))
	second, err := linuxSigningKeyDigest(path)
	require.NoError(t, err)
	assert.NotEqual(t, first, second)
	assert.NotContains(t, first, "key-one", "the digest must not expose key material")
}

func TestPackageIdentityIncludesSigningKeyDigest(t *testing.T) {
	root := t.TempDir()
	keyPath := filepath.Join(root, "signing.asc")
	require.NoError(t, os.WriteFile(keyPath, []byte("key-one"), 0o600))
	node := pipeline.Node{Kind: pipeline.PackageArtifact, Spec: pipeline.PackageSpec{TargetOS: "linux", TargetArch: "amd64", Format: "deb", NativeSigning: manifest.SigningPlatform{Certificate: "signing.asc", Identity: "origin"}}}
	handler := &manifestHandler{root: root}
	first, err := handler.Identity(t.Context(), node)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(keyPath, []byte("key-two"), 0o600))
	second, err := handler.Identity(t.Context(), node)
	require.NoError(t, err)
	assert.NotEqual(t, first, second, "rotating the signing key must invalidate the package action")
	assert.Contains(t, second, "linux-signing-key:")
}

func TestDebHasNativeSignatureDetectsSignatureMember(t *testing.T) {
	dir := t.TempDir()
	signed := filepath.Join(dir, "signed.deb")
	require.NoError(t, os.WriteFile(signed, arArchive(t, "debian-binary", "_gpgorigin", "control.tar.gz"), 0o644))
	ok, err := debHasNativeSignature(signed)
	require.NoError(t, err)
	assert.True(t, ok)

	unsigned := filepath.Join(dir, "unsigned.deb")
	require.NoError(t, os.WriteFile(unsigned, arArchive(t, "debian-binary", "control.tar.gz", "data.tar.xz"), 0o644))
	ok, err = debHasNativeSignature(unsigned)
	require.NoError(t, err)
	assert.False(t, ok)

	notAr := filepath.Join(dir, "plain.deb")
	require.NoError(t, os.WriteFile(notAr, []byte("unsigned"), 0o644))
	_, err = debHasNativeSignature(notAr)
	assert.ErrorContains(t, err, "not a deb archive")
}

// arArchive builds a minimal ar archive with the given members, each holding
// a two-byte body, which is enough to exercise the member walk.
func arArchive(t *testing.T, names ...string) []byte {
	t.Helper()
	var buf bytes.Buffer
	buf.WriteString("!<arch>\n")
	for _, name := range names {
		header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8o%-10d`\n", name, 0, 0, 0, 0o644, 2)
		buf.WriteString(header)
		buf.WriteString("ok")
	}
	return buf.Bytes()
}

func TestPassphraseProtectedLinuxDebIsSignedFromCredential(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("gpg key generation in this test uses POSIX tooling")
	}
	root := t.TempDir()
	t.Chdir(root)
	const passphrase = "wails-test-passphrase"
	writeTestPGPKey(t, root, "release-key", passphrase)
	t.Setenv("WAILS_TEST_PGP_PASSPHRASE", passphrase)
	yamlText := `version: 3
project:
  name: signed
  product_name: Signed
  identifier: com.example.signed
  version: 1.0.0
linux:
  signing:
    identity: origin
    certificate: release-key
    credential: WAILS_TEST_PGP_PASSPHRASE
profiles:
  release:
    targets:
      linux/amd64:
        formats: [deb]
        sign: true
`
	require.NoError(t, os.WriteFile(filepath.Join(root, manifest.Filename), []byte(yamlText), 0o644))
	loaded, err := manifest.Load(root, "release")
	require.NoError(t, err)
	plan, err := pipeline.PlanBuild(loaded.Config, pipeline.Request{Verb: "build"})
	require.NoError(t, err)
	handler := &manifestHandler{root: root, config: loaded.Config}
	_, err = handler.Run(t.Context(), plan.Nodes[pipeline.NodeKey("target:linux/amd64:assets")])
	require.NoError(t, err)
	packageNode := plan.Nodes[pipeline.NodeKey("package:linux/amd64:deb")]
	spec := packageNode.Spec.(pipeline.PackageSpec)
	require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(root, spec.Binary)), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, spec.Binary), []byte("binary"), 0o755))
	_, err = handler.Run(t.Context(), packageNode)
	require.NoError(t, err)
	signed, err := debHasNativeSignature(filepath.Join(root, spec.Output))
	require.NoError(t, err)
	assert.True(t, signed)
	config, err := os.ReadFile(filepath.Join(root, spec.Output))
	require.NoError(t, err)
	assert.NotContains(t, string(config), passphrase, "the passphrase must never be written into the package output path")
	// The generated nfpm configuration lives in the package workspace; it must not contain the secret.
	generated, err := os.ReadFile(filepath.Join(handler.packageWorkspace(spec), "nfpm.yaml"))
	require.NoError(t, err)
	assert.NotContains(t, string(generated), passphrase)
	assert.True(t, strings.Contains(string(generated), "key_file:"))
}
