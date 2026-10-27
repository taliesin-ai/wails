package commands

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

func writeIDETestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readIDETestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// fakeIDESource builds a generated tree shaped like the iOS assets output.
func fakeIDESource(t *testing.T, version string) string {
	t.Helper()
	source := t.TempDir()
	writeIDETestFile(t, filepath.Join(source, "main", "main.m"), "int main(void) { return 0; }\n")
	writeIDETestFile(t, filepath.Join(source, "main", "Info.plist"), "<plist>"+version+"</plist>\n")
	writeIDETestFile(t, filepath.Join(source, "main.xcodeproj", "project.pbxproj"), "project "+version+"\n")
	return source
}

func TestMaterialiseIDEProjectCopiesAndTracksSnapshot(t *testing.T) {
	source := fakeIDESource(t, "v1")
	destination := filepath.Join(t.TempDir(), "ide", "xcode")
	options := ideMaterialiseOptions{Source: source, Destination: destination, Prepare: func(staged string) error {
		writeIDETestFile(t, filepath.Join(staged, "lib", "app.a"), "archive-1")
		return nil
	}}

	result, err := materialiseIDEProject(options)
	if err != nil {
		t.Fatal(err)
	}
	if result.UserModified || result.Backup != "" {
		t.Fatalf("first refresh reported changes: %+v", result)
	}
	if got := readIDETestFile(t, filepath.Join(destination, "lib", "app.a")); got != "archive-1" {
		t.Fatalf("archive not published, got %q", got)
	}
	if got := readIDETestFile(t, filepath.Join(destination, "main.xcodeproj", "project.pbxproj")); got != "project v1\n" {
		t.Fatalf("generated project not copied, got %q", got)
	}
	if _, err := os.Stat(destination + ideDigestSuffix); err != nil {
		t.Fatalf("digest not recorded: %v", err)
	}

	// Source changes and the archive changes: an unmodified project is refreshed silently.
	writeIDETestFile(t, filepath.Join(source, "main.xcodeproj", "project.pbxproj"), "project v2\n")
	options.Prepare = func(staged string) error {
		writeIDETestFile(t, filepath.Join(staged, "lib", "app.a"), "archive-2")
		return nil
	}
	result, err = materialiseIDEProject(options)
	if err != nil {
		t.Fatal(err)
	}
	if result.UserModified || result.Backup != "" {
		t.Fatalf("unmodified project reported as changed: %+v", result)
	}
	if got := readIDETestFile(t, filepath.Join(destination, "main.xcodeproj", "project.pbxproj")); got != "project v2\n" {
		t.Fatalf("refresh did not replace project, got %q", got)
	}
	if got := readIDETestFile(t, filepath.Join(destination, "lib", "app.a")); got != "archive-2" {
		t.Fatalf("refresh did not replace archive, got %q", got)
	}
	assertNoIDETemporaries(t, filepath.Dir(destination))
}

func TestMaterialiseIDEProjectKeepsUserChanges(t *testing.T) {
	source := fakeIDESource(t, "v1")
	destination := filepath.Join(t.TempDir(), "xcode")
	options := ideMaterialiseOptions{Source: source, Destination: destination}
	if _, err := materialiseIDEProject(options); err != nil {
		t.Fatal(err)
	}

	writeIDETestFile(t, filepath.Join(destination, "main", "main.m"), "user edit\n")
	writeIDETestFile(t, filepath.Join(destination, "notes.txt"), "user file\n")

	result, err := materialiseIDEProject(options)
	if err != nil {
		t.Fatal(err)
	}
	if !result.UserModified {
		t.Fatal("user edits were not detected")
	}
	if got := readIDETestFile(t, filepath.Join(result.Backup, "main", "main.m")); got != "user edit\n" {
		t.Fatalf("backup lost user edit, got %q", got)
	}
	if got := readIDETestFile(t, filepath.Join(result.Backup, "notes.txt")); got != "user file\n" {
		t.Fatalf("backup lost user file, got %q", got)
	}
	if got := readIDETestFile(t, filepath.Join(destination, "main", "main.m")); got != "int main(void) { return 0; }\n" {
		t.Fatalf("generated file not restored, got %q", got)
	}
	if _, err := os.Stat(filepath.Join(destination, "notes.txt")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("user file leaked into regenerated project: %v", err)
	}

	// A second change keeps the earlier backup under a numbered name.
	writeIDETestFile(t, filepath.Join(destination, "main", "main.m"), "second edit\n")
	second, err := materialiseIDEProject(options)
	if err != nil {
		t.Fatal(err)
	}
	if second.Backup == result.Backup || second.Backup == "" {
		t.Fatalf("second backup reused %q", second.Backup)
	}
	if got := readIDETestFile(t, filepath.Join(result.Backup, "main", "main.m")); got != "user edit\n" {
		t.Fatalf("earlier backup overwritten, got %q", got)
	}
}

func TestMaterialiseIDEProjectTreatsUntrackedProjectAsUserOwned(t *testing.T) {
	source := fakeIDESource(t, "v1")
	destination := filepath.Join(t.TempDir(), "xcode")
	writeIDETestFile(t, filepath.Join(destination, "main.xcodeproj", "project.pbxproj"), "hand made\n")

	result, err := materialiseIDEProject(ideMaterialiseOptions{Source: source, Destination: destination})
	if err != nil {
		t.Fatal(err)
	}
	if !result.UserModified {
		t.Fatal("project without a recorded snapshot was not treated as user owned")
	}
	if got := readIDETestFile(t, filepath.Join(result.Backup, "main.xcodeproj", "project.pbxproj")); got != "hand made\n" {
		t.Fatalf("backup did not keep the existing project, got %q", got)
	}
}

func TestMaterialiseIDEProjectPreservesIDEUserState(t *testing.T) {
	source := fakeIDESource(t, "v1")
	destination := filepath.Join(t.TempDir(), "xcode")
	options := ideMaterialiseOptions{Source: source, Destination: destination}
	if _, err := materialiseIDEProject(options); err != nil {
		t.Fatal(err)
	}
	writeIDETestFile(t, filepath.Join(destination, "main.xcodeproj", "xcuserdata", "dev.xcuserdatad", "xcschemes", "scheme.xcscheme"), "scheme")
	writeIDETestFile(t, filepath.Join(destination, "build", "output.bin"), "build cache")

	result, err := materialiseIDEProject(options)
	if err != nil {
		t.Fatal(err)
	}
	if result.UserModified {
		t.Fatal("IDE state or build output was treated as a user edit")
	}
	if got := readIDETestFile(t, filepath.Join(destination, "main.xcodeproj", "xcuserdata", "dev.xcuserdatad", "xcschemes", "scheme.xcscheme")); got != "scheme" {
		t.Fatalf("xcuserdata not preserved, got %q", got)
	}
	if _, err := os.Stat(filepath.Join(destination, "build")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("build output should not be carried over: %v", err)
	}
}

func TestMaterialiseIDEProjectFailedPrepareLeavesProjectIntact(t *testing.T) {
	source := fakeIDESource(t, "v1")
	destination := filepath.Join(t.TempDir(), "xcode")
	if _, err := materialiseIDEProject(ideMaterialiseOptions{Source: source, Destination: destination}); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("link failed")
	_, err := materialiseIDEProject(ideMaterialiseOptions{Source: source, Destination: destination, Prepare: func(string) error { return failure }})
	if !errors.Is(err, failure) {
		t.Fatalf("expected prepare failure, got %v", err)
	}
	if got := readIDETestFile(t, filepath.Join(destination, "main.xcodeproj", "project.pbxproj")); got != "project v1\n" {
		t.Fatalf("failed refresh changed project, got %q", got)
	}
	assertNoIDETemporaries(t, filepath.Dir(destination))
}

func TestIDEExcludedPaths(t *testing.T) {
	cases := []struct {
		rel   string
		isDir bool
		want  bool
	}{
		{"main.xcodeproj/xcuserdata", true, true},
		{"app/src/.idea", true, true},
		{"local.properties", false, true},
		{"app/build", true, true},
		{"app/build.gradle", false, false},
		{"main/main.m", false, false},
		{".gradle", true, true},
	}
	for _, tc := range cases {
		if got := ideExcluded(tc.rel, tc.isDir); got != tc.want {
			t.Errorf("ideExcluded(%q, %v) = %v, want %v", tc.rel, tc.isDir, got, tc.want)
		}
	}
}

func TestDigestIDETreeTracksContentModeAndLinks(t *testing.T) {
	root := t.TempDir()
	writeIDETestFile(t, filepath.Join(root, "a.txt"), "one")
	base, err := digestIDETree(root)
	if err != nil {
		t.Fatal(err)
	}
	writeIDETestFile(t, filepath.Join(root, "a.txt"), "two")
	changed, err := digestIDETree(root)
	if err != nil {
		t.Fatal(err)
	}
	if changed == base {
		t.Fatal("content change did not change digest")
	}
	if err := os.Chmod(filepath.Join(root, "a.txt"), 0o755); err != nil {
		t.Fatal(err)
	}
	moded, err := digestIDETree(root)
	if err != nil {
		t.Fatal(err)
	}
	if moded == changed {
		t.Fatal("mode change did not change digest")
	}
	writeIDETestFile(t, filepath.Join(root, "xcuserdata", "state"), "a")
	ignored, err := digestIDETree(root)
	if err != nil {
		t.Fatal(err)
	}
	if ignored != moded {
		t.Fatal("excluded IDE state changed digest")
	}
}

func TestNextIDEBackupPathSkipsExistingBackups(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "xcode")
	writeIDETestFile(t, destination+".bak", "")
	writeIDETestFile(t, destination+".bak.1", "")
	got, err := nextIDEBackupPath(destination)
	if err != nil {
		t.Fatal(err)
	}
	if got != destination+".bak.2" {
		t.Fatalf("nextIDEBackupPath = %q", got)
	}
}

func TestRewriteIOSXcodeProjectLinksPrebuiltArchive(t *testing.T) {
	source := []byte(`		C0 /* Prebuild */ = {
			isa = PBXShellScriptBuildPhase;
			shellScript = "set -e\nAPP_ROOT=\"${PROJECT_DIR}/../../..\"\ngo build\n";
		};
		C0 /* App.a */ = {isa = PBXFileReference; lastKnownFileType = archive.ar; name = "App.a"; path = "../../../bin/App.a"; sourceTree = SOURCE_ROOT; };
		SDKROOT = iphoneos;
		SDKROOT = iphoneos;
`)
	output, err := rewriteIOSXcodeProject(source, "App.a", iosXcodeSigning{})
	if err != nil {
		t.Fatal(err)
	}
	text := string(output)
	if !strings.Contains(text, `path = "lib/App.a";`) {
		t.Fatalf("archive reference not rewritten:\n%s", text)
	}
	if strings.Contains(text, "../../../bin") || strings.Contains(text, "APP_ROOT") {
		t.Fatalf("legacy Go build still referenced:\n%s", text)
	}
	if !strings.Contains(text, "prebuilt by wails3 ios xcode") {
		t.Fatalf("build phase not replaced:\n%s", text)
	}
	if strings.Contains(text, "SDKROOT = iphoneos;") || strings.Count(text, "SDKROOT = auto;") != 2 {
		t.Fatalf("SDKROOT not set to auto:\n%s", text)
	}
}

func TestRewriteIOSXcodeProjectAcceptsGeneratedTemplate(t *testing.T) {
	out := filepath.Join(t.TempDir(), "xcode")
	if err := IOSXcodeGen(&IOSXcodeGenOptions{OutDir: out, Config: filepath.Join(t.TempDir(), "missing.yml")}); err != nil {
		t.Fatal(err)
	}
	source := readIDETestFile(t, filepath.Join(out, "main.xcodeproj", "project.pbxproj"))
	output, err := rewriteIOSXcodeProject([]byte(source), "Wails App.a", iosXcodeSigning{})
	if err != nil {
		t.Fatalf("generated template no longer rewrites: %v", err)
	}
	if !strings.Contains(string(output), `path = "lib/Wails App.a";`) {
		t.Fatal("generated template archive reference not rewritten")
	}
}

func TestRewriteIOSXcodeProjectRejectsUnexpectedTemplate(t *testing.T) {
	if _, err := rewriteIOSXcodeProject([]byte("nothing to rewrite"), "App.a", iosXcodeSigning{}); err == nil {
		t.Fatal("expected error for a project without the archive reference")
	}
}

func TestOpenXcodeProjectSelection(t *testing.T) {
	var calls []string
	run := func(name string, args ...string) error {
		calls = append(calls, strings.Join(append([]string{name}, args...), " "))
		return nil
	}
	if err := openXcodeProject(ideOpener{goos: "linux", run: run}, "/p/main.xcodeproj"); err == nil {
		t.Fatal("expected error opening Xcode off macOS")
	}
	if err := openXcodeProject(ideOpener{goos: "darwin", run: run}, "/p/main.xcodeproj"); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "open /p/main.xcodeproj" {
		t.Fatalf("calls = %v", calls)
	}
}

func TestOpenAndroidStudioSelection(t *testing.T) {
	notFound := errors.New("not found")
	missing := func(string) (string, error) { return "", notFound }
	found := func(name string) (string, error) { return "/usr/bin/" + name, nil }
	statInstalled := func(string) (fs.FileInfo, error) { return nil, nil }
	statMissing := func(string) (fs.FileInfo, error) { return nil, notFound }

	cases := []struct {
		name     string
		opener   ideOpener
		wantOpen bool
		wantCall string
	}{
		{"studio on PATH", ideOpener{goos: "darwin", lookPath: found, stat: statMissing}, true, "studio /p/android"},
		{"macOS app bundle", ideOpener{goos: "darwin", lookPath: missing, stat: statInstalled}, true, `open -a Android Studio /p/android`},
		{"macOS without Studio", ideOpener{goos: "darwin", lookPath: missing, stat: statMissing}, false, ""},
		{"linux studio.sh", ideOpener{goos: "linux", lookPath: func(name string) (string, error) {
			if name == "studio.sh" {
				return "/opt/studio.sh", nil
			}
			return "", notFound
		}, stat: statMissing}, true, "studio.sh /p/android"},
		{"windows falls back to instructions", ideOpener{goos: "windows", lookPath: missing, stat: statMissing}, false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			tc.opener.run = func(name string, args ...string) error {
				calls = append(calls, strings.Join(append([]string{name}, args...), " "))
				return nil
			}
			opened, err := openAndroidStudio(tc.opener, "/p/android")
			if err != nil {
				t.Fatal(err)
			}
			if opened != tc.wantOpen {
				t.Fatalf("opened = %v, want %v", opened, tc.wantOpen)
			}
			if tc.wantCall == "" {
				if len(calls) != 0 {
					t.Fatalf("unexpected calls %v", calls)
				}
				return
			}
			if len(calls) != 1 || calls[0] != tc.wantCall {
				t.Fatalf("calls = %v, want %q", calls, tc.wantCall)
			}
		})
	}
}

func TestIDEFindPackageMatchesTargetAndFormat(t *testing.T) {
	run := manifestPipelineRun{Plan: pipeline.Plan{
		Artifacts: []pipeline.NodeKey{"apk", "app"},
		Nodes: map[pipeline.NodeKey]pipeline.Node{
			"apk": {Spec: pipeline.PackageSpec{TargetOS: "android", Binary: "a.so"}, Artifact: pipeline.ArtifactIdentity{Format: "apk"}},
			"app": {Spec: pipeline.PackageSpec{TargetOS: "ios", Binary: "lib.a"}, Artifact: pipeline.ArtifactIdentity{Format: "app"}},
		},
	}}
	spec, err := ideFindPackage(run, "ios", "app")
	if err != nil || spec.Binary != "lib.a" {
		t.Fatalf("ideFindPackage(ios, app) = %+v, %v", spec, err)
	}
	if _, err := ideFindPackage(run, "ios", "apk"); err == nil {
		t.Fatal("expected error for missing ios apk")
	}
}

func assertNoIDETemporaries(t *testing.T, parent string) {
	t.Helper()
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, "-stage-") || strings.HasPrefix(name, ".wails-replace-") {
			t.Fatalf("temporary directory left behind: %s", name)
		}
	}
}
