package commands

import (
	"maps"
	"reflect"
	"regexp"
	"sort"
	"strings"

	wakeast "github.com/wailsapp/wails/v3/internal/wake/ast"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

// desktopRoles are the platform Taskfiles whose build:native task carries the
// Go build flags and environment that the manifest models per target.
var desktopRoles = []string{"darwin", "linux", "windows"}

// installScopeTasks are the windows tasks that default INSTALL_SCOPE.
var installScopeTasks = map[string]bool{"package": true, "create:nsis:installer": true}

var (
	installScopePattern     = regexp.MustCompile(`(default ")(machine|user)(")`)
	frontendScriptPattern   = regexp.MustCompile(`(\{\{else\}\})(build[^{}]*)(\{\{end\}\})`)
	macOSVersionPattern     = regexp.MustCompile(`^(?:-mmacosx-version-min=)?([0-9][0-9.]*)$`)
	templateVariablePattern = regexp.MustCompile(`\{\{\.?([A-Za-z_][A-Za-z0-9_]*)`)
)

// customTaskOutcome records how a customised stock task was handled. When
// translated is false, code and message describe why it still blocks.
type customTaskOutcome struct {
	translated bool
	code       string
	message    string
	// note describes a translation for the migration report.
	note string
}

// customBuildAdjustments collects translations from customised stock tasks.
// Per-platform values are written at build level only when every desktop
// platform agrees; otherwise they go to that platform's targets.
type customBuildAdjustments struct {
	tags           map[string][]string
	linkerFlags    map[string][]string
	environment    map[string]map[string]string
	minimumVersion map[string]string
	installScope   string
	frontendScript string
}

func newCustomBuildAdjustments() *customBuildAdjustments {
	return &customBuildAdjustments{
		tags:           map[string][]string{},
		linkerFlags:    map[string][]string{},
		environment:    map[string]map[string]string{},
		minimumVersion: map[string]string{},
	}
}

// normaliseRecognisedDefaults undoes the default values that translate can
// express as manifest settings, returning the task as the stock file would
// have it.
func normaliseRecognisedDefaults(task *wakeast.Task) *wakeast.Task {
	normalised, _, _ := normaliseTaskVariable(task, "INSTALL_SCOPE", installScopePattern, "machine")
	normalised, _, _ = normaliseTaskVariable(normalised, "SCRIPT", frontendScriptPattern, "build")
	return normalised
}

// stockTaskNamed returns the canonical definition of name for role, or nil.
func stockTaskNamed(role, name string) *wakeast.Task {
	for _, variant := range stockTaskVariants(role) {
		if task, ok := variant[name]; ok {
			return task
		}
	}
	return nil
}

// translate attempts to express the customised task as manifest settings. It
// never guesses: anything that is not an exact, recognised customisation blocks.
func (a *customBuildAdjustments) translate(role, name string, actual, expected *wakeast.Task) customTaskOutcome {
	if expected == nil {
		return blockedCustomTask("task is not part of the stock definition")
	}
	switch {
	case role == "common" && name == "build:frontend":
		return a.translateFrontendScript(actual, expected)
	case containsString(desktopRoles, role) && name == "build:native":
		return a.translateNativeBuild(role, actual, expected)
	case role == "windows" && installScopeTasks[name]:
		return a.translateInstallScope(actual, expected)
	}
	return blockedCustomTask(describeTaskDifference(expected, actual))
}

func blockedCustomTask(reason string) customTaskOutcome {
	return customTaskOutcome{code: "modified-task", message: reason}
}

func (a *customBuildAdjustments) translateFrontendScript(actual, expected *wakeast.Task) customTaskOutcome {
	normalised, script, found := normaliseTaskVariable(actual, "SCRIPT", frontendScriptPattern, "build")
	if !reflect.DeepEqual(normalised, expected) {
		return blockedCustomTask(describeTaskDifference(expected, normalised))
	}
	if !found || script == "build" {
		return blockedCustomTask("frontend build script is not recognised")
	}
	a.frontendScript = script
	return customTaskOutcome{translated: true, note: "frontend build script " + script}
}

func (a *customBuildAdjustments) translateInstallScope(actual, expected *wakeast.Task) customTaskOutcome {
	normalised, scope, found := normaliseTaskVariable(actual, "INSTALL_SCOPE", installScopePattern, "machine")
	if !reflect.DeepEqual(normalised, expected) {
		return blockedCustomTask(describeTaskDifference(expected, normalised))
	}
	if !found || scope == "machine" {
		return blockedCustomTask("INSTALL_SCOPE default is not recognised")
	}
	if a.installScope != "" && a.installScope != scope {
		return blockedCustomTask("INSTALL_SCOPE defaults conflict across windows tasks")
	}
	a.installScope = scope
	return customTaskOutcome{translated: true, note: "NSIS install scope " + scope}
}

func (a *customBuildAdjustments) translateNativeBuild(role string, actual, expected *wakeast.Task) customTaskOutcome {
	changedEnv := map[string]string{}
	for key, value := range actual.Env {
		if expectedValue, ok := expected.Env[key]; !ok || expectedValue != value {
			changedEnv[key] = value
		}
	}
	for key := range expected.Env {
		if _, ok := actual.Env[key]; !ok {
			return blockedCustomTask("environment variable " + key + " was removed")
		}
	}

	// Undo the recognised customisations, then require the rest to match the
	// stock task exactly. Anything left over is an unrepresentable change.
	check := cloneCustomTask(actual)
	check.Env = maps.Clone(expected.Env)
	flagsChanged := !reflect.DeepEqual(expected.Vars["BUILD_FLAGS"], actual.Vars["BUILD_FLAGS"])
	if flagsChanged {
		check.Vars["BUILD_FLAGS"] = expected.Vars["BUILD_FLAGS"]
	}
	if !reflect.DeepEqual(check, expected) {
		return blockedCustomTask(describeTaskDifference(expected, check))
	}

	if flagsChanged {
		if outcome := a.translateBuildFlags(role, varText(expected.Vars["BUILD_FLAGS"]), varText(actual.Vars["BUILD_FLAGS"])); !outcome.translated {
			return outcome
		}
	}
	if len(changedEnv) > 0 {
		if version, ok := macOSMinimumVersion(role, changedEnv); ok {
			a.minimumVersion[role] = version
			changedEnv = map[string]string{}
		}
	}
	if len(changedEnv) > 0 {
		for key, value := range changedEnv {
			if strings.Contains(value, "{{") {
				return customTaskOutcome{code: "dynamic-build-flag", message: "environment variable " + key + " is computed at build time; set it to a literal value or write it from a before_build hook"}
			}
		}
		a.environment[role] = changedEnv
	}
	return customTaskOutcome{translated: true, note: role + " build:native customisation"}
}

// translateBuildFlags turns the single inserted fragment of a BUILD_FLAGS
// change into tags (",tag") or linker flags ("-X ..."). Template references
// cannot be known at migration time, so they block with dynamic-build-flag.
func (a *customBuildAdjustments) translateBuildFlags(role, expected, actual string) customTaskOutcome {
	prefix := 0
	for prefix < len(expected) && prefix < len(actual) && expected[prefix] == actual[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(expected)-prefix && suffix < len(actual)-prefix && expected[len(expected)-1-suffix] == actual[len(actual)-1-suffix] {
		suffix++
	}
	removed := expected[prefix : len(expected)-suffix]
	added := actual[prefix : len(actual)-suffix]
	if removed != "" {
		return blockedCustomTask("build flags " + strings.TrimSpace(removed) + " were removed")
	}
	// A pure insertion can slide left over equal characters, so it is shifted
	// until it starts at a token boundary ("-H" against "-X" aligns on "-").
	for prefix > 0 && added != "" && expected[prefix-1] == added[len(added)-1] {
		added = added[len(added)-1:] + added[:len(added)-1]
		prefix--
	}
	if strings.Contains(added, "{{") {
		names := templateVariableNames(added)
		return customTaskOutcome{code: "dynamic-build-flag", message: "build flags reference " + strings.Join(names, ", ") + " at build time; set build.go.tags or build.go.linker_flags to literal values, or write the value from a before_build hook"}
	}
	trimmed := strings.TrimSpace(added)
	switch {
	case strings.HasPrefix(added, ",") && !strings.ContainsAny(trimmed, " \t"):
		a.tags[role] = append(a.tags[role], strings.Split(strings.TrimPrefix(trimmed, ","), ",")...)
		return customTaskOutcome{translated: true, note: role + " build tags"}
	case trimmed != "" && strings.HasPrefix(trimmed, "-") && (strings.HasPrefix(added, " ") || strings.HasPrefix(added, "-")):
		a.linkerFlags[role] = append(a.linkerFlags[role], groupLinkerFlags(strings.Fields(trimmed))...)
		return customTaskOutcome{translated: true, note: role + " linker flags"}
	}
	return blockedCustomTask("build flag change " + strings.TrimSpace(added) + " is not a recognised tag or linker flag")
}

// groupLinkerFlags joins each flag with its value: "-X" "a=b" -> "-X a=b".
func groupLinkerFlags(fields []string) []string {
	var flags []string
	for _, field := range fields {
		if strings.HasPrefix(field, "-") || len(flags) == 0 {
			flags = append(flags, field)
			continue
		}
		flags[len(flags)-1] += " " + field
	}
	return flags
}

func templateVariableNames(text string) []string {
	seen := map[string]bool{}
	var names []string
	for _, match := range templateVariablePattern.FindAllStringSubmatch(text, -1) {
		if !seen[match[1]] {
			seen[match[1]] = true
			names = append(names, match[1])
		}
	}
	sort.Strings(names)
	return names
}

// macOSMinimumVersion recognises env changes that only restate the macOS
// deployment target, which the manifest models as minimum_version.
func macOSMinimumVersion(role string, changed map[string]string) (string, bool) {
	if role != "darwin" {
		return "", false
	}
	version := ""
	for key, value := range changed {
		switch key {
		case "CGO_CFLAGS", "CGO_LDFLAGS", "MACOSX_DEPLOYMENT_TARGET":
		default:
			return "", false
		}
		match := macOSVersionPattern.FindStringSubmatch(value)
		if match == nil || (version != "" && version != match[1]) {
			return "", false
		}
		version = match[1]
	}
	return version, version != ""
}

// apply writes the collected translations into the migrated document.
func (a *customBuildAdjustments) apply(doc *manifest.Document, report *MigrationReport) {
	applyPerDesktopList(&doc.Build.Go.Tags, &doc.Targets, a.tags, func(target *manifest.Target) *[]string { return &target.Tags })
	applyPerDesktopList(&doc.Build.Go.LinkerFlags, &doc.Targets, a.linkerFlags, func(target *manifest.Target) *[]string { return &target.LinkerFlags })
	applyPerDesktopEnvironment(doc, a.environment)
	for role, version := range a.minimumVersion {
		platformByOS(&doc.Targets, role).MinimumVersion = version
	}
	if a.installScope != "" {
		doc.Package.Windows.NSIS.InstallScope = a.installScope
	}
	if a.frontendScript != "" && len(doc.Frontend.Build) == 3 {
		doc.Frontend.Build = []string{doc.Frontend.Build[0], "run", a.frontendScript}
	}
}

// applyPerDesktopList places values at build level when all desktop platforms
// carry the same list, and per target otherwise.
func applyPerDesktopList(build *[]string, targets *manifest.Targets, values map[string][]string, field func(*manifest.Target) *[]string) {
	if len(values) == 0 {
		return
	}
	if sameAcrossDesktop(values) {
		*build = append([]string(nil), values["darwin"]...)
		return
	}
	for role, list := range values {
		platform := platformByOS(targets, role)
		*field(&platform.AMD64) = append([]string(nil), list...)
		*field(&platform.ARM64) = append([]string(nil), list...)
	}
}

func applyPerDesktopEnvironment(doc *manifest.Document, values map[string]map[string]string) {
	if len(values) == 0 {
		return
	}
	shared := len(values) == len(desktopRoles)
	for _, role := range desktopRoles {
		if !reflect.DeepEqual(values[role], values[desktopRoles[0]]) {
			shared = false
		}
	}
	if shared {
		doc.Build.Environment = maps.Clone(values[desktopRoles[0]])
		return
	}
	for role, environment := range values {
		platform := platformByOS(&doc.Targets, role)
		platform.AMD64.Environment = maps.Clone(environment)
		platform.ARM64.Environment = maps.Clone(environment)
	}
}

func sameAcrossDesktop(values map[string][]string) bool {
	if len(values) != len(desktopRoles) {
		return false
	}
	for _, role := range desktopRoles {
		if !reflect.DeepEqual(values[role], values[desktopRoles[0]]) {
			return false
		}
	}
	return true
}

func platformByOS(targets *manifest.Targets, os string) *manifest.Platform {
	switch os {
	case "darwin":
		return &targets.Darwin
	case "linux":
		return &targets.Linux
	case "ios":
		return &targets.IOS
	case "android":
		return &targets.Android
	default:
		return &targets.Windows
	}
}

// normaliseTaskVariable undoes one customised value of key in task (task vars
// and command vars) by rewriting the custom text to canonical. The pattern
// must have three groups: the text before the value, the value, and the text
// after it. It returns the normalised copy and the custom value found.
func normaliseTaskVariable(task *wakeast.Task, key string, pattern *regexp.Regexp, canonical string) (*wakeast.Task, string, bool) {
	result := cloneCustomTask(task)
	captured := ""
	found := false
	rewrite := func(variable *wakeast.Var) {
		if variable == nil {
			return
		}
		for _, field := range []*string{&variable.Static, &variable.Value, &variable.Ref, &variable.Shell} {
			if match := pattern.FindStringSubmatch(*field); match != nil {
				captured, found = match[2], true
				*field = pattern.ReplaceAllString(*field, "${1}"+canonical+"${3}")
			}
		}
	}
	rewrite(result.Vars[key])
	for _, command := range result.Cmds {
		if command != nil {
			rewrite(command.Vars[key])
		}
	}
	return result, captured, found
}

func cloneCustomTask(task *wakeast.Task) *wakeast.Task {
	clone := *task
	clone.Vars = cloneVarMap(task.Vars)
	clone.Env = maps.Clone(task.Env)
	if task.Cmds != nil {
		clone.Cmds = make([]*wakeast.Cmd, len(task.Cmds))
		for index, command := range task.Cmds {
			if command == nil {
				continue
			}
			copied := *command
			copied.Vars = cloneVarMap(command.Vars)
			clone.Cmds[index] = &copied
		}
	}
	return &clone
}

func cloneVarMap(vars map[string]*wakeast.Var) map[string]*wakeast.Var {
	if vars == nil {
		return nil
	}
	result := make(map[string]*wakeast.Var, len(vars))
	for key, value := range vars {
		if value == nil {
			result[key] = nil
			continue
		}
		copied := *value
		result[key] = &copied
	}
	return result
}

func varText(variable *wakeast.Var) string {
	if variable == nil {
		return ""
	}
	for _, text := range []string{variable.Static, variable.Value, variable.Ref} {
		if text != "" {
			return text
		}
	}
	return ""
}

// describeTaskDifference names the first structural difference between the
// stock task and the customised one, for the modified-task diagnostic.
func describeTaskDifference(expected, actual *wakeast.Task) string {
	if actual == nil {
		return "task is missing from the customised Taskfile"
	}
	if !reflect.DeepEqual(expected.Deps, actual.Deps) {
		return "dependencies changed"
	}
	if len(expected.Cmds) != len(actual.Cmds) {
		return "commands changed"
	}
	for index := range expected.Cmds {
		if !reflect.DeepEqual(expected.Cmds[index], actual.Cmds[index]) {
			return "commands changed"
		}
	}
	for _, key := range unionKeys(expected.Vars, actual.Vars) {
		if !reflect.DeepEqual(expected.Vars[key], actual.Vars[key]) {
			return "variable " + key + " changed"
		}
	}
	for _, key := range unionKeys(expected.Env, actual.Env) {
		if expected.Env[key] != actual.Env[key] {
			return "environment variable " + key + " changed"
		}
	}
	return "task fields changed"
}

func unionKeys[V any](left, right map[string]V) []string {
	seen := map[string]bool{}
	var keys []string
	for key := range left {
		if !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
	}
	for key := range right {
		if !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// customTaskNote formats a translation note for the report.
func customTaskNote(outcome customTaskOutcome) string {
	return "translated customised task: " + strings.TrimSpace(outcome.note)
}
