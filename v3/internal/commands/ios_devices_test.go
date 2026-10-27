package commands

import (
	"bytes"
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

const simctlDevicesFixture = `{
  "devices": {
    "com.apple.CoreSimulator.SimRuntime.iOS-17-5": [
      {"dataPath": "/d/17", "isAvailable": true, "name": "iPhone 15", "state": "Shutdown", "udid": "AAAAAAAA-0000-0000-0000-000000000017"}
    ],
    "com.apple.CoreSimulator.SimRuntime.iOS-18-2": [
      {"dataPath": "/d/18a", "isAvailable": true, "name": "iPhone 16", "state": "Shutdown", "udid": "BBBBBBBB-0000-0000-0000-000000000182"},
      {"dataPath": "/d/18b", "isAvailable": true, "name": "iPad Air", "state": "Booted", "udid": "CCCCCCCC-0000-0000-0000-000000000182"}
    ],
    "com.apple.CoreSimulator.SimRuntime.iOS-9-3": [
      {"dataPath": "/d/9", "isAvailable": true, "name": "iPhone 6s", "state": "Shutdown", "udid": "DDDDDDDD-0000-0000-0000-000000000093"}
    ],
    "com.apple.CoreSimulator.SimRuntime.tvOS-18-0": [
      {"dataPath": "/d/tv", "isAvailable": true, "name": "Apple TV", "state": "Shutdown", "udid": "EEEEEEEE-0000-0000-0000-000000000180"}
    ]
  }
}`

const simctlRuntimesFixture = `{
  "runtimes": [
    {
      "identifier": "com.apple.CoreSimulator.SimRuntime.iOS-17-5",
      "isAvailable": true,
      "name": "iOS 17.5",
      "platform": "com.apple.platform.iphonesimulator",
      "supportedDeviceTypes": [
        {"identifier": "com.apple.CoreSimulator.SimDeviceType.iPhone-15", "name": "iPhone 15", "productFamily": "iPhone"}
      ],
      "version": "17.5"
    },
    {
      "identifier": "com.apple.CoreSimulator.SimRuntime.iOS-18-2",
      "isAvailable": true,
      "name": "iOS 18.2",
      "platform": "com.apple.platform.iphonesimulator",
      "supportedDeviceTypes": [
        {"identifier": "com.apple.CoreSimulator.SimDeviceType.iPhone-16", "name": "iPhone 16", "productFamily": "iPhone"},
        {"identifier": "com.apple.CoreSimulator.SimDeviceType.iPhone-16-Pro", "name": "iPhone 16 Pro", "productFamily": "iPhone"},
        {"identifier": "com.apple.CoreSimulator.SimDeviceType.iPad-Air", "name": "iPad Air", "productFamily": "iPad"}
      ],
      "version": "18.2"
    },
    {
      "identifier": "com.apple.CoreSimulator.SimRuntime.iOS-26-0",
      "isAvailable": false,
      "name": "iOS 26.0",
      "platform": "com.apple.platform.iphonesimulator",
      "supportedDeviceTypes": [
        {"identifier": "com.apple.CoreSimulator.SimDeviceType.iPhone-17", "name": "iPhone 17", "productFamily": "iPhone"}
      ],
      "version": "26.0"
    }
  ]
}`

const devicectlFixture = `{
  "info": {"version": "1"},
  "result": {
    "devices": [
      {
        "identifier": "11111111-2222-3333-4444-555555555555",
        "connectionProperties": {"pairingState": "paired", "tunnelState": "connected", "transportType": "wired"},
        "deviceProperties": {"name": "Lea's iPhone", "osVersionNumber": "18.2"},
        "hardwareProperties": {"udid": "00008110-001234560ABCDEF0", "marketingName": "iPhone 16 Pro", "productType": "iPhone17,1"}
      },
      {
        "identifier": "66666666-7777-8888-9999-000000000000",
        "connectionProperties": {"pairingState": "paired", "tunnelState": "disconnected"},
        "deviceProperties": {"name": "Test iPad", "osVersionNumber": "17.5"},
        "hardwareProperties": {"udid": "00008103-000000000000002E", "productType": "iPad13,1"}
      }
    ]
  }
}`

func TestParseSimctlDevicesSortsNewestIOSRuntimeFirst(t *testing.T) {
	simulators, err := parseSimctlDevices([]byte(simctlDevicesFixture))
	require.NoError(t, err)
	require.Len(t, simulators, 5)
	// iOS first, numerically newest runtime first (18.2 before 17.5 before 9.3), then non-iOS platforms.
	assert.Equal(t, "iPad Air", simulators[0].Name)
	assert.Equal(t, "Booted", simulators[0].State)
	assert.Equal(t, "iPhone 16", simulators[1].Name)
	assert.Equal(t, "iOS 18.2", simulators[1].Runtime)
	assert.Equal(t, "iOS 17.5", simulators[2].Runtime)
	assert.Equal(t, "iOS 9.3", simulators[3].Runtime)
	assert.Equal(t, "tvOS 18.0", simulators[4].Runtime)
}

func TestParseSimctlDevicesRejectsInvalidJSON(t *testing.T) {
	_, err := parseSimctlDevices([]byte(`{"devices":`))
	assert.ErrorContains(t, err, "parse simctl devices")
}

func TestParseSimctlRuntimesKeepsAvailabilityAndDeviceTypes(t *testing.T) {
	runtimes, err := parseSimctlRuntimes([]byte(simctlRuntimesFixture))
	require.NoError(t, err)
	require.Len(t, runtimes, 3)
	assert.True(t, runtimes[1].Available)
	assert.False(t, runtimes[2].Available)
	require.Len(t, runtimes[1].DeviceTypes, 3)
	assert.Equal(t, "com.apple.CoreSimulator.SimDeviceType.iPhone-16-Pro", runtimes[1].DeviceTypes[1].Identifier)
}

func TestParseDevicectlDevicesReadsConnectionAndHardware(t *testing.T) {
	devices, err := parseDevicectlDevices([]byte(devicectlFixture))
	require.NoError(t, err)
	require.Len(t, devices, 2)
	assert.Equal(t, iosPhysicalDevice{
		Name:       "Lea's iPhone",
		Identifier: "11111111-2222-3333-4444-555555555555",
		UDID:       "00008110-001234560ABCDEF0",
		Model:      "iPhone 16 Pro",
		OSVersion:  "18.2",
		State:      "connected",
	}, devices[0])
	// Without marketingName the product type is shown; tunnel state wins over pairing state.
	assert.Equal(t, "disconnected", devices[1].State)
	assert.Equal(t, "iPad13,1", devices[1].Model)
}

func TestCollectIOSDevicesReportsMissingDevicectlAsNote(t *testing.T) {
	report := collectIOSDevices(context.Background(), iosDeviceOperations{
		simulators: func(context.Context) ([]iosSimulator, error) {
			return []iosSimulator{{Name: "iPhone 16", UDID: "B", Runtime: "iOS 18.2", State: "Shutdown"}}, nil
		},
		devices: func(context.Context) ([]iosPhysicalDevice, error) {
			return nil, errors.New("devicectl: unknown command")
		},
	})
	require.Len(t, report.Simulators, 1)
	assert.Empty(t, report.Devices)
	require.Len(t, report.Notes, 1)
	assert.Contains(t, report.Notes[0], "physical devices unavailable")
	assert.Contains(t, report.Notes[0], "unknown command")
}

func TestPrintIOSDevicesRendersAlignedTableAndNotes(t *testing.T) {
	report := iosDevicesReport{
		Simulators: []iosSimulator{{Name: "iPhone 16", UDID: "UDID-1", Runtime: "iOS 18.2", State: "Booted"}},
		Devices:    []iosPhysicalDevice{{Name: "Lea's iPhone", Identifier: "ID-1", UDID: "U-1", Model: "iPhone 16 Pro", OSVersion: "18.2", State: "connected"}},
		Notes:      []string{"something skipped"},
	}
	var buffer bytes.Buffer
	require.NoError(t, printIOSDevices(&buffer, report))
	output := buffer.String()
	assert.Contains(t, output, "SIMULATORS")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "UDID-1")
	assert.Contains(t, output, "iPhone 16 Pro")
	assert.Contains(t, output, "note: something skipped")
}

func TestChooseIOSRunDeviceAutoPicksTheOnlyConnectedDevice(t *testing.T) {
	devices := []iosPhysicalDevice{
		{Name: "Spare", Identifier: "spare-id", State: "disconnected"},
		{Name: "Lea's iPhone", Identifier: "phone-id", UDID: "phone-udid", State: "connected"},
	}
	device, err := chooseIOSRunDevice(devices)
	require.NoError(t, err)
	assert.Equal(t, "phone-id", device)

	device, err = chooseIOSRunDevice([]iosPhysicalDevice{{Name: "No UUID", UDID: "udid-only", State: "connected"}})
	require.NoError(t, err)
	assert.Equal(t, "udid-only", device)
}

func TestChooseIOSRunDeviceRejectsZeroOrAmbiguousDevices(t *testing.T) {
	_, err := chooseIOSRunDevice([]iosPhysicalDevice{{Name: "Spare", Identifier: "spare", State: "disconnected"}})
	assert.ErrorContains(t, err, "no connected iOS device")

	_, err = chooseIOSRunDevice([]iosPhysicalDevice{
		{Name: "A", Identifier: "a", State: "connected"},
		{Name: "B", Identifier: "b", State: "connected"},
	})
	assert.ErrorContains(t, err, "2 connected iOS devices found")
	assert.ErrorContains(t, err, "A (a)")
	assert.ErrorContains(t, err, "B (b)")
}

func TestConfigureRunTargetAutoSelectsPhysicalDevice(t *testing.T) {
	original := iosPhysicalDeviceLister
	t.Cleanup(func() { iosPhysicalDeviceLister = original })
	iosPhysicalDeviceLister = func(context.Context) ([]iosPhysicalDevice, error) {
		return []iosPhysicalDevice{{Name: "Lea's iPhone", Identifier: "phone-id", State: "connected"}}, nil
	}
	if runtime.GOOS != "darwin" {
		t.Skip("iOS device run configuration requires macOS")
	}
	options := &RunOptions{Destination: "device"}
	require.NoError(t, configureRunTarget(options, "ios", "arm64", manifest.Run{}, &manifestRunOptions{Loaded: &manifest.Loaded{}}))
	assert.Equal(t, "phone-id", options.Device)

	// Plan mode keeps the explicit requirement and never queries devices.
	iosPhysicalDeviceLister = func(context.Context) ([]iosPhysicalDevice, error) {
		t.Fatal("plan mode must not query devices")
		return nil, nil
	}
	err := configureRunTarget(&RunOptions{Destination: "device", Plan: true}, "ios", "arm64", manifest.Run{}, &manifestRunOptions{})
	assert.ErrorContains(t, err, "requires --device")
}

func TestConfigureRunTargetReportsDeviceDiscoveryFailure(t *testing.T) {
	original := iosPhysicalDeviceLister
	t.Cleanup(func() { iosPhysicalDeviceLister = original })
	iosPhysicalDeviceLister = func(context.Context) ([]iosPhysicalDevice, error) {
		return nil, errors.New("devicectl missing")
	}
	if runtime.GOOS != "darwin" {
		t.Skip("iOS device run configuration requires macOS")
	}
	err := configureRunTarget(&RunOptions{Destination: "device"}, "ios", "arm64", manifest.Run{}, &manifestRunOptions{Loaded: &manifest.Loaded{}})
	assert.ErrorContains(t, err, "devicectl missing")
}

func simulatorsFixture(t *testing.T) []iosSimulator {
	t.Helper()
	simulators, err := parseSimctlDevices([]byte(simctlDevicesFixture))
	require.NoError(t, err)
	return simulators
}

func fakeSimulatorOperations(simulators []iosSimulator, runtimes []iosRuntime) (iosSimulatorOperations, *[]string) {
	var calls []string
	operations := iosSimulatorOperations{
		listSimulators: func(context.Context) ([]iosSimulator, error) { return simulators, nil },
		listRuntimes:   func(context.Context) ([]iosRuntime, error) { return runtimes, nil },
		create: func(_ context.Context, name, deviceType, runtime string) (string, error) {
			calls = append(calls, "create "+name+" "+deviceType+" "+runtime)
			return "NEW-UDID", nil
		},
		bootStatus: func(_ context.Context, udid string) error {
			calls = append(calls, "bootstatus "+udid)
			return nil
		},
		openSimulator: func(context.Context) error {
			calls = append(calls, "open")
			return errors.New("no Simulator.app")
		},
	}
	return operations, &calls
}

func TestSelectIOSSimulatorUsesTheBootedSimulator(t *testing.T) {
	operations, calls := fakeSimulatorOperations(simulatorsFixture(t), nil)
	udid, err := selectIOSDevSimulatorWithOperations(context.Background(), "", operations)
	require.NoError(t, err)
	assert.Equal(t, "CCCCCCCC-0000-0000-0000-000000000182", udid)
	assert.Equal(t, []string{"bootstatus CCCCCCCC-0000-0000-0000-000000000182"}, *calls)
}

func TestSelectIOSSimulatorRejectsMultipleBootedSimulators(t *testing.T) {
	simulators := simulatorsFixture(t)
	// Index 0 is already the booted iPad; a second booted iOS simulator makes the choice ambiguous.
	simulators[1].State = "Booted"
	operations, calls := fakeSimulatorOperations(simulators, nil)
	_, err := selectIOSDevSimulatorWithOperations(context.Background(), "", operations)
	assert.ErrorContains(t, err, "multiple iOS simulators are booted")
	assert.Empty(t, *calls)
}

func TestSelectIOSSimulatorPicksNewestAvailableIPhoneWhenNoneBooted(t *testing.T) {
	simulators := simulatorsFixture(t)
	for i := range simulators {
		simulators[i].State = "Shutdown"
	}
	operations, calls := fakeSimulatorOperations(simulators, nil)
	udid, err := selectIOSDevSimulatorWithOperations(context.Background(), "", operations)
	require.NoError(t, err)
	// iPhone 16 on iOS 18.2 is newest; the iPad and tvOS entries are never candidates.
	assert.Equal(t, "BBBBBBBB-0000-0000-0000-000000000182", udid)
	// Opening Simulator is best effort: its failure must not block the boot.
	assert.Equal(t, []string{"open", "bootstatus BBBBBBBB-0000-0000-0000-000000000182"}, *calls)
}

func TestSelectIOSSimulatorCreatesWailsIPhoneFromNewestRuntime(t *testing.T) {
	runtimes, err := parseSimctlRuntimes([]byte(simctlRuntimesFixture))
	require.NoError(t, err)
	operations, calls := fakeSimulatorOperations(nil, runtimes)
	udid, err := selectIOSDevSimulatorWithOperations(context.Background(), "", operations)
	require.NoError(t, err)
	assert.Equal(t, "NEW-UDID", udid)
	// iOS 26.0 is unavailable so the 18.2 runtime is used; the Pro model is the newest supported iPhone.
	assert.Equal(t, []string{
		"create Wails iPhone com.apple.CoreSimulator.SimDeviceType.iPhone-16-Pro com.apple.CoreSimulator.SimRuntime.iOS-18-2",
		"open",
		"bootstatus NEW-UDID",
	}, *calls)
}

func TestSelectIOSSimulatorReportsMissingRuntime(t *testing.T) {
	operations, _ := fakeSimulatorOperations(nil, []iosRuntime{{Identifier: "com.apple.CoreSimulator.SimRuntime.iOS-18-2", Available: false}})
	_, err := selectIOSDevSimulatorWithOperations(context.Background(), "", operations)
	assert.ErrorContains(t, err, "no iOS simulator runtime is installed")
}

func TestSelectIOSSimulatorHonoursExplicitDevice(t *testing.T) {
	operations, calls := fakeSimulatorOperations(nil, nil)
	udid, err := selectIOSDevSimulatorWithOperations(context.Background(), "EXPLICIT", operations)
	require.NoError(t, err)
	assert.Equal(t, "EXPLICIT", udid)
	assert.Equal(t, []string{"bootstatus EXPLICIT"}, *calls)
}

func TestIPhoneModelNumberParsesNames(t *testing.T) {
	number, ok := iPhoneModelNumber("iPhone 16 Pro")
	assert.True(t, ok)
	assert.Equal(t, 16, number)
	_, ok = iPhoneModelNumber("iPad Air")
	assert.False(t, ok)
	_, ok = iPhoneModelNumber("iPhone SE")
	assert.False(t, ok)
}

func TestCompareSimRuntimeVersionsIsNumeric(t *testing.T) {
	assert.Greater(t, compareSimRuntimeVersions("com.apple.CoreSimulator.SimRuntime.iOS-18-2", "com.apple.CoreSimulator.SimRuntime.iOS-9-3"), 0)
	assert.Equal(t, "iOS 18.2", simRuntimeDisplay("com.apple.CoreSimulator.SimRuntime.iOS-18-2"))
}
