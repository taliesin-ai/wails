package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeTestPGPKey creates a throwaway ASCII-armoured secret key in dir and returns
// its path. The key lives in a private GNUPG home that is torn down with the test.
// Tests that need it are skipped when gpg is not installed.
func writeTestPGPKey(t *testing.T, dir, name, passphrase string) string {
	t.Helper()
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg is required to create a test PGP key")
	}
	home, err := os.MkdirTemp("", "wgpg")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = exec.Command("gpgconf", "--homedir", home, "--kill", "gpg-agent").Run()
		_ = os.RemoveAll(home)
	})
	t.Setenv("GNUPGHOME", home)
	generate := []string{"--batch", "--pinentry-mode", "loopback", "--passphrase", passphrase, "--quick-gen-key", "Wails Test <test@example.com>", "default", "default", "never"}
	require.NoError(t, exec.Command("gpg", generate...).Run())
	exported, err := exec.Command("gpg", "--batch", "--pinentry-mode", "loopback", "--passphrase", passphrase, "--armor", "--export-secret-keys", "test@example.com").Output()
	require.NoError(t, err)
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, exported, 0o600))
	return path
}
