package commands

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
	"golang.org/x/crypto/pkcs12"
)

var certificateThumbprintPattern = regexp.MustCompile(`^[0-9A-Fa-f]+$`)

// verifyMSIXPublisher fails before signtool runs when the Publisher packaged in
// AppxManifest.xml differs from the signing certificate subject. signtool would
// otherwise fail later with an opaque SignerSign() error. When either side
// cannot be read on this host, the check is skipped with a warning and signtool
// remains the authority.
func verifyMSIXPublisher(packagePath, certificate, password, thumbprint string) error {
	publisher, err := msixPackagePublisher(packagePath)
	if err != nil {
		pterm.Warning.Printfln("Could not read the MSIX publisher from %s; signtool will validate it: %v", filepath.Base(packagePath), err)
		return nil
	}
	subject, err := signingCertificateSubject(certificate, password, thumbprint)
	if err != nil {
		return fmt.Errorf("read MSIX signing certificate: %w", err)
	}
	if subject == "" {
		pterm.Warning.Printfln("Could not read the signing certificate subject; signtool will check the MSIX publisher %q", publisher)
		return nil
	}
	if !manifest.MSIXPublisherMatchesSubject(publisher, subject) {
		return fmt.Errorf("MSIX Publisher %q in AppxManifest.xml must equal the signing certificate subject %q; set packages.msix.publisher (or the Publisher in the user-owned manifest) to the certificate subject", publisher, subject)
	}
	return nil
}

// msixPackagePublisher returns the Identity Publisher from the AppxManifest.xml
// inside a packaged .msix.
func msixPackagePublisher(packagePath string) (string, error) {
	reader, err := zip.OpenReader(packagePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	for _, file := range reader.File {
		if !strings.EqualFold(file.Name, "AppxManifest.xml") {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", err
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return "", err
		}
		publisher, _, found, err := manifest.MSIXManifestPublisher(data)
		if err != nil {
			return "", fmt.Errorf("AppxManifest.xml: %w", err)
		}
		if !found {
			return "", fmt.Errorf("AppxManifest.xml has no Identity Publisher")
		}
		return publisher, nil
	}
	return "", fmt.Errorf("package has no AppxManifest.xml")
}

// signingCertificateSubject returns the subject of the certificate that will
// sign the package. It returns "" with no error when the subject cannot be
// determined here: a PFX using encryption this host cannot decode, a wrong
// password (signtool reports that itself), or a thumbprint not in the store.
func signingCertificateSubject(certificate, password, thumbprint string) (string, error) {
	if certificate != "" {
		data, err := os.ReadFile(certificate)
		if err != nil {
			return "", err
		}
		_, cert, err := pkcs12.Decode(data, password)
		if err != nil || cert == nil {
			return "", nil
		}
		return cert.Subject.String(), nil
	}
	if thumbprint != "" {
		return certificateSubjectFromStore(thumbprint)
	}
	return "", nil
}

// certificateSubjectFromStore looks a certificate up by thumbprint in the
// current-user and local-machine personal stores. It only runs on Windows.
func certificateSubjectFromStore(thumbprint string) (string, error) {
	if runtime.GOOS != "windows" {
		return "", nil
	}
	if !certificateThumbprintPattern.MatchString(thumbprint) {
		return "", fmt.Errorf("signing.windows.thumbprint %q is not a hexadecimal thumbprint", thumbprint)
	}
	script := fmt.Sprintf(`Get-ChildItem -Path Cert:\CurrentUser\My, Cert:\LocalMachine\My | Where-Object { $_.Thumbprint -eq '%s' } | Select-Object -First 1 -ExpandProperty Subject`, thumbprint)
	output, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}
