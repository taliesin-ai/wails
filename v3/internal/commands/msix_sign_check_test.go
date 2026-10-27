package commands

import (
	"archive/zip"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/pkcs12"
)

const msixSignCheckManifest = `<?xml version="1.0" encoding="utf-8"?>
<Package><Identity Name="com.example.app" Publisher="CN=Wails Wake Test" Version="1.0.0.0"/></Package>`

func writeTestMSIX(t *testing.T, manifestXML string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "app.msix")
	file, err := os.Create(path)
	require.NoError(t, err)
	writer := zip.NewWriter(file)
	entry, err := writer.Create("AppxManifest.xml")
	require.NoError(t, err)
	_, err = entry.Write([]byte(manifestXML))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	require.NoError(t, file.Close())
	return path
}

// writeTestPFX creates a self-signed certificate and exports it as a PFX with
// the openssl CLI. It skips when openssl is unavailable or its output uses an
// PKCS#12 encryption this package cannot decode.
func writeTestPFX(t *testing.T, subject string) (pfxPath, password string) {
	t.Helper()
	openssl, err := exec.LookPath("openssl")
	if err != nil {
		t.Skip("openssl not available to generate a PFX")
	}
	dir := t.TempDir()
	key := filepath.Join(dir, "key.pem")
	cert := filepath.Join(dir, "cert.pem")
	pfxPath = filepath.Join(dir, "cert.pfx")
	password = "test-pass"
	if out, err := exec.Command(openssl, "req", "-x509", "-newkey", "ec", "-pkeyopt", "ec_paramgen_curve:prime256v1", "-nodes", "-keyout", key, "-out", cert, "-days", "1", "-subj", "/"+subject).CombinedOutput(); err != nil {
		t.Skipf("openssl req failed: %v: %s", err, out)
	}
	// OpenSSL 3 defaults to AES-based PKCS#12 encryption that golang.org/x/crypto/pkcs12
	// cannot read; -legacy selects the RC2/3DES profile the decoder supports.
	for _, extra := range [][]string{{"-legacy"}, {}} {
		args := append([]string{"pkcs12", "-export", "-inkey", key, "-in", cert, "-out", pfxPath, "-passout", "pass:" + password}, extra...)
		if out, err := exec.Command(openssl, args...).CombinedOutput(); err != nil {
			t.Logf("openssl pkcs12 %v failed: %v: %s", extra, err, out)
			continue
		}
		data, err := os.ReadFile(pfxPath)
		require.NoError(t, err)
		if _, _, err := pkcs12.Decode(data, password); err == nil {
			return pfxPath, password
		}
	}
	t.Skip("no openssl PKCS#12 profile that golang.org/x/crypto/pkcs12 can decode")
	return "", ""
}

func TestMSIXPackagePublisherReadsAppxManifest(t *testing.T) {
	publisher, err := msixPackagePublisher(writeTestMSIX(t, msixSignCheckManifest))
	require.NoError(t, err)
	assert.Equal(t, "CN=Wails Wake Test", publisher)

	_, err = msixPackagePublisher(writeTestMSIX(t, `<Package><Identity Name="x"/></Package>`))
	assert.ErrorContains(t, err, "no Identity Publisher")
}

func TestVerifyMSIXPublisherRejectsMismatchBeforeSigning(t *testing.T) {
	pfxPath, password := writeTestPFX(t, "CN=Certificate Holder")
	pkg := writeTestMSIX(t, msixSignCheckManifest)

	err := verifyMSIXPublisher(pkg, pfxPath, password, "")
	require.Error(t, err)
	assert.ErrorContains(t, err, `"CN=Wails Wake Test"`)
	assert.ErrorContains(t, err, `"CN=Certificate Holder"`)
}

func TestVerifyMSIXPublisherAcceptsMatchingCertificate(t *testing.T) {
	pfxPath, password := writeTestPFX(t, "CN=Wails Wake Test")
	pkg := writeTestMSIX(t, msixSignCheckManifest)
	assert.NoError(t, verifyMSIXPublisher(pkg, pfxPath, password, ""))
}

func TestVerifyMSIXPublisherDefersWhenSubjectUnknown(t *testing.T) {
	// No certificate and no thumbprint: nothing to compare, signtool decides.
	pkg := writeTestMSIX(t, msixSignCheckManifest)
	assert.NoError(t, verifyMSIXPublisher(pkg, "", "", ""))
}

func TestVerifyMSIXPublisherDefersWhenPackageUnreadable(t *testing.T) {
	pkg := filepath.Join(t.TempDir(), "app.msix")
	require.NoError(t, os.WriteFile(pkg, []byte("not a zip"), 0o644))
	assert.NoError(t, verifyMSIXPublisher(pkg, "", "", ""))
}
