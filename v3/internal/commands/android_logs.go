package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

// AndroidLogsOptions controls `wails3 android logs` target selection and filtering.
type AndroidLogsOptions struct {
	Device string `name:"device" description:"Connected adb device serial (default: the only connected target)"`
	All    bool   `name:"all" description:"Stream unfiltered logcat instead of only this app's lines"`
}

type androidLogsIdentity struct {
	appName   string
	packageID string
}

type androidLogsOperations struct {
	chooseDevice func(context.Context, AndroidRunOptions) (androidDevice, error)
	identity     func() (androidLogsIdentity, error)
	stream       func(ctx context.Context, serial string, stdout, stderr io.Writer) (<-chan error, error)
	stdout       io.Writer
	stderr       io.Writer
}

// AndroidLogs streams logcat for a connected device. The app filter runs in Go
// over the whole stream, so no shell pipe or PID is needed.
func AndroidLogs(options *AndroidLogsOptions) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return androidLogsWithOperations(ctx, *options, androidLogsOperations{
		chooseDevice: func(ctx context.Context, run AndroidRunOptions) (androidDevice, error) {
			return chooseAndroidDevice(ctx, run, realAndroidDeployOperations())
		},
		identity: realAndroidLogsIdentity,
		stream: func(ctx context.Context, serial string, stdout, stderr io.Writer) (<-chan error, error) {
			return startAndroidLogcatWithFormat(ctx, serial, "", "time", stdout, stderr)
		},
		stdout: os.Stdout,
		stderr: os.Stderr,
	})
}

func androidLogsWithOperations(ctx context.Context, options AndroidLogsOptions, operations androidLogsOperations) error {
	device, err := operations.chooseDevice(ctx, AndroidRunOptions{Device: options.Device})
	if err != nil {
		return err
	}
	identity, err := operations.identity()
	if err != nil {
		return err
	}
	stdout := operations.stdout
	var filter *lineFilterWriter
	if !options.All {
		filter = newLineFilterWriter(operations.stdout, androidLogLineMatcher(identity))
		stdout = filter
	}
	done, err := operations.stream(ctx, device.Serial, stdout, operations.stderr)
	if err != nil {
		return fmt.Errorf("start Android logcat on %s: %w", device.Serial, err)
	}
	logErr := <-done
	if filter != nil {
		if flushErr := filter.Flush(); flushErr != nil && logErr == nil {
			logErr = flushErr
		}
	}
	if ctx.Err() != nil {
		// Ctrl+C is the normal way to stop logcat.
		return nil
	}
	if logErr != nil {
		return fmt.Errorf("Android logcat stopped on %s: %w", device.Serial, logErr)
	}
	return nil
}

func realAndroidLogsIdentity() (androidLogsIdentity, error) {
	if !manifest.Exists(".") {
		return androidLogsIdentity{}, nil
	}
	loaded, err := manifest.Load(".", "")
	if err != nil {
		return androidLogsIdentity{}, err
	}
	return androidLogsIdentity{appName: loaded.Config.Project.Name, packageID: androidApplicationID(loaded.Config)}, nil
}

// androidLogLineMatcher mirrors the Taskfile's grep for "Wails" or the app name,
// extended with the application package ID.
func androidLogLineMatcher(identity androidLogsIdentity) func(string) bool {
	needles := []string{"Wails"}
	for _, value := range []string{identity.appName, identity.packageID} {
		if value != "" {
			needles = append(needles, value)
		}
	}
	return func(line string) bool {
		for _, needle := range needles {
			if strings.Contains(line, needle) {
				return true
			}
		}
		return false
	}
}

// lineFilterWriter forwards only complete lines accepted by keep. A partial
// trailing line is held until its newline arrives or Flush is called.
type lineFilterWriter struct {
	destination io.Writer
	keep        func(string) bool
	pending     []byte
}

func newLineFilterWriter(destination io.Writer, keep func(string) bool) *lineFilterWriter {
	return &lineFilterWriter{destination: destination, keep: keep}
}

func (w *lineFilterWriter) Write(p []byte) (int, error) {
	w.pending = append(w.pending, p...)
	for {
		index := bytes.IndexByte(w.pending, '\n')
		if index < 0 {
			break
		}
		line := w.pending[:index+1]
		w.pending = w.pending[index+1:]
		if w.keep(string(line)) {
			if _, err := w.destination.Write(line); err != nil {
				return 0, err
			}
		}
	}
	return len(p), nil
}

// Flush writes a final line that had no trailing newline.
func (w *lineFilterWriter) Flush() error {
	if len(w.pending) == 0 {
		return nil
	}
	line := w.pending
	w.pending = nil
	if w.keep(string(line)) {
		_, err := w.destination.Write(line)
		return err
	}
	return nil
}
