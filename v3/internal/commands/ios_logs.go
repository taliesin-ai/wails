package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

// IOSLogsOptions controls `wails3 ios logs` filtering and target selection.
type IOSLogsOptions struct {
	Device   string `name:"device" description:"iOS simulator UDID (default: the booted simulator)"`
	Wide     bool   `name:"wide" description:"Stream all simulator logs instead of only this app's entries"`
	BundleID string `name:"bundle-id" description:"Bundle identifier to filter on (default: from wails.yaml)"`
}

type iosLogsIdentity struct {
	processName string
	bundleID    string
}

type iosLogsOperations struct {
	identity    func(bundleFlag string) (iosLogsIdentity, error)
	listDevices func(context.Context) ([]iosPhysicalDevice, error)
	stream      func(ctx context.Context, arguments []string) error
}

// IOSLogs streams the simulator unified log filtered to this application.
func IOSLogs(options *IOSLogsOptions) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return iosLogsWithOperations(ctx, *options, iosLogsOperations{
		identity:    realIOSLogsIdentity,
		listDevices: listIOSPhysicalDevices,
		stream:      streamIOSLogs,
	})
}

func iosLogsWithOperations(ctx context.Context, options IOSLogsOptions, operations iosLogsOperations) error {
	target := options.Device
	if target == "" {
		target = "booted"
	} else if isPhysicalIOSTarget(ctx, target, operations) {
		return fmt.Errorf("unified-log streaming is not available for physical iOS device %s; open Console.app and select the device, or run the app with `xcrun devicectl device process launch --console` to see its output", target)
	}
	var predicate string
	if !options.Wide {
		identity, err := operations.identity(options.BundleID)
		if err != nil {
			return err
		}
		predicate = iosLogPredicate(identity.processName, identity.bundleID)
	}
	arguments := iosLogStreamArguments(target, predicate)
	err := operations.stream(ctx, arguments)
	if ctx.Err() != nil {
		// Ctrl+C is the normal way to stop a log stream.
		return nil
	}
	return err
}

// isPhysicalIOSTarget reports whether target names a physical device. A
// devicectl failure falls back to the classic 25-character device UDID shape.
func isPhysicalIOSTarget(ctx context.Context, target string, operations iosLogsOperations) bool {
	devices, err := operations.listDevices(ctx)
	if err == nil {
		for _, device := range devices {
			if device.Identifier == target || device.UDID == target {
				return true
			}
		}
		return false
	}
	return len(target) == 25 && target[8] == '-'
}

func realIOSLogsIdentity(bundleFlag string) (iosLogsIdentity, error) {
	var identity iosLogsIdentity
	if manifest.Exists(".") {
		loaded, err := manifest.Load(".", "")
		if err != nil {
			return iosLogsIdentity{}, err
		}
		identity.bundleID = iosApplicationID(loaded.Config)
		identity.processName = loaded.Config.Project.BinaryName
		if identity.processName == "" {
			identity.processName = loaded.Config.Project.Name
		}
	}
	if bundleFlag != "" {
		identity.bundleID = bundleFlag
	}
	if identity.bundleID == "" && identity.processName == "" {
		return iosLogsIdentity{}, errors.New("no wails.yaml found to derive the app; pass -bundle-id to filter the simulator log, or use -wide")
	}
	return identity, nil
}

func iosApplicationID(config manifest.Config) string {
	if config.Targets.IOS.Identifier != "" {
		return config.Targets.IOS.Identifier
	}
	return config.Project.Identifier
}

// iosLogPredicate matches entries sent by this app's executable or subsystem.
// Subsystem BEGINSWITH covers suffixed bundle identifiers such as ".dev".
func iosLogPredicate(processName, bundleID string) string {
	var clauses []string
	if processName != "" {
		name := escapeNSPredicateString(processName)
		clauses = append(clauses,
			fmt.Sprintf(`senderImagePath CONTAINS[c] "%s.app/"`, name),
			fmt.Sprintf(`process ==[c] "%s"`, name),
		)
	}
	if bundleID != "" {
		id := escapeNSPredicateString(bundleID)
		clauses = append(clauses,
			fmt.Sprintf(`subsystem BEGINSWITH "%s"`, id),
			fmt.Sprintf(`category CONTAINS[c] "%s"`, id),
		)
	}
	return strings.Join(clauses, " OR ")
}

func escapeNSPredicateString(value string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value)
}

func iosLogStreamArguments(target, predicate string) []string {
	arguments := []string{"simctl", "spawn", target, "log", "stream", "--level", "debug", "--style", "compact"}
	if predicate != "" {
		arguments = append(arguments, "--predicate", predicate)
	}
	return arguments
}

func streamIOSLogs(ctx context.Context, arguments []string) error {
	command := exec.CommandContext(ctx, "xcrun", arguments...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	configureManifestProcess(command)
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		return killManifestProcess(command.Process)
	}
	command.WaitDelay = 5 * time.Second
	return command.Run()
}
