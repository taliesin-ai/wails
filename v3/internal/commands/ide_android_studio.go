package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"

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
	return nil
}
