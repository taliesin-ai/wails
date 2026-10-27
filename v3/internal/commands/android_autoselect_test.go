package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAndroidNonInteractiveStartsTheOnlyAVD(t *testing.T) {
	operations := fakeAndroidDeployOperations()
	operations.listDevices = func(context.Context) ([]androidDevice, error) { return nil, nil }
	operations.listAVDs = func(context.Context) ([]string, error) { return []string{"Pixel_8"}, nil }
	var started []string
	operations.startAVD = func(_ context.Context, name string) (androidDevice, error) {
		started = append(started, name)
		return androidDevice{Serial: "emulator-5554", State: "device", Emulator: true}, nil
	}
	device, err := chooseAndroidDevice(context.Background(), AndroidRunOptions{}, operations)
	require.NoError(t, err)
	assert.Equal(t, "emulator-5554", device.Serial)
	assert.Equal(t, []string{"Pixel_8"}, started)
}

func TestAndroidNonInteractiveStartsFirstSortedAVDWhenSeveralExist(t *testing.T) {
	operations := fakeAndroidDeployOperations()
	operations.listDevices = func(context.Context) ([]androidDevice, error) { return nil, nil }
	// Deliberately unsorted: the choice must not depend on the lister's order.
	operations.listAVDs = func(context.Context) ([]string, error) { return []string{"Tablet", "Pixel_8", "Wear"}, nil }
	var started []string
	operations.startAVD = func(_ context.Context, name string) (androidDevice, error) {
		started = append(started, name)
		return androidDevice{Serial: "emulator-5554", State: "device", Emulator: true}, nil
	}
	_, err := chooseAndroidDevice(context.Background(), AndroidRunOptions{}, operations)
	require.NoError(t, err)
	assert.Equal(t, []string{"Pixel_8"}, started)
}

func TestAndroidNonInteractiveNeverCreatesAnAVD(t *testing.T) {
	operations := fakeAndroidDeployOperations()
	operations.listDevices = func(context.Context) ([]androidDevice, error) { return nil, nil }
	operations.listAVDs = func(context.Context) ([]string, error) { return nil, nil }
	operations.startAVD = func(context.Context, string) (androidDevice, error) {
		t.Fatal("no AVD exists, so none may be started")
		return androidDevice{}, nil
	}
	_, err := chooseAndroidDevice(context.Background(), AndroidRunOptions{}, operations)
	require.Error(t, err)
	assert.ErrorContains(t, err, `avdmanager create avd -n Wails_API_35 -k "system-images;android-35;google_apis;`)
	assert.ErrorContains(t, err, "sdkmanager")
}

func TestAndroidNonInteractiveReportsAVDListingFailure(t *testing.T) {
	operations := fakeAndroidDeployOperations()
	operations.listDevices = func(context.Context) ([]androidDevice, error) { return nil, nil }
	operations.listAVDs = func(context.Context) ([]string, error) { return nil, errors.New("emulator not found") }
	_, err := chooseAndroidDevice(context.Background(), AndroidRunOptions{}, operations)
	assert.ErrorContains(t, err, "emulator not found")
}

func TestAndroidNonInteractiveKeepsErrorForMultipleConnectedDevices(t *testing.T) {
	operations := fakeAndroidDeployOperations()
	operations.listDevices = func(context.Context) ([]androidDevice, error) {
		return []androidDevice{{Serial: "a", State: "device"}, {Serial: "b", State: "device"}}, nil
	}
	operations.startAVD = func(context.Context, string) (androidDevice, error) {
		t.Fatal("connected devices must not trigger an emulator start")
		return androidDevice{}, nil
	}
	_, err := chooseAndroidDevice(context.Background(), AndroidRunOptions{}, operations)
	assert.ErrorContains(t, err, "found 2 available Android targets")
}
