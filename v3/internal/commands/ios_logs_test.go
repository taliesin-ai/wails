package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIOSLogPredicateMatchesProcessAndBundleAndEscapesQuotes(t *testing.T) {
	predicate := iosLogPredicate("myapp", "com.example.myapp")
	assert.Equal(t, `senderImagePath CONTAINS[c] "myapp.app/" OR process ==[c] "myapp" OR subsystem BEGINSWITH "com.example.myapp" OR category CONTAINS[c] "com.example.myapp"`, predicate)

	assert.Equal(t, `senderImagePath CONTAINS[c] "odd\"name.app/" OR process ==[c] "odd\"name"`, iosLogPredicate(`odd"name`, ""))
	assert.Equal(t, "", iosLogPredicate("", ""))
}

func TestIOSLogStreamArgumentsOmitPredicateWhenWide(t *testing.T) {
	assert.Equal(t, []string{"simctl", "spawn", "booted", "log", "stream", "--level", "debug", "--style", "compact"},
		iosLogStreamArguments("booted", ""))
	assert.Equal(t, []string{"simctl", "spawn", "UDID-1", "log", "stream", "--level", "debug", "--style", "compact", "--predicate", "process ==[c] \"app\""},
		iosLogStreamArguments("UDID-1", `process ==[c] "app"`))
}

func TestIOSLogsStreamsTheBootedSimulatorWithAppPredicate(t *testing.T) {
	var streamed []string
	err := iosLogsWithOperations(context.Background(), IOSLogsOptions{}, iosLogsOperations{
		identity: func(string) (iosLogsIdentity, error) {
			return iosLogsIdentity{processName: "myapp", bundleID: "com.example.myapp"}, nil
		},
		listDevices: func(context.Context) ([]iosPhysicalDevice, error) { return nil, nil },
		stream: func(_ context.Context, arguments []string) error {
			streamed = arguments
			return nil
		},
	})
	require.NoError(t, err)
	require.Equal(t, "booted", streamed[2])
	assert.Equal(t, "--predicate", streamed[9])
	assert.Contains(t, streamed[10], `process ==[c] "myapp"`)
}

func TestIOSLogsWideDropsThePredicateWithoutReadingTheManifest(t *testing.T) {
	var streamed []string
	err := iosLogsWithOperations(context.Background(), IOSLogsOptions{Wide: true, Device: "AAAAAAAA-0000-0000-0000-000000000182"}, iosLogsOperations{
		identity: func(string) (iosLogsIdentity, error) {
			t.Fatal("wide logs must not derive the app identity")
			return iosLogsIdentity{}, nil
		},
		listDevices: func(context.Context) ([]iosPhysicalDevice, error) { return nil, nil },
		stream: func(_ context.Context, arguments []string) error {
			streamed = arguments
			return nil
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "AAAAAAAA-0000-0000-0000-000000000182", streamed[2])
	assert.NotContains(t, streamed, "--predicate")
}

func TestIOSLogsRejectsPhysicalDeviceWithClearMessage(t *testing.T) {
	err := iosLogsWithOperations(context.Background(), IOSLogsOptions{Device: "phone-id", Wide: true}, iosLogsOperations{
		identity: func(string) (iosLogsIdentity, error) { return iosLogsIdentity{}, nil },
		listDevices: func(context.Context) ([]iosPhysicalDevice, error) {
			return []iosPhysicalDevice{{Name: "Lea's iPhone", Identifier: "phone-id", UDID: "udid", PairingState: "paired", TunnelState: "disconnected"}}, nil
		},
		stream: func(context.Context, []string) error {
			t.Fatal("physical devices must not be streamed through simctl")
			return nil
		},
	})
	require.ErrorContains(t, err, "not available for physical iOS device phone-id")
	assert.ErrorContains(t, err, "Console.app")
}

func TestIOSLogsUsesUDIDShapeWhenDevicectlIsUnavailable(t *testing.T) {
	err := iosLogsWithOperations(context.Background(), IOSLogsOptions{Device: "00008110-001234560ABCDEF0", Wide: true}, iosLogsOperations{
		identity:    func(string) (iosLogsIdentity, error) { return iosLogsIdentity{}, nil },
		listDevices: func(context.Context) ([]iosPhysicalDevice, error) { return nil, errors.New("no devicectl") },
		stream: func(context.Context, []string) error {
			t.Fatal("device-shaped UDID must not be streamed through simctl")
			return nil
		},
	})
	assert.ErrorContains(t, err, "physical iOS device")
}

func TestIOSLogsRequiresAppIdentityUnlessWide(t *testing.T) {
	err := iosLogsWithOperations(context.Background(), IOSLogsOptions{}, iosLogsOperations{
		identity:    func(string) (iosLogsIdentity, error) { return iosLogsIdentity{}, errors.New("no wails.yaml found") },
		listDevices: func(context.Context) ([]iosPhysicalDevice, error) { return nil, nil },
		stream: func(context.Context, []string) error {
			t.Fatal("logs must not stream without an app identity")
			return nil
		},
	})
	assert.ErrorContains(t, err, "no wails.yaml found")
}

func TestIOSLogsCancellationIsACleanExit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := iosLogsWithOperations(ctx, IOSLogsOptions{Wide: true}, iosLogsOperations{
		listDevices: func(context.Context) ([]iosPhysicalDevice, error) { return nil, nil },
		stream:      func(context.Context, []string) error { return errors.New("signal: killed") },
	})
	assert.NoError(t, err)
}

func TestIOSLogsTreatsAnyListedDeviceUDIDAsPhysical(t *testing.T) {
	// A paired but disconnected device is still physical, so its UDID must not reach simctl.
	err := iosLogsWithOperations(context.Background(), IOSLogsOptions{Device: "00008103-000000000000002E", Wide: true}, iosLogsOperations{
		identity: func(string) (iosLogsIdentity, error) { return iosLogsIdentity{}, nil },
		listDevices: func(context.Context) ([]iosPhysicalDevice, error) {
			return []iosPhysicalDevice{{Name: "Test iPad", Identifier: "66666666", UDID: "00008103-000000000000002E", PairingState: "paired", TunnelState: "disconnected"}}, nil
		},
		stream: func(context.Context, []string) error {
			t.Fatal("a listed physical device UDID must not be streamed through simctl")
			return nil
		},
	})
	assert.ErrorContains(t, err, "physical iOS device 00008103-000000000000002E")
}
