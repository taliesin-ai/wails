package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/flags"
	"github.com/wailsapp/wails/v3/internal/templates"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
	"github.com/wailsapp/wails/v3/internal/wake/parse"
	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

// The migration helpers change the working directory and the stdout stream,
// so these tests run sequentially.
func TestStockTemplatesMigrateWithoutBlockers(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	templateNames := make([]string, 0, len(templates.GetDefaultTemplates()))
	for _, template := range templates.GetDefaultTemplates() {
		templateNames = append(templateNames, template.Name)
	}
	require.NotEmpty(t, templateNames)

	for _, templateName := range templateNames {
		t.Run(templateName, func(t *testing.T) {
			root := initLegacyStockProject(t, templateName)
			require.NoFileExists(t, filepath.Join(root, manifest.Filename))

			t.Setenv("WAILS_EXP_USE_WAKE", "1")
			withMigrationWorkingDirectory(t, root)

			var report MigrationReport
			output := captureMigrationStdout(t, func() {
				require.NoError(t, Migrate(&MigrateOptions{JSON: true, DryRun: true}))
			})
			require.NoError(t, json.Unmarshal([]byte(output), &report))
			assert.True(t, report.Complete, "stock %s project must migrate completely", templateName)
			for _, diagnostic := range report.Diagnostics {
				assert.NotContains(t, []string{"warning", "error"}, diagnostic.Severity, "%s: %+v", diagnostic.Code, diagnostic)
			}

			captureMigrationStdout(t, func() {
				require.NoError(t, Migrate(&MigrateOptions{}))
			})
			assert.FileExists(t, filepath.Join(root, manifest.MigratedFilename))
			captureMigrationStdout(t, func() {
				require.NoError(t, Migrate(&MigrateOptions{Activate: true}))
			})
			assert.FileExists(t, filepath.Join(root, manifest.Filename))

			loaded, err := manifest.Load(root, "")
			require.NoError(t, err)
			assert.Equal(t, root, loaded.Config.Root)
			expected := legacyBindingsSettingsFromTaskfile(t, root)
			assert.Equal(t, templates.IsTypescript(templateName), expected.typescript, "generated TypeScript flag must follow the template")
			assert.Equal(t, expected.typescript, loaded.Config.Frontend.Bindings.TypeScript, "typescript")
			assert.Equal(t, expected.interfaces, loaded.Config.Frontend.Bindings.Interfaces, "interfaces")
			assert.Equal(t, expected.timeType, loaded.Config.Frontend.Bindings.TimeType, "time_type")
			assert.True(t, loaded.Config.Build.Tidy, "stock go:mod:tidy must carry over as build.tidy")
			// Stock Linux packaging creates the same four formats as wails3 package.
			assert.Equal(t, []string{"appimage", "deb", "rpm", "archlinux"}, loaded.Config.Package.Linux.Formats)
			assert.Contains(t, migrationDiagnosticCodes(report), "translated-package-formats")
			assert.Contains(t, migrationDiagnosticCodes(report), "darwin-dmg-default")
			buildServer := 0
			for _, diagnostic := range report.Diagnostics {
				if diagnostic.Code == "stock-task-replacement" && strings.Contains(diagnostic.Message, "build:server") {
					buildServer++
					assert.Contains(t, diagnostic.Message, "`wails3 build --server`")
				}
			}
			assert.Equal(t, 1, buildServer, "each replacement is reported once per project")
			_, err = pipeline.PlanBuild(loaded.Config, pipeline.Request{Verb: "build", TargetOS: "linux", TargetArch: "amd64"})
			require.NoError(t, err)
		})
	}
}

func TestStockDarwinDynamicSelectionsStayReachableWithoutIssues(t *testing.T) {
	root := initLegacyStockProject(t, "vanilla")
	rootTask := filepath.Join(root, "Taskfile.yml")

	files, err := discoverTaskfiles(rootTask)
	require.NoError(t, err)
	reachability, err := analyseTaskReachability(rootTask, files)
	require.NoError(t, err)
	for _, issue := range reachability.issues {
		assert.NotEqual(t, "dynamic-task-reference", issue.Code, "stock dynamic selection reported: %+v", issue)
	}

	platform := filepath.Join(root, "build", "darwin", "Taskfile.yml")
	assert.True(t, reachability.contains(platform, "build"))
	assert.True(t, reachability.contains(platform, "build:native"), "the stock if/else selection must keep both alternatives reachable")
	assert.True(t, reachability.contains(platform, "build:docker"), "the stock if/else selection must keep both alternatives reachable")
}

func TestLegacyVitePortTranslatesToDevPort(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "Taskfile.yml"), []byte("version: '3'\nvars:\n  VITE_PORT: '9333'\ntasks: {}\n"), 0o644))

	report, doc, err := analyseMigration(root)
	require.NoError(t, err)
	assert.True(t, report.Complete)
	assert.Equal(t, 9333, doc.Dev.Port)
	assert.Contains(t, migrationDiagnosticCodes(report), "translated-dev-port")
}

func TestCustomTaskDynamicReferenceStillWarns(t *testing.T) {
	root := t.TempDir()
	custom := "version: '3'\ntasks:\n  build:\n    cmds:\n      - task: '{{if eq OS \"darwin\"}}build:native{{else}}build:docker{{end}}'\n    desc: customised\n  build:native:\n    cmds: ['echo native']\n  build:docker:\n    cmds: ['echo docker']\n"
	require.NoError(t, os.MkdirAll(filepath.Join(root, "build", "darwin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "build", "darwin", "Taskfile.yml"), []byte(custom), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "Taskfile.yml"), []byte("version: '3'\nincludes:\n  darwin: ./build/darwin/Taskfile.yml\ntasks: {}\n"), 0o644))
	rootTask := filepath.Join(root, "Taskfile.yml")

	files, err := discoverTaskfiles(rootTask)
	require.NoError(t, err)
	reachability, err := analyseTaskReachability(rootTask, files)
	require.NoError(t, err)
	require.Len(t, reachability.issues, 1)
	assert.Equal(t, "dynamic-task-reference", reachability.issues[0].Code)
	assert.Equal(t, "build", reachability.issues[0].Task)
}

func TestDynamicTaskReferencePatternExpandsSelectionsAndVariables(t *testing.T) {
	for _, test := range []struct {
		name      string
		prefix    string
		reference string
		matches   []string
		misses    []string
	}{
		{
			name:      "package manager variable",
			reference: "install:frontend:deps:{{.PACKAGE_MANAGER}}",
			matches:   []string{"install:frontend:deps:npm", "install:frontend:deps:bun"},
			misses:    []string{"install:frontend:deps", "frontend:run:npm"},
		},
		{
			name:      "if else alternatives with prefix",
			prefix:    "darwin:",
			reference: `{{if eq OS "darwin"}}build:native{{else}}build:docker{{end}}`,
			matches:   []string{"build:native", "darwin:build:native", "darwin:build:docker"},
			misses:    []string{"darwin:build:native:extra", "build:nativebuild:docker"},
		},
		{
			name:      "single conditional keeps the one target",
			reference: `{{if eq OS "darwin"}}codesign:adhoc{{end}}`,
			matches:   []string{"codesign:adhoc"},
			misses:    []string{"codesign:skip"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			pattern := dynamicTaskReferencePattern(test.prefix, test.reference)
			for _, key := range test.matches {
				assert.True(t, pattern.MatchString(key), "%q should match %q", key, pattern)
			}
			for _, key := range test.misses {
				assert.False(t, pattern.MatchString(key), "%q should not match %q", key, pattern)
			}
		})
	}
}

func TestViteOutputDirectoryIsInferredForFreshFrontends(t *testing.T) {
	for _, test := range []struct {
		name       string
		viteConfig string
		output     string
		complete   bool
		code       string
	}{
		{name: "no vite config assumes dist", output: "frontend/dist", complete: true, code: "frontend-output-default"},
		{name: "config without outDir assumes dist", viteConfig: "export default {}\n", output: "frontend/dist", complete: true, code: "frontend-output-default"},
		{name: "literal outDir", viteConfig: "export default { build: { outDir: 'build' } }\n", output: "frontend/build", complete: true},
		{name: "literal outDir with double quotes", viteConfig: `export default { build: { outDir: "../out/web" } }` + "\n", output: "out/web", complete: true},
		{name: "non literal outDir blocks", viteConfig: "export default { build: { outDir: resolveOutput() } }\n", output: "frontend/dist", complete: false, code: "frontend-output"},
		{name: "outDir escaping the project blocks", viteConfig: "export default { build: { outDir: '../../escape' } }\n", output: "frontend/dist", complete: false, code: "frontend-output"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			frontend := filepath.Join(root, "frontend")
			require.NoError(t, os.MkdirAll(frontend, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(frontend, "package.json"), []byte(`{"scripts":{"build":"vite build","dev":"vite"}}`), 0o644))
			if test.viteConfig != "" {
				require.NoError(t, os.WriteFile(filepath.Join(frontend, "vite.config.ts"), []byte(test.viteConfig), 0o644))
			}

			report := MigrationReport{Version: 1, Complete: true, Sources: map[string]string{}}
			doc := manifest.NewDocument(manifest.Project{Name: "vite", ProductName: "vite", Identifier: "com.example.vite", Version: "0.1.0"})
			require.NoError(t, applyLegacyFrontendConfiguration(root, "", &report, &doc))
			assert.Equal(t, test.complete, report.Complete)
			assert.Equal(t, test.output, doc.Frontend.OutputDirectory)
			if test.code == "" {
				assert.Empty(t, report.Diagnostics)
				return
			}
			assert.Contains(t, migrationDiagnosticCodes(report), test.code)
		})
	}
}

type legacyBindingsSettings struct {
	typescript, interfaces bool
	timeType               string
}

// legacyBindingsSettingsFromTaskfile derives the bindings settings implied by
// the generated generate:bindings command, so the expectation comes from the
// Taskfile text rather than from the template logic.
func legacyBindingsSettingsFromTaskfile(t *testing.T, root string) legacyBindingsSettings {
	t.Helper()
	tf, err := parse.Parse(filepath.Join(root, "build", "Taskfile.yml"))
	require.NoError(t, err)
	task := tf.Tasks["generate:bindings"]
	require.NotNil(t, task)
	var command string
	for _, cmd := range task.Cmds {
		if cmd != nil && strings.Contains(cmd.Cmd, "generate bindings") {
			command = cmd.Cmd
		}
	}
	require.NotEmpty(t, command)
	// Task templates such as {{if ...}} remain in the generated text; the
	// flag tokens are read up to their first template delimiter.
	fields := strings.Fields(command)
	for i := range fields {
		fields[i], _, _ = strings.Cut(fields[i], "{{")
	}
	settings := legacyBindingsSettings{timeType: "string"}
	for i, field := range fields {
		switch {
		case field == "-ts":
			settings.typescript = true
		case field == "-i":
			settings.interfaces = true
		case field == "-time-type=Date", field == "-time-type" && i+1 < len(fields) && fields[i+1] == "Date":
			settings.timeType = "Date"
		}
	}
	return settings
}

// initLegacyStockProject generates a stock project with the experiment disabled,
// so it contains the Taskfiles that shipped before the manifest cutover.
func initLegacyStockProject(t *testing.T, templateName string) string {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Chdir(t.TempDir())
	t.Setenv("WAILS_EXP_USE_WAKE", "restore")
	require.NoError(t, os.Unsetenv("WAILS_EXP_USE_WAKE"))
	options := &flags.Init{
		TemplateName: templateName, ProjectName: "Stock App", ProjectDir: filepath.Join(t.TempDir(), "project"),
		ModulePath: "example.com/stock", ProductName: "Stock App", ProductIdentifier: "com.example.stock",
		ProductVersion: "1.0.0", SkipGoModTidy: true,
	}
	require.NoError(t, Init(options))
	// Init places the project in a directory named after the project.
	require.FileExists(t, filepath.Join(options.ProjectDir, "Taskfile.yml"))
	return options.ProjectDir
}
