package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	wakeast "github.com/wailsapp/wails/v3/internal/wake/ast"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
	"github.com/wailsapp/wails/v3/internal/wake/parse"
)

// generatedHookScript is a new POSIX shell script that a migrated hook runs.
// It is written beside the draft and never overwrites an existing file.
type generatedHookScript struct {
	Phase   string
	Path    string
	Content string
}

// plannedHook is the translated form of customised steps for one phase.
type plannedHook struct {
	source   []string
	commands []string
}

// hookRenderError ties a refused step to the task that contains it.
type hookRenderError struct {
	node migrationTaskNode
	err  error
}

func (e *hookRenderError) Error() string { return e.err.Error() }
func (e *hookRenderError) Unwrap() error { return e.err }

// dynamicHookCommandError reports a template with no hook equivalent.
type dynamicHookCommandError struct{ template string }

func (e dynamicHookCommandError) Error() string {
	return fmt.Sprintf("template %s has no hook equivalent", e.template)
}

var (
	hookTemplatePattern   = regexp.MustCompile(`\{\{([^{}]*)\}\}`)
	safeHookLiteralRegexp = regexp.MustCompile(`^[A-Za-z0-9_./-]+$`)
)

// hookEnvironmentVariables maps task template variables onto the environment
// that hooks receive. Other values are substituted from the manifest.
var hookEnvironmentVariables = map[string]string{
	"ROOT_DIR": "${WAILS_PROJECT_DIR}",
	"OS":       "${WAILS_TARGET_OS}",
	"ARCH":     "${WAILS_TARGET_ARCH}",
}

// planHooks finds customised steps a manifest hook can express: extra
// dependencies of desktop build:native tasks, and extra commands after the
// root package dispatch. A hook runs for every platform, so a build step is
// taken only when every desktop platform carries the identical step. Taken
// steps are consumed so their own tasks are not reported as unrepresentable.
func (a *customBuildAdjustments) planHooks(reachability migrationReachability, rootTask string) {
	projectRoot := filepath.Dir(filepath.Clean(rootTask))
	static := rootStaticVariables(rootTask)
	nativeRoles := map[string]bool{}
	builds := map[string]plannedHook{}
	buildTargets := map[string][]migrationTaskNode{}
	buildNodes := map[string]migrationTaskNode{}
	var packageNode *migrationTaskNode

	for _, list := range reachability.nodes {
		for _, node := range list {
			if !reachability.contains(node.path, node.name) {
				continue
			}
			role := migrationTaskfileRole(projectRoot, node.path)
			switch {
			case node.name == "build:native" && containsString(desktopRoles, role):
				nativeRoles[role] = true
				extras := extraBuildDependencies(node.task, role)
				if len(extras) == 0 {
					continue
				}
				hook, targets, err := renderHookDependencies(reachability.nodes, projectRoot, node.prefix, extras, static)
				if err != nil {
					a.blockHook(projectRoot, err)
					continue
				}
				builds[role] = hook
				buildTargets[role] = targets
				buildNodes[role] = node
			case node.name == "package" && node.prefix == "" && filepath.Clean(node.path) == filepath.Clean(rootTask) && node.task != nil && len(node.task.Cmds) > 1:
				copied := node
				packageNode = &copied
			}
		}
	}

	if len(builds) > 0 && len(builds) == len(nativeRoles) {
		var reference plannedHook
		consistent := true
		first := true
		for _, role := range desktopRoles {
			hook, ok := builds[role]
			if !ok {
				continue
			}
			if first {
				reference, first = hook, false
			} else if !reflect.DeepEqual(hook.commands, reference.commands) {
				consistent = false
			}
		}
		if consistent {
			a.hooks["before_build"] = reference
			for _, role := range desktopRoles {
				node, ok := buildNodes[role]
				if !ok {
					continue
				}
				a.hookDependencies[nativeKey(node.path)] = dependencyNames(extraBuildDependencies(node.task, role))
				for _, target := range buildTargets[role] {
					a.consume(target.path, target.name)
				}
			}
		}
	}

	if packageNode != nil {
		commands, err := renderRootPackageHook(packageNode.task, static)
		if err != nil {
			a.blockHook(projectRoot, &hookRenderError{node: *packageNode, err: err})
		} else {
			a.hooks["after_package"] = plannedHook{source: []string{"package"}, commands: commands}
			a.consume(packageNode.path, packageNode.name)
		}
	}
}

// blockHook records why a customised step stays unrepresentable. A template
// with no hook equivalent gets its own code; the step keeps blocking through
// the ordinary task diagnostics.
func (a *customBuildAdjustments) blockHook(projectRoot string, err error) {
	var renderErr *hookRenderError
	if !errors.As(err, &renderErr) {
		return
	}
	relative, relErr := filepath.Rel(projectRoot, renderErr.node.path)
	if relErr != nil {
		relative = renderErr.node.path
	}
	code := "unsupported-hook-command"
	var dynamic dynamicHookCommandError
	if errors.As(err, &dynamic) {
		code = "dynamic-hook-command"
	}
	a.diagnostics = append(a.diagnostics, MigrationDiagnostic{Severity: "warning", Code: code, File: filepath.ToSlash(relative), Task: renderErr.node.name, Message: "custom step cannot become a hook: " + err.Error() + "; keep using Taskfiles or move the logic into a Go or script file the hook can call"})
}

// nativeKey identifies a platform build:native task within the adjustments.
func nativeKey(taskfile string) string {
	return filepath.Clean(taskfile) + "\x00build:native"
}

// extraBuildDependencies returns the dependencies of build:native that the
// stock task for role does not declare.
func extraBuildDependencies(task *wakeast.Task, role string) []*wakeast.Dep {
	stock := stockTaskNamed(role, "build:native")
	if task == nil || stock == nil {
		return nil
	}
	var extras []*wakeast.Dep
	for _, dep := range task.Deps {
		known := false
		for _, stockDep := range stock.Deps {
			if reflect.DeepEqual(dep, stockDep) {
				known = true
				break
			}
		}
		if !known {
			extras = append(extras, dep)
		}
	}
	return extras
}

func dependencyNames(deps []*wakeast.Dep) []string {
	var names []string
	for _, dep := range deps {
		if dep != nil && dep.Task != "" && !containsString(names, dep.Task) {
			names = append(names, dep.Task)
		}
	}
	return names
}

// renderHookDependencies renders the commands of each extra dependency, in order.
func renderHookDependencies(nodes map[string][]migrationTaskNode, projectRoot, prefix string, deps []*wakeast.Dep, static map[string]string) (plannedHook, []migrationTaskNode, error) {
	var result plannedHook
	var targets []migrationTaskNode
	for _, dep := range deps {
		if dep == nil || dep.Task == "" || len(dep.Vars) > 0 || dep.Silent {
			return plannedHook{}, nil, fmt.Errorf("dependency is not a plain task reference")
		}
		candidate := dep.Task
		if len(nodes[candidate]) == 0 && prefix != "" {
			candidate = prefix + dep.Task
		}
		list := nodes[candidate]
		if len(list) == 0 {
			return plannedHook{}, nil, fmt.Errorf("dependency %s is not defined", dep.Task)
		}
		target := list[0]
		commands, err := renderHookCommands(target, projectRoot, static)
		if err != nil {
			return plannedHook{}, nil, &hookRenderError{node: target, err: err}
		}
		result.source = append(result.source, dep.Task)
		result.commands = append(result.commands, commands...)
		targets = append(targets, target)
	}
	return result, targets, nil
}

// renderHookCommands turns a plain custom task into shell lines run from the
// Taskfile's own directory, as the Taskfile would have run them.
func renderHookCommands(target migrationTaskNode, projectRoot string, static map[string]string) ([]string, error) {
	task := target.task
	if len(task.Deps) > 0 || len(task.Vars) > 0 || len(task.Env) > 0 || len(task.Precondition) > 0 {
		return nil, fmt.Errorf("task %s has dependencies, variables, environment or preconditions that a hook cannot express", target.name)
	}
	relative, err := filepath.Rel(projectRoot, filepath.Dir(target.path))
	if err != nil {
		return nil, err
	}
	directory := filepath.ToSlash(relative)
	if task.Dir != "" {
		if strings.Contains(task.Dir, "{{") {
			return nil, fmt.Errorf("task %s has a templated directory", target.name)
		}
		directory = path.Join(directory, filepath.ToSlash(task.Dir))
	}
	if directory == ".." || strings.HasPrefix(directory, "../") {
		return nil, fmt.Errorf("task %s runs outside the project", target.name)
	}
	lines := []string{`cd "${WAILS_PROJECT_DIR}"`}
	if directory != "." && directory != "" {
		lines = append(lines, fmt.Sprintf(`cd "${WAILS_PROJECT_DIR}/%s"`, directory))
	}
	for _, command := range task.Cmds {
		if !plainHookCommand(command) {
			return nil, fmt.Errorf("task %s has a command that a hook cannot express", target.name)
		}
		translated, err := translateHookTemplates(command.Cmd, static)
		if err != nil {
			return nil, fmt.Errorf("task %s: %w", target.name, err)
		}
		lines = append(lines, translated)
	}
	return lines, nil
}

// renderRootPackageHook returns the commands that the root package task runs
// after its stock dispatch to the platform package task.
func renderRootPackageHook(task *wakeast.Task, static map[string]string) ([]string, error) {
	if task == nil || len(task.Cmds) < 2 || len(task.Deps) > 0 || len(task.Vars) > 0 {
		return nil, errors.New("root package task is not a dispatch with extra commands")
	}
	if first := task.Cmds[0]; first == nil || first.Task != "{{.GOOS}}:package" || first.Cmd != "" {
		return nil, errors.New("root package task does not start with the stock dispatch")
	}
	lines := []string{`cd "${WAILS_PROJECT_DIR}"`}
	for _, command := range task.Cmds[1:] {
		if !plainHookCommand(command) {
			return nil, errors.New("root package task has a command that a hook cannot express")
		}
		translated, err := translateHookTemplates(command.Cmd, static)
		if err != nil {
			return nil, err
		}
		lines = append(lines, translated)
	}
	return lines, nil
}

// rootDispatchWithExtras recognises the root package task with extra commands
// after its stock dispatch. Reachability uses it to follow the dispatch; the
// extra commands are still only accepted when they become a hook.
func rootDispatchWithExtras(name string, task *wakeast.Task) bool {
	return name == "package" && task != nil && len(task.Cmds) > 1 && task.Cmds[0] != nil && task.Cmds[0].Task == "{{.GOOS}}:package" && task.Cmds[0].Cmd == ""
}

func plainHookCommand(command *wakeast.Cmd) bool {
	return command != nil && command.Task == "" && command.For == nil && !command.IgnoreError && command.Cmd != "" && !strings.Contains(command.Cmd, "\n")
}

// translateHookTemplates replaces known template variables with what a hook
// can see. Any other template is refused, so no behaviour is guessed.
func translateHookTemplates(text string, static map[string]string) (string, error) {
	var builder strings.Builder
	last := 0
	for _, match := range hookTemplatePattern.FindAllStringSubmatchIndex(text, -1) {
		builder.WriteString(text[last:match[0]])
		inner := strings.TrimSpace(text[match[2]:match[3]])
		name := strings.TrimPrefix(inner, ".")
		if replacement, ok := hookEnvironmentVariables[name]; ok {
			builder.WriteString(replacement)
		} else if value, ok := static[name]; ok && safeHookLiteralRegexp.MatchString(value) {
			builder.WriteString(value)
		} else {
			return "", dynamicHookCommandError{template: "{{" + inner + "}}"}
		}
		last = match[1]
	}
	builder.WriteString(text[last:])
	return builder.String(), nil
}

// rootStaticVariables returns the root Taskfile variables with literal values.
func rootStaticVariables(rootTask string) map[string]string {
	result := map[string]string{}
	tf, err := parse.Parse(rootTask)
	if err != nil {
		return result
	}
	for name, value := range tf.Vars {
		if value != nil && value.Static != "" && !strings.Contains(value.Static, "{{") {
			result[name] = value.Static
		}
	}
	return result
}

// renderHookScript writes the POSIX script for one phase.
func renderHookScript(phase string, hook plannedHook) string {
	var builder strings.Builder
	builder.WriteString("#!/bin/sh\n")
	fmt.Fprintf(&builder, "# Generated by wails3 migrate from Taskfile task(s): %s.\n", strings.Join(hook.source, ", "))
	fmt.Fprintf(&builder, "# Runs the %s steps; edit this file freely after activation.\n", phase)
	builder.WriteString("set -eu\n")
	for _, line := range hook.commands {
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	return builder.String()
}

// writeGeneratedHookScripts creates each script with exclusive-create
// semantics. An existing file with identical content counts as generated; a
// dry run only reports what it would write.
func writeGeneratedHookScripts(root string, scripts []generatedHookScript, dryRun bool) (written, pending []string, err error) {
	for _, script := range scripts {
		resolved, resolveErr := manifest.ResolveProjectPath(root, "hook script", script.Path, false)
		if resolveErr != nil {
			return written, pending, resolveErr
		}
		existing, readErr := os.ReadFile(resolved)
		if readErr == nil {
			if !bytes.Equal(existing, []byte(script.Content)) {
				return written, pending, fmt.Errorf("hook script %s already exists with different content", script.Path)
			}
			continue
		}
		if !errors.Is(readErr, fs.ErrNotExist) {
			return written, pending, readErr
		}
		if dryRun {
			pending = append(pending, script.Path)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
			return written, pending, err
		}
		file, err := os.OpenFile(resolved, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
		if err != nil {
			return written, pending, fmt.Errorf("create hook script %s: %w", script.Path, err)
		}
		_, writeErr := file.WriteString(script.Content)
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			_ = os.Remove(resolved)
			return written, pending, errors.Join(writeErr, closeErr)
		}
		if err := os.Chmod(resolved, 0o755); err != nil {
			return written, pending, err
		}
		written = append(written, script.Path)
	}
	sort.Strings(written)
	return written, pending, nil
}
