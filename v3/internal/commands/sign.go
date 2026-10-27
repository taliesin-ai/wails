package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/wailsapp/wails/v3/internal/defaults"
	"github.com/wailsapp/wails/v3/internal/flags"
	"github.com/wailsapp/wails/v3/internal/keychain"
)

// resolveSigningDefaults fills in any signing options not supplied via flags from
// the global config (~/.config/wails/defaults.yaml). This makes the signing
// configuration written by `wails3 setup` take effect at sign time for every
// platform — flags (and therefore project Taskfile vars) still win when present.
func resolveSigningDefaults(options *flags.Sign) {
	cfg, err := defaults.Load()
	if err != nil {
		return
	}

	// macOS
	if options.Identity == "" && cfg.Signing.Darwin.Identity != "" {
		options.Identity = cfg.Signing.Darwin.Identity
	}
	if options.KeychainProfile == "" && cfg.Signing.Darwin.KeychainProfile != "" {
		options.KeychainProfile = cfg.Signing.Darwin.KeychainProfile
	}
	if options.Entitlements == "" && cfg.Signing.Darwin.Entitlements != "" {
		options.Entitlements = cfg.Signing.Darwin.Entitlements
	}

	// Windows
	if options.Certificate == "" && cfg.Signing.Windows.CertificatePath != "" {
		options.Certificate = cfg.Signing.Windows.CertificatePath
	}
	if options.Thumbprint == "" && cfg.Signing.Windows.Thumbprint != "" {
		options.Thumbprint = cfg.Signing.Windows.Thumbprint
	}
	if options.Timestamp == "" && cfg.Signing.Windows.TimestampServer != "" {
		options.Timestamp = cfg.Signing.Windows.TimestampServer
	}

	// Linux (PGP). The signer needs an exported key *file*, so use the stored
	// key path; the bare key ID is not usable on its own.
	if options.PGPKey == "" && cfg.Signing.Linux.GPGKeyPath != "" {
		options.PGPKey = cfg.Signing.Linux.GPGKeyPath
	}
	if options.Role == "" && cfg.Signing.Linux.SignRole != "" {
		options.Role = cfg.Signing.Linux.SignRole
	}
}

// Sign signs a binary or package
func Sign(options *flags.Sign) error {
	if options.Input == "" {
		return fmt.Errorf("--input is required")
	}

	// Check input file exists
	info, err := os.Stat(options.Input)
	if err != nil {
		return fmt.Errorf("input file not found: %w", err)
	}

	// Backfill any unset options from the global signing config so the values
	// configured via `wails3 setup` are honored across all platforms.
	resolveSigningDefaults(options)

	// Determine what type of signing to do based on file extension and flags
	ext := strings.ToLower(filepath.Ext(options.Input))

	// macOS app bundle (directory)
	if info.IsDir() && strings.HasSuffix(options.Input, ".app") {
		return signMacOSApp(options)
	}
	if ext == ".dmg" {
		return signMacOSDiskImage(options)
	}

	// macOS binary or Windows executable
	if ext == ".exe" || ext == ".msi" || ext == ".msix" || ext == ".appx" {
		return signWindows(options)
	}

	// Linux packages
	if ext == ".deb" {
		return signDEB(options)
	}
	if ext == ".rpm" {
		return signRPM(options)
	}

	// macOS binary (no extension typically)
	if runtime.GOOS == "darwin" && options.Identity != "" {
		return signMacOSBinary(options)
	}

	return fmt.Errorf("unsupported file type: %s", ext)
}

func signMacOSDiskImage(options *flags.Sign) error {
	resolveSigningDefaults(options)
	if options.Identity == "" {
		return fmt.Errorf("--identity is required for macOS disk image signing (set via `wails3 setup` or --identity flag)")
	}
	args := []string{"--force", "--sign", options.Identity, options.Input}
	cmd := exec.Command("codesign", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("codesign disk image failed: %w", err)
	}
	pterm.Success.Printfln("Signed: %s", options.Input)
	if !options.Notarize {
		return nil
	}
	if options.KeychainProfile == "" {
		return fmt.Errorf("--keychain-profile is required for notarization (set via `wails3 setup` or --keychain-profile flag)")
	}
	submit := exec.Command("xcrun", "notarytool", "submit", options.Input, "--keychain-profile", options.KeychainProfile, "--wait")
	submit.Stdout = os.Stdout
	submit.Stderr = os.Stderr
	if err := submit.Run(); err != nil {
		return fmt.Errorf("disk image notarization failed: %w", err)
	}
	staple := exec.Command("xcrun", "stapler", "staple", options.Input)
	staple.Stdout = os.Stdout
	staple.Stderr = os.Stderr
	if err := staple.Run(); err != nil {
		return fmt.Errorf("disk image stapling failed: %w", err)
	}
	pterm.Success.Println("Notarization complete and ticket stapled")
	return nil
}

func signMacOSApp(options *flags.Sign) error {
	resolveSigningDefaults(options)
	if options.Identity == "" {
		return fmt.Errorf("--identity is required for macOS signing (set via `wails3 setup` or --identity flag)")
	}

	if options.Verbose {
		pterm.Info.Printfln("Signing macOS app bundle: %s", options.Input)
	}

	// Build codesign command
	args := []string{
		"--force",
		"--deep",
		"--sign", options.Identity,
	}

	if options.Entitlements != "" {
		args = append(args, "--entitlements", options.Entitlements)
	}

	if options.HardenedRuntime || options.Notarize {
		args = append(args, "--options", "runtime")
	}

	args = append(args, options.Input)

	cmd := exec.Command("codesign", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("codesign failed: %w", err)
	}

	pterm.Success.Printfln("Signed: %s", options.Input)

	// Notarize if requested
	if options.Notarize {
		return notarizeMacOSApp(options)
	}

	return nil
}

func signMacOSBinary(options *flags.Sign) error {
	resolveSigningDefaults(options)
	if options.Identity == "" {
		return fmt.Errorf("--identity is required for macOS signing (set via `wails3 setup` or --identity flag)")
	}

	if options.Verbose {
		pterm.Info.Printfln("Signing macOS binary: %s", options.Input)
	}

	args := []string{
		"--force",
		"--sign", options.Identity,
	}

	if options.Entitlements != "" {
		args = append(args, "--entitlements", options.Entitlements)
	}

	if options.HardenedRuntime {
		args = append(args, "--options", "runtime")
	}

	args = append(args, options.Input)

	cmd := exec.Command("codesign", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("codesign failed: %w", err)
	}

	pterm.Success.Printfln("Signed: %s", options.Input)
	return nil
}

func notarizeMacOSApp(options *flags.Sign) error {
	if options.KeychainProfile == "" {
		return fmt.Errorf("--keychain-profile is required for notarization (set via `wails3 setup` or --keychain-profile flag)")
	}

	if options.Verbose {
		pterm.Info.Println("Submitting for notarization...")
	}

	// Create a zip for notarization
	zipPath := options.Input + ".zip"
	zipCmd := exec.Command("ditto", "-c", "-k", "--keepParent", options.Input, zipPath)
	if err := zipCmd.Run(); err != nil {
		return fmt.Errorf("failed to create zip for notarization: %w", err)
	}
	defer os.Remove(zipPath)

	// Submit for notarization
	args := []string{
		"notarytool", "submit",
		zipPath,
		"--keychain-profile", options.KeychainProfile,
		"--wait",
	}

	cmd := exec.Command("xcrun", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("notarization failed: %w", err)
	}

	// Staple the ticket
	stapleCmd := exec.Command("xcrun", "stapler", "staple", options.Input)
	stapleCmd.Stdout = os.Stdout
	stapleCmd.Stderr = os.Stderr

	if err := stapleCmd.Run(); err != nil {
		return fmt.Errorf("stapling failed: %w", err)
	}

	pterm.Success.Println("Notarization complete and ticket stapled")
	return nil
}

func signWindows(options *flags.Sign) error {
	// Get password from keychain if not provided
	password := options.Password
	if password == "" && options.Certificate != "" {
		var err error
		password, err = keychain.Get(keychain.KeyWindowsCertPassword)
		if err != nil {
			pterm.Warning.Printfln("Could not get password from keychain: %v", err)
			// Continue without password - might work for some certificates
		}
	}

	if options.Verbose {
		pterm.Info.Printfln("Signing Windows executable: %s", options.Input)
	}

	// Try native signtool first on Windows
	if runtime.GOOS == "windows" {
		if _, err := findSigntool(); err == nil {
			// A present signer rejected this request. Preserve its diagnostic;
			// an unrelated fallback must not hide a bad password or certificate.
			return signWindowsNative(options, password)
		}
	}

	// Use built-in signing (works cross-platform)
	return signWindowsBuiltin(options, password)
}

func signWindowsNative(options *flags.Sign, password string) error {
	// Find signtool.exe
	signtool, err := findSigntool()
	if err != nil {
		return err
	}

	args := []string{"sign"}

	if options.Certificate != "" {
		args = append(args, "/f", options.Certificate)
		if password != "" {
			args = append(args, "/p", password)
		}
	} else if options.Thumbprint != "" {
		args = append(args, "/sha1", options.Thumbprint)
	} else {
		return fmt.Errorf("either --certificate or --thumbprint is required")
	}

	// Add timestamp server
	timestamp := options.Timestamp
	if timestamp == "" {
		timestamp = "http://timestamp.digicert.com"
	}
	args = append(args, "/tr", timestamp, "/td", "SHA256", "/fd", "SHA256")

	args = append(args, options.Input)

	cmd := exec.Command(signtool, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if password != "" {
			detail = strings.ReplaceAll(detail, password, "<redacted>")
		}
		return fmt.Errorf("signtool failed: %s: %w", detail, err)
	}

	pterm.Success.Printfln("Signed: %s", options.Input)
	return nil
}

func findSigntool() (string, error) {
	// Check if signtool is in PATH
	path, err := exec.LookPath("signtool.exe")
	if err == nil {
		return path, nil
	}

	// Common Windows SDK locations
	sdkPaths := []string{
		`C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x64\signtool.exe`,
		`C:\Program Files (x86)\Windows Kits\10\bin\10.0.19041.0\x64\signtool.exe`,
		`C:\Program Files (x86)\Windows Kits\10\bin\x64\signtool.exe`,
	}

	for _, p := range sdkPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("signtool.exe not found")
}

func signWindowsBuiltin(options *flags.Sign, password string) error {
	// This would use a Go library for Authenticode signing
	// For now, we'll return an error indicating it needs implementation
	// In a full implementation, you'd use something like:
	// - github.com/AkarinLiu/osslsigncode-go
	// - or implement PE signing directly

	if options.Certificate == "" {
		return fmt.Errorf("--certificate is required for cross-platform signing")
	}

	return fmt.Errorf("built-in Windows signing not yet implemented - please use signtool.exe on Windows, or install osslsigncode")
}

// linuxPGPSigning is the key material a legacy dpkg-sig or rpmsign run needs.
// The key is referenced by key ID and the passphrase by file, so the passphrase
// never appears in argv or the environment.
type linuxPGPSigning struct {
	keyID          string // fingerprint or key ID to sign with
	homedir        string // GNUPG home holding an imported key file; empty uses the default keyring
	passphraseFile string // 0600 file holding the passphrase; empty when none is needed
}

// prepareLinuxPGPSigning resolves a key file or key ID. A key file is imported into
// a private GNUPG home so the signature can be made by fingerprint. The returned
// cleanup removes every temporary file and directory it created.
func prepareLinuxPGPSigning(pgpKey, password string) (linuxPGPSigning, func(), error) {
	var created []string
	cleanup := func() {
		for _, path := range created {
			_ = os.RemoveAll(path)
		}
	}
	signing := linuxPGPSigning{keyID: pgpKey}
	if info, err := os.Stat(pgpKey); err == nil && !info.IsDir() {
		home, err := os.MkdirTemp("", "wails-gnupg-*")
		if err != nil {
			return linuxPGPSigning{}, cleanup, err
		}
		created = append(created, home)
		if output, err := exec.Command("gpg", "--homedir", home, "--batch", "--import", pgpKey).CombinedOutput(); err != nil {
			return linuxPGPSigning{}, cleanup, fmt.Errorf("import PGP key %s: %w: %s", pgpKey, err, strings.TrimSpace(string(output)))
		}
		output, err := exec.Command("gpg", "--homedir", home, "--batch", "--with-colons", "--list-secret-keys").Output()
		if err != nil {
			return linuxPGPSigning{}, cleanup, fmt.Errorf("list imported PGP secret keys: %w", err)
		}
		fingerprint := firstGPGFingerprint(string(output))
		if fingerprint == "" {
			return linuxPGPSigning{}, cleanup, fmt.Errorf("PGP key file %s contains no secret key", pgpKey)
		}
		signing.keyID = fingerprint
		signing.homedir = home
	}
	if password != "" {
		file, err := os.CreateTemp("", "wails-pgp-passphrase-*")
		if err != nil {
			return linuxPGPSigning{}, cleanup, err
		}
		created = append(created, file.Name())
		// CreateTemp opens the file 0600; the passphrase is written once and read by gpg only.
		if _, err := file.WriteString(password); err != nil {
			file.Close()
			return linuxPGPSigning{}, cleanup, err
		}
		if err := file.Close(); err != nil {
			return linuxPGPSigning{}, cleanup, err
		}
		signing.passphraseFile = file.Name()
	}
	return signing, cleanup, nil
}

// firstGPGFingerprint returns the first fingerprint in `gpg --with-colons` output.
func firstGPGFingerprint(colons string) string {
	for _, line := range strings.Split(colons, "\n") {
		fields := strings.Split(line, ":")
		if len(fields) > 9 && fields[0] == "fpr" {
			return fields[9]
		}
	}
	return ""
}

// gpgOptions returns the gpg options that make the signing non-interactive. The
// passphrase is read from its file, never from the command line.
func (s linuxPGPSigning) gpgOptions() string {
	options := []string{}
	if s.homedir != "" {
		options = append(options, "--homedir", s.homedir)
	}
	if s.passphraseFile != "" {
		options = append(options, "--batch", "--pinentry-mode", "loopback", "--passphrase-file", s.passphraseFile)
	}
	return strings.Join(options, " ")
}

// rpmsignArgs builds the rpmsign arguments for input. The key and passphrase are
// supplied as rpm macros, so no secret is placed on the command line.
func rpmsignArgs(s linuxPGPSigning, input string) []string {
	args := []string{"--define", "_gpg_name " + s.keyID}
	if s.homedir != "" {
		args = append(args, "--define", "_gpg_path "+s.homedir)
	}
	if s.passphraseFile != "" {
		args = append(args, "--define", "_gpg_sign_cmd_extra_args --batch --pinentry-mode loopback --passphrase-file "+s.passphraseFile)
	}
	return append(args, "--addsign", input)
}

// debsignArgs builds the dpkg-sig arguments for a role and input.
func debsignArgs(s linuxPGPSigning, role, input string) []string {
	args := []string{"-k", s.keyID, "--sign", role}
	if options := s.gpgOptions(); options != "" {
		args = append(args, "--gpg-options", options)
	}
	return append(args, input)
}

// linuxSignerCommand runs a signing tool, pointing gpg at the imported key home
// when there is one.
func linuxSignerCommand(s linuxPGPSigning, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if s.homedir != "" {
		cmd.Env = append(os.Environ(), "GNUPGHOME="+s.homedir)
	}
	return cmd
}

func signDEB(options *flags.Sign) error {
	if options.PGPKey == "" {
		return fmt.Errorf("no PGP signing key found for DEB signing — configure linux.signing in wails.yaml, pass --pgp-key, or run `wails3 setup` / `wails3 setup signing`")
	}

	// Get password from keychain if not provided
	password := options.PGPPassword
	if password == "" {
		var err error
		password, err = keychain.Get(keychain.KeyPGPPassword)
		if err != nil {
			// Password might not be required if key is unencrypted
			if options.Verbose {
				pterm.Warning.Printfln("Could not get PGP password from keychain: %v", err)
			}
		}
	}

	role := options.Role
	if role == "" {
		role = "builder"
	}

	signing, cleanup, err := prepareLinuxPGPSigning(options.PGPKey, password)
	defer cleanup()
	if err != nil {
		return err
	}

	if options.Verbose {
		pterm.Info.Printfln("Signing DEB package: %s", options.Input)
	}

	if _, err := exec.LookPath("dpkg-sig"); err == nil {
		cmd := linuxSignerCommand(signing, "dpkg-sig", debsignArgs(signing, role, options.Input)...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("dpkg-sig failed: %w", err)
		}
		pterm.Success.Printfln("Signed: %s", options.Input)
		return nil
	}
	// dpkg-sig is no longer packaged on current Debian and Fedora. debsigs is its
	// maintained replacement; it cannot take a passphrase, so only unprotected keys
	// (or keys unlocked in gpg-agent) can use it.
	if _, err := exec.LookPath("debsigs"); err == nil {
		if password != "" {
			return fmt.Errorf("debsigs cannot use a passphrase-protected key; sign with wake instead (linux.signing in wails.yaml), which signs deb packages natively with nfpm")
		}
		args := []string{"--sign=" + role, "-k", signing.keyID, options.Input}
		if err := linuxSignerCommand(signing, "debsigs", args...).Run(); err != nil {
			return fmt.Errorf("debsigs failed: %w", err)
		}
		pterm.Success.Printfln("Signed: %s", options.Input)
		return nil
	}
	return fmt.Errorf("neither dpkg-sig nor debsigs is installed; install one, or sign the deb with wake (linux.signing in wails.yaml), which signs deb packages natively with nfpm")
}

func signRPM(options *flags.Sign) error {
	if options.PGPKey == "" {
		return fmt.Errorf("no PGP signing key found for RPM signing — configure linux.signing in wails.yaml, pass --pgp-key, or run `wails3 setup` / `wails3 setup signing`")
	}

	// Get password from keychain if not provided
	password := options.PGPPassword
	if password == "" {
		var err error
		password, err = keychain.Get(keychain.KeyPGPPassword)
		if err != nil {
			if options.Verbose {
				pterm.Warning.Printfln("Could not get PGP password from keychain: %v", err)
			}
		}
	}

	signing, cleanup, err := prepareLinuxPGPSigning(options.PGPKey, password)
	defer cleanup()
	if err != nil {
		return err
	}

	if options.Verbose {
		pterm.Info.Printfln("Signing RPM package: %s", options.Input)
	}

	if err := linuxSignerCommand(signing, "rpmsign", rpmsignArgs(signing, options.Input)...).Run(); err != nil {
		return fmt.Errorf("rpmsign failed: %w", err)
	}

	pterm.Success.Printfln("Signed: %s", options.Input)
	return nil
}
