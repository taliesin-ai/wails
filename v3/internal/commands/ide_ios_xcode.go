package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/pterm/pterm"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

// IOSXcodeOptions controls the persistent Xcode project written by wails3 ios xcode.
type IOSXcodeOptions struct {
	Target string `name:"target" description:"Target platform and architecture" default:"ios/arm64"`
	NoOpen bool   `name:"no-open" description:"Refresh the Xcode project without opening it"`
}

// Patterns that the generated project.pbxproj uses for its legacy Go build.
// The wake pipeline has already built the c-archive, so the copy links it from
// a path inside the project instead of rebuilding it from the legacy layout.
var (
	iosXcodeArchiveRef   = regexp.MustCompile(`path = "\.\./\.\./\.\./bin/[^"]*";`)
	iosXcodeBuildScript  = regexp.MustCompile(`shellScript = "(?:[^"\\]|\\.)*";`)
	iosXcodeSDKRoot      = regexp.MustCompile(`SDKROOT = iphoneos;`)
	iosXcodeBuildConfig  = regexp.MustCompile(`buildSettings = \{\n`)
	iosXcodePrebuiltNote = `shellScript = "set -e\necho \"Go c-archive is prebuilt by wails3 ios xcode\"\n";`
)

// IOSXcode refreshes the persistent Xcode project at .wails/ide/ios/xcode from
// a wake iOS build and opens it on macOS.
func IOSXcode(options *IOSXcodeOptions, arguments []string) error {
	profile, err := manifestProfile("", arguments)
	if err != nil {
		return err
	}
	goos, arch, err := splitTarget(options.Target)
	if err != nil {
		return err
	}
	if goos != "ios" {
		return fmt.Errorf("wails3 ios xcode requires an ios/<arch> target, got %q", options.Target)
	}
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("wails3 ios xcode requires macOS and Xcode")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	loaded, err := manifest.Load(".", profile)
	if err != nil {
		return err
	}
	run, err := ideRunOptions(ctx, loaded, profile, goos, arch, nil, false)
	if err != nil {
		return err
	}
	build, err := runManifestPipelineResult(run)
	if err != nil {
		return err
	}
	pkg, err := ideFindPackage(build, goos, "app")
	if err != nil {
		return err
	}
	root := loaded.Config.Root
	if err := ensureWailsStateDirectory(root); err != nil {
		return err
	}
	project := filepath.Join(root, ".wails", "ide", "ios", "xcode")
	archive := filepath.Join(root, filepath.FromSlash(pkg.Binary))
	result, err := materialiseIDEProject(ideMaterialiseOptions{
		Source:      filepath.Join(root, filepath.FromSlash(pkg.Assets), "ios", "xcode"),
		Destination: project,
		Prepare: func(staged string) error {
			return prepareIOSXcodeProject(staged, archive, loaded.Config.Signing.IOS.TeamID)
		},
	})
	if err != nil {
		return err
	}
	reportIDEMaterialised(result)
	xcodeproj := filepath.Join(result.Path, "main.xcodeproj")
	if options.NoOpen {
		pterm.Success.Printfln("Xcode project ready: %s", xcodeproj)
		return nil
	}
	if err := openXcodeProject(ideOpener{goos: runtime.GOOS, run: runIDECommand}, xcodeproj); err != nil {
		return err
	}
	pterm.Success.Printfln("Opened Xcode project: %s", xcodeproj)
	return nil
}

// prepareIOSXcodeProject links the wake-built c-archive into the staged project
// and rewrites the generated project to use it.
func prepareIOSXcodeProject(staged, archive, team string) error {
	libName := filepath.Base(archive)
	libDir := filepath.Join(staged, "lib")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		return err
	}
	if err := copyManifestPath(archive, filepath.Join(libDir, libName)); err != nil {
		return fmt.Errorf("copy Go c-archive: %w", err)
	}
	pbxprojPath := filepath.Join(staged, "main.xcodeproj", "project.pbxproj")
	source, err := os.ReadFile(pbxprojPath)
	if err != nil {
		return fmt.Errorf("read generated project: %w", err)
	}
	rewritten, err := rewriteIOSXcodeProject(source, libName, team)
	if err != nil {
		return err
	}
	info, err := os.Stat(pbxprojPath)
	if err != nil {
		return err
	}
	return os.WriteFile(pbxprojPath, rewritten, info.Mode().Perm())
}

// rewriteIOSXcodeProject points the archive reference at lib/<name>, replaces
// the legacy Go build phase with a no-op, and sets SDKROOT to auto so Xcode
// picks the simulator or device SDK from the chosen destination. When team is
// set, DEVELOPMENT_TEAM is added to every build configuration.
func rewriteIOSXcodeProject(source []byte, libName, team string) ([]byte, error) {
	if team != "" && !manifest.IsValidAppleTeamID(team) {
		return nil, fmt.Errorf("ios.signing.team_id %q is not a valid Apple team identifier", team)
	}
	if len(iosXcodeArchiveRef.FindAll(source, -1)) != 1 {
		return nil, fmt.Errorf("generated project.pbxproj has an unexpected Go archive reference")
	}
	if len(iosXcodeBuildScript.FindAll(source, -1)) == 0 {
		return nil, fmt.Errorf("generated project.pbxproj has no Go build phase")
	}
	if len(iosXcodeSDKRoot.FindAll(source, -1)) == 0 {
		return nil, fmt.Errorf("generated project.pbxproj has no SDKROOT setting")
	}
	output := iosXcodeArchiveRef.ReplaceAllLiteral(source, []byte(fmt.Sprintf(`path = "lib/%s";`, libName)))
	output = iosXcodeBuildScript.ReplaceAllLiteral(output, []byte(iosXcodePrebuiltNote))
	output = iosXcodeSDKRoot.ReplaceAllLiteral(output, []byte("SDKROOT = auto;"))
	if team != "" {
		if len(iosXcodeBuildConfig.FindAll(output, -1)) == 0 {
			return nil, fmt.Errorf("generated project.pbxproj has no build configurations for DEVELOPMENT_TEAM")
		}
		output = iosXcodeBuildConfig.ReplaceAllLiteral(output, []byte("buildSettings = {\n\t\t\t\tDEVELOPMENT_TEAM = "+team+";\n"))
	}
	return output, nil
}

// runIDECommand runs an external opener and folds its output into the error.
func runIDECommand(name string, args ...string) error {
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s: %w", name, string(output), err)
	}
	return nil
}

// reportIDEMaterialised prints the warning for a changed project.
func reportIDEMaterialised(result ideMaterialiseResult) {
	if result.UserModified {
		pterm.Warning.Printfln("%s changed since the last wails3 refresh; your version was kept at %s", result.Path, result.Backup)
	}
}
