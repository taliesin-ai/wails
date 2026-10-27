package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAndroidSDKDirPrecedence(t *testing.T) {
	existing := map[string]bool{"/env/home": true, "/env/root": true, "/home/u/Library/Android/sdk": true}
	exists := func(path string) bool { return existing[path] }
	env := func(values map[string]string) func(string) string {
		return func(key string) string { return values[key] }
	}

	if got := androidSDKDir(env(map[string]string{"ANDROID_HOME": "/env/home", "ANDROID_SDK_ROOT": "/env/root"}), "darwin", "/home/u", exists); got != "/env/home" {
		t.Fatalf("ANDROID_HOME should win, got %q", got)
	}
	if got := androidSDKDir(env(map[string]string{"ANDROID_HOME": "/missing", "ANDROID_SDK_ROOT": "/env/root"}), "darwin", "/home/u", exists); got != "/env/root" {
		t.Fatalf("ANDROID_SDK_ROOT should be used when ANDROID_HOME is unusable, got %q", got)
	}
	if got := androidSDKDir(env(nil), "darwin", "/home/u", exists); got != "/home/u/Library/Android/sdk" {
		t.Fatalf("darwin default, got %q", got)
	}
	if got := androidSDKDir(env(nil), "linux", "/home/u", func(string) bool { return false }); got != "" {
		t.Fatalf("missing SDK should resolve to empty, got %q", got)
	}
}

func TestAndroidSDKDirWindowsDefault(t *testing.T) {
	local := `C:\Users\u\AppData\Local`
	want := filepath.Join(local, "Android", "Sdk")
	env := func(key string) string {
		if key == "LOCALAPPDATA" {
			return local
		}
		return ""
	}
	exists := func(path string) bool { return path == want }
	if got := androidSDKDir(env, "windows", "", exists); got != want {
		t.Fatalf("windows default = %q, want %q", got, want)
	}
}

func TestEscapeJavaPropertyWindowsPath(t *testing.T) {
	got := androidLocalPropertiesContent(`C:\Users\dev\Android\Sdk`)
	want := `sdk.dir=C\:\\Users\\dev\\Android\\Sdk` + "\n"
	if got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
	if escapeJavaProperty("/Users/dev/sdk") != "/Users/dev/sdk" {
		t.Fatal("POSIX path should not gain escapes")
	}
}

func TestWriteAndroidLocalPropertiesOnlyWhenAbsent(t *testing.T) {
	sdk := t.TempDir()
	t.Setenv("ANDROID_HOME", sdk)
	t.Setenv("ANDROID_SDK_ROOT", "")

	staged := t.TempDir()
	if err := writeAndroidLocalProperties(staged); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(staged, "local.properties"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != androidLocalPropertiesContent(sdk) {
		t.Fatalf("local.properties = %q", data)
	}

	userFile := filepath.Join(staged, "local.properties")
	if err := os.WriteFile(userFile, []byte("sdk.dir=/user/choice\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeAndroidLocalProperties(staged); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(userFile); string(got) != "sdk.dir=/user/choice\n" {
		t.Fatalf("user local.properties was replaced: %q", got)
	}
}

func TestWriteAndroidLocalPropertiesSkipsMissingSDK(t *testing.T) {
	t.Setenv("ANDROID_HOME", filepath.Join(t.TempDir(), "absent"))
	t.Setenv("ANDROID_SDK_ROOT", "")
	t.Setenv("HOME", t.TempDir())
	staged := t.TempDir()
	if err := writeAndroidLocalProperties(staged); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(staged, "local.properties")); !os.IsNotExist(err) {
		t.Fatalf("local.properties written without an SDK: %v", err)
	}
}
