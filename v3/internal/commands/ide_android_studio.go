package commands

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

// AndroidStudioOptions controls the persistent Gradle project written by wails3 android studio.
type AndroidStudioOptions struct {
	Target string `name:"target" description:"Target platform and architecture" default:"android/arm64"`
	NoOpen bool   `name:"no-open" description:"Refresh the Gradle project without opening Android Studio"`
}

// AndroidStudio refreshes the persistent Gradle project at .wails/ide/android
// from a wake development build and opens it in Android Studio.
func AndroidStudio(options *AndroidStudioOptions, arguments []string) error {
	profile, err := manifestProfile("", arguments)
	if err != nil {
		return err
	}
	goos, arch, err := splitTarget(options.Target)
	if err != nil {
		return err
	}
	if goos != "android" {
		return fmt.Errorf("wails3 android studio requires an android/<arch> target, got %q", options.Target)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	loaded, err := manifest.Load(".", profile)
	if err != nil {
		return err
	}
	run, err := ideRunOptions(ctx, loaded, profile, goos, arch, []string{"apk"}, profile == "")
	if err != nil {
		return err
	}
	build, err := runManifestPipelineResult(run)
	if err != nil {
		return err
	}
	pkg, err := ideFindPackage(build, goos, "apk")
	if err != nil {
		return err
	}
	root := loaded.Config.Root
	if err := ensureWailsStateDirectory(root); err != nil {
		return err
	}
	project := filepath.Join(root, ".wails", "ide", "android")
	result, err := materialiseIDEProject(ideMaterialiseOptions{
		Source:      filepath.Join(root, filepath.FromSlash(pkg.Assets), "android"),
		Destination: project,
		Prepare: func(staged string) error {
			return prepareAndroidStudioProject(staged, root, pkg)
		},
	})
	if err != nil {
		return err
	}
	reportIDEMaterialised(result)
	if options.NoOpen {
		pterm.Success.Printfln("Gradle project ready: %s", result.Path)
		return nil
	}
	opened, err := openAndroidStudio(ideOpener{goos: runtime.GOOS, lookPath: exec.LookPath, stat: os.Stat, run: runIDECommand}, result.Path)
	if err != nil {
		return err
	}
	if !opened {
		pterm.Warning.Printfln("Android Studio was not found. Install it from https://developer.android.com/studio, then open %s", result.Path)
		return nil
	}
	pterm.Success.Printfln("Opened Android Studio project: %s", result.Path)
	return nil
}

// prepareAndroidStudioProject adds the wake-built native libraries to the
// staged Gradle project under app/src/main/jniLibs, matching the packaging
// step, and makes the Gradle wrapper executable.
func prepareAndroidStudioProject(staged, root string, pkg pipeline.PackageSpec) error {
	if pkg.Development {
		if err := prepareAndroidDevNetwork(staged); err != nil {
			return err
		}
	}
	binaries := pkg.Binaries
	if len(binaries) == 0 {
		binaries = []pipeline.ComponentBinary{{Arch: pkg.TargetArch, Path: pkg.Binary}}
	}
	for _, binary := range binaries {
		abi, err := androidABI(binary.Arch)
		if err != nil {
			return err
		}
		jni := filepath.Join(staged, "app", "src", "main", "jniLibs", abi, "libwails.so")
		if err := copyManifestPath(filepath.Join(root, filepath.FromSlash(binary.Path)), jni); err != nil {
			return fmt.Errorf("copy %s native library: %w", abi, err)
		}
	}
	if err := os.Chmod(filepath.Join(staged, "gradlew"), 0o755); err != nil {
		return err
	}
	return writeAndroidLocalProperties(staged)
}

// writeAndroidLocalProperties records the Android SDK location so Studio does
// not have to discover it. A local.properties already present in the staged
// copy came from the user through preserveIDEUserState and is never replaced.
func writeAndroidLocalProperties(staged string) error {
	target := filepath.Join(staged, "local.properties")
	if _, err := os.Lstat(target); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	home, _ := os.UserHomeDir()
	sdk := androidSDKDir(os.Getenv, runtime.GOOS, home, dirExists)
	if sdk == "" {
		return nil
	}
	return os.WriteFile(target, []byte(androidLocalPropertiesContent(sdk)), 0o644)
}

// androidSDKDir resolves the SDK location with the same precedence the
// pipeline uses (ANDROID_HOME, then ANDROID_SDK_ROOT), falling back to the
// platform default. It returns "" when no existing directory is found.
func androidSDKDir(getenv func(string) string, goos, home string, exists func(string) bool) string {
	for _, candidate := range []string{getenv("ANDROID_HOME"), getenv("ANDROID_SDK_ROOT")} {
		if candidate != "" && exists(candidate) {
			return candidate
		}
	}
	var fallback string
	switch goos {
	case "darwin":
		fallback = filepath.Join(home, "Library", "Android", "sdk")
	case "windows":
		if local := getenv("LOCALAPPDATA"); local != "" {
			fallback = filepath.Join(local, "Android", "Sdk")
		}
	default:
		fallback = filepath.Join(home, "Android", "Sdk")
	}
	if fallback != "" && exists(fallback) {
		return fallback
	}
	return ""
}

// androidLocalPropertiesContent renders local.properties with the sdk.dir
// value escaped by the Java properties rules.
func androidLocalPropertiesContent(sdk string) string {
	return "sdk.dir=" + escapeJavaProperty(sdk) + "\n"
}

// escapeJavaProperty escapes backslashes, colons, equals signs and '#' or '!'
// so Windows paths such as C:\Users survive a properties parser.
func escapeJavaProperty(value string) string {
	var builder strings.Builder
	for _, r := range value {
		switch r {
		case '\\', ':', '=', '#', '!':
			builder.WriteRune('\\')
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
