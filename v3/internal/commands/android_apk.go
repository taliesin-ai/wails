package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/internal/term"
	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

// androidDebugSignWarning is reported whenever a release APK is signed with the
// Android debug keystore. Such an APK installs, but it is not a distributable
// build: the debug key is public and cannot be used for distribution.
const androidDebugSignWarning = "Android release APK was signed with the debug keystore (~/.android/debug.keystore) because no release signing ran in this build. It is installable for testing but NOT suitable for distribution. Run `wails3 sign` with signing.android configured to produce a distributable APK."

// needsAndroidDebugSign reports whether a packaged artifact must be signed with
// the debug keystore. Unsigned release APKs cannot be installed, so this is the
// only way a production APK is produced when no sign stage runs.
func needsAndroidDebugSign(s pipeline.PackageSpec) bool {
	return s.Format == "apk" && s.Release && s.DebugSign
}

// debugKeystoreSignArgs are the apksigner arguments for the Android debug key.
// The password is the well-known Android tooling default, not a secret.
func debugKeystoreSignArgs(keystore string) []string {
	return []string{"--ks", keystore, "--ks-key-alias", "androiddebugkey", "--ks-pass", "pass:android", "--key-pass", "pass:android"}
}

// ensureAndroidDebugKeystore creates the debug keystore exactly as the Android
// tooling does when it is missing.
func ensureAndroidDebugKeystore(ctx context.Context, keystore string) error {
	if _, err := os.Stat(keystore); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(keystore), 0o755); err != nil {
		return err
	}
	output, err := runManifestCommand(ctx, filepath.Dir(keystore), nil, "keytool", "-genkeypair", "-v",
		"-keystore", keystore, "-storepass", "android", "-alias", "androiddebugkey", "-keypass", "android",
		"-keyalg", "RSA", "-keysize", "2048", "-validity", "10000", "-dname", "CN=Android Debug,O=Android,C=US")
	if err != nil {
		return fmt.Errorf("create Android debug keystore %s: %w: %s", keystore, err, output)
	}
	return nil
}

// alignAndSignAPK zipaligns input and then signs the aligned copy with apksigner.
// apksigner requires the zipalign step to run first for a valid release APK.
func alignAndSignAPK(ctx context.Context, dir, input, output string, signArgs []string) (string, error) {
	zipalign, ok := pipeline.AndroidBuildToolPath("zipalign")
	if !ok {
		return "", fmt.Errorf("zipalign not found; install Android SDK build-tools and set ANDROID_HOME")
	}
	apksigner, ok := pipeline.AndroidBuildToolPath("apksigner")
	if !ok {
		return "", fmt.Errorf("apksigner not found; install Android SDK build-tools and set ANDROID_HOME")
	}
	aligning, err := os.MkdirTemp("", "wails-apk-align-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(aligning)
	aligned := filepath.Join(aligning, "aligned.apk")
	if output, err := runManifestCommand(ctx, dir, nil, zipalign, "-f", "-p", "4", input, aligned); err != nil {
		return output, fmt.Errorf("zipalign: %w", err)
	}
	args := append([]string{"sign"}, signArgs...)
	args = append(args, "--out", output, aligned)
	return runManifestCommand(ctx, dir, nil, apksigner, args...)
}

// signReleaseAPKWithDebugKeystore produces a debug-signed, aligned copy of the
// unsigned release APK and reports the warning through the build session.
func signReleaseAPKWithDebugKeystore(ctx context.Context, dir, input, output string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	keystore := filepath.Join(home, ".android", "debug.keystore")
	if err := ensureAndroidDebugKeystore(ctx, keystore); err != nil {
		return err
	}
	if output, err := alignAndSignAPK(ctx, dir, input, output, debugKeystoreSignArgs(keystore)); err != nil {
		return fmt.Errorf("sign release APK with debug keystore: %w: %s", err, output)
	}
	// The wake pipeline reporter has no warning channel, and the buildwarnings file
	// is only set up by the legacy task wrapper, so print directly.
	term.Warning(androidDebugSignWarning)
	return nil
}
