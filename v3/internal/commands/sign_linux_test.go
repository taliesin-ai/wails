package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirstGPGFingerprintReadsColonOutput(t *testing.T) {
	output := "sec:u:255:22:ABCDEF0123456789:1700000000:::u:::scESC:::::ed25519:::0:\nfpr:::::::::0123456789ABCDEF0123456789ABCDEF01234567:\n"
	assert.Equal(t, "0123456789ABCDEF0123456789ABCDEF01234567", firstGPGFingerprint(output))
	assert.Empty(t, firstGPGFingerprint("sec:u:255\n"))
}

func TestLegacyLinuxSigningWithKeyIDNeedsNoImport(t *testing.T) {
	signing, cleanup, err := prepareLinuxPGPSigning("ABCDEF0123456789", "")
	require.NoError(t, err)
	defer cleanup()
	assert.Equal(t, "ABCDEF0123456789", signing.keyID)
	assert.Empty(t, signing.homedir)
	assert.Empty(t, signing.passphraseFile)
}

func TestLegacyRPMSignArgsNeverCarryThePassphrase(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not meaningful on windows")
	}
	const passphrase = "legacy-test-passphrase"
	keyFile := writeTestPGPKey(t, t.TempDir(), "release.asc", passphrase)
	signing, cleanup, err := prepareLinuxPGPSigning(keyFile, passphrase)
	require.NoError(t, err)

	args := rpmsignArgs(signing, "/tmp/app.rpm")
	joined := strings.Join(args, " ")
	assert.NotContains(t, joined, passphrase, "the passphrase must reach rpmsign through a file, never argv")
	assert.Contains(t, joined, "_gpg_name "+signing.keyID)
	assert.Contains(t, joined, "--passphrase-file "+signing.passphraseFile)
	assert.Equal(t, "--addsign", args[len(args)-2])
	assert.Equal(t, "/tmp/app.rpm", args[len(args)-1])

	info, err := os.Stat(signing.passphraseFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "the passphrase file must be private")
	content, err := os.ReadFile(signing.passphraseFile)
	require.NoError(t, err)
	assert.Equal(t, passphrase, string(content))

	passphraseFile := signing.passphraseFile
	cleanup()
	_, err = os.Stat(passphraseFile)
	assert.True(t, os.IsNotExist(err), "cleanup must remove the passphrase file")
}

func TestLegacyDEBSignArgsUseKeyIDAndNeverArgvPassphrase(t *testing.T) {
	const passphrase = "legacy-deb-passphrase"
	keyFile := writeTestPGPKey(t, t.TempDir(), "release.asc", passphrase)
	signing, cleanup, err := prepareLinuxPGPSigning(keyFile, passphrase)
	require.NoError(t, err)
	defer cleanup()

	args := debsignArgs(signing, "origin", "/tmp/app.deb")
	joined := strings.Join(args, " ")
	assert.NotContains(t, joined, passphrase)
	assert.Equal(t, "-k", args[0])
	assert.Equal(t, signing.keyID, args[1], "dpkg-sig takes a key ID, not a file path")
	assert.NotContains(t, joined, keyFile)
	assert.Contains(t, joined, "--passphrase-file "+signing.passphraseFile)
	assert.Equal(t, "/tmp/app.deb", args[len(args)-1])
}

func TestLegacyLinuxSigningImportsKeyFileIntoPrivateHome(t *testing.T) {
	keyFile := writeTestPGPKey(t, t.TempDir(), "release.asc", "")
	signing, cleanup, err := prepareLinuxPGPSigning(keyFile, "")
	require.NoError(t, err)
	home := signing.homedir
	require.NotEmpty(t, home)
	assert.Len(t, signing.keyID, 40, "a key file is signed by its fingerprint")
	assert.Empty(t, signing.passphraseFile)
	cleanup()
	_, err = os.Stat(filepath.Join(home, "pubring.kbx"))
	assert.True(t, os.IsNotExist(err), "cleanup must remove the imported key home")
}
