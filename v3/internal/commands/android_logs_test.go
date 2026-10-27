package commands

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLineFilterWriterHoldsPartialLinesAcrossWrites(t *testing.T) {
	var out bytes.Buffer
	filter := newLineFilterWriter(&out, androidLogLineMatcher(androidLogsIdentity{packageID: "com.example.app"}))
	_, err := filter.Write([]byte("06-01 10:00:00.000 I/Wails: first\n06-01 10:00:01.000 I/Sys"))
	require.NoError(t, err)
	_, err = filter.Write([]byte("tem: noise\n06-01 10:00:02.000 I/Activity: com.example.app started\n"))
	require.NoError(t, err)
	require.NoError(t, filter.Flush())
	assert.Equal(t, "06-01 10:00:00.000 I/Wails: first\n06-01 10:00:02.000 I/Activity: com.example.app started\n", out.String())
}

func TestLineFilterWriterFlushesFinalUnterminatedLine(t *testing.T) {
	var out bytes.Buffer
	filter := newLineFilterWriter(&out, androidLogLineMatcher(androidLogsIdentity{appName: "Demo"}))
	_, err := filter.Write([]byte("E/Demo: crashed"))
	require.NoError(t, err)
	assert.Empty(t, out.String(), "a partial line is held until it completes")
	require.NoError(t, filter.Flush())
	assert.Equal(t, "E/Demo: crashed", out.String())
	require.NoError(t, filter.Flush())
	assert.Equal(t, "E/Demo: crashed", out.String(), "flushing twice must not repeat the line")
}

func TestAndroidLogLineMatcherMirrorsTaskfileFilter(t *testing.T) {
	match := androidLogLineMatcher(androidLogsIdentity{appName: "Demo", packageID: "com.example.demo"})
	assert.True(t, match("I/WailsBridge: started"))
	assert.True(t, match("W/Demo: warning"))
	assert.True(t, match("I/ActivityManager: Start proc com.example.demo"))
	assert.False(t, match("D/SurfaceFlinger: frame composed"))
	assert.False(t, match("W/wails: lower case tag does not match grep -E"))

	bare := androidLogLineMatcher(androidLogsIdentity{})
	assert.True(t, bare("I/Wails: ok"))
	assert.False(t, bare("I/Demo: no identity known"))
}

func TestAndroidLogcatArgumentsOmitPIDWhenEmpty(t *testing.T) {
	assert.Equal(t, []string{"-s", "emulator-5554", "logcat", "-v", "time"}, androidLogcatArguments("emulator-5554", "", "time"))
	assert.Equal(t, []string{"-s", "phone", "logcat", "--pid=42", "-v", "brief"}, androidLogcatArguments("phone", "42", "brief"))
}

func TestAndroidLogsFiltersTheStreamUnlessAll(t *testing.T) {
	stream := func(_ context.Context, serial string, stdout, stderr io.Writer) (<-chan error, error) {
		assert.Equal(t, "phone", serial)
		_, _ = io.WriteString(stdout, "I/Wails: kept\nD/Other: dropped\nI/Demo: kept too")
		_, _ = io.WriteString(stderr, "adb noise\n")
		done := make(chan error, 1)
		done <- nil
		return done, nil
	}
	chooser := func(_ context.Context, options AndroidRunOptions) (androidDevice, error) {
		assert.Equal(t, "phone", options.Device)
		return androidDevice{Serial: "phone", State: "device"}, nil
	}
	identity := func() (androidLogsIdentity, error) { return androidLogsIdentity{appName: "Demo"}, nil }

	var filtered, stderrBuffer bytes.Buffer
	require.NoError(t, androidLogsWithOperations(context.Background(), AndroidLogsOptions{Device: "phone"}, androidLogsOperations{
		chooseDevice: chooser, identity: identity, stream: stream, stdout: &filtered, stderr: &stderrBuffer,
	}))
	assert.Equal(t, "I/Wails: kept\nI/Demo: kept too", filtered.String())
	assert.Equal(t, "adb noise\n", stderrBuffer.String())

	var everything bytes.Buffer
	require.NoError(t, androidLogsWithOperations(context.Background(), AndroidLogsOptions{Device: "phone", All: true}, androidLogsOperations{
		chooseDevice: chooser, identity: identity, stream: stream, stdout: &everything, stderr: &stderrBuffer,
	}))
	assert.Equal(t, "I/Wails: kept\nD/Other: dropped\nI/Demo: kept too", everything.String())
}

func TestAndroidLogsReportsDeviceAndStreamFailures(t *testing.T) {
	err := androidLogsWithOperations(context.Background(), AndroidLogsOptions{}, androidLogsOperations{
		chooseDevice: func(context.Context, AndroidRunOptions) (androidDevice, error) {
			return androidDevice{}, errors.New("no Android targets")
		},
		identity: func() (androidLogsIdentity, error) { return androidLogsIdentity{}, nil },
		stream: func(context.Context, string, io.Writer, io.Writer) (<-chan error, error) {
			t.Fatal("no stream without a target")
			return nil, nil
		},
	})
	assert.ErrorContains(t, err, "no Android targets")

	done := make(chan error, 1)
	done <- errors.New("logcat exited 1")
	err = androidLogsWithOperations(context.Background(), AndroidLogsOptions{}, androidLogsOperations{
		chooseDevice: func(context.Context, AndroidRunOptions) (androidDevice, error) {
			return androidDevice{Serial: "phone"}, nil
		},
		identity: func() (androidLogsIdentity, error) { return androidLogsIdentity{}, nil },
		stream:   func(context.Context, string, io.Writer, io.Writer) (<-chan error, error) { return done, nil },
		stdout:   io.Discard,
		stderr:   io.Discard,
	})
	assert.ErrorContains(t, err, "Android logcat stopped on phone")
}

func TestAndroidLogsCancellationIsACleanExit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan error, 1)
	done <- errors.New("signal: killed")
	err := androidLogsWithOperations(ctx, AndroidLogsOptions{}, androidLogsOperations{
		chooseDevice: func(context.Context, AndroidRunOptions) (androidDevice, error) {
			return androidDevice{Serial: "phone"}, nil
		},
		identity: func() (androidLogsIdentity, error) { return androidLogsIdentity{}, nil },
		stream:   func(context.Context, string, io.Writer, io.Writer) (<-chan error, error) { return done, nil },
		stdout:   io.Discard,
		stderr:   io.Discard,
	})
	assert.NoError(t, err)
}
