package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// IOSDevicesOptions controls `wails3 ios devices` output.
type IOSDevicesOptions struct {
	JSON bool `name:"json" description:"Print machine-readable JSON instead of a table"`
}

// iosSimulator is one available simulator reported by simctl.
type iosSimulator struct {
	Name      string `json:"name"`
	UDID      string `json:"udid"`
	Runtime   string `json:"runtime"`
	RuntimeID string `json:"runtime_id"`
	State     string `json:"state"`
}

// iosPhysicalDevice is one device reported by devicectl. Identifier is the
// value accepted by `devicectl --device`; UDID is the hardware identifier.
type iosPhysicalDevice struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	UDID       string `json:"udid,omitempty"`
	Model      string `json:"model,omitempty"`
	OSVersion  string `json:"os_version,omitempty"`
	State      string `json:"state"`
}

// iosRuntime is an installed simulator runtime and the device types it supports.
type iosRuntime struct {
	Identifier  string
	Available   bool
	DeviceTypes []iosDeviceType
}

type iosDeviceType struct {
	Name       string
	Identifier string
}

type iosDevicesReport struct {
	Simulators []iosSimulator      `json:"simulators"`
	Devices    []iosPhysicalDevice `json:"devices"`
	Notes      []string            `json:"notes,omitempty"`
}

type iosDeviceOperations struct {
	simulators func(context.Context) ([]iosSimulator, error)
	devices    func(context.Context) ([]iosPhysicalDevice, error)
}

// IOSDevices lists available simulators and physical devices. Failures to
// query either source are reported as notes so the other list still prints.
func IOSDevices(options *IOSDevicesOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	report := collectIOSDevices(ctx, iosDeviceOperations{simulators: listIOSSimulators, devices: listIOSPhysicalDevices})
	if options.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	return printIOSDevices(os.Stdout, report)
}

func collectIOSDevices(ctx context.Context, operations iosDeviceOperations) iosDevicesReport {
	report := iosDevicesReport{Simulators: []iosSimulator{}, Devices: []iosPhysicalDevice{}}
	if simulators, err := operations.simulators(ctx); err != nil {
		report.Notes = append(report.Notes, "simulators unavailable: "+err.Error())
	} else {
		report.Simulators = simulators
	}
	if devices, err := operations.devices(ctx); err != nil {
		report.Notes = append(report.Notes, "physical devices unavailable: "+err.Error())
	} else {
		report.Devices = devices
	}
	return report
}

func printIOSDevices(w io.Writer, report iosDevicesReport) error {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "SIMULATORS")
	if len(report.Simulators) == 0 {
		fmt.Fprintln(tw, "  none available")
	} else {
		fmt.Fprintln(tw, "NAME\tUDID\tRUNTIME\tSTATE")
		for _, simulator := range report.Simulators {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", simulator.Name, simulator.UDID, simulator.Runtime, simulator.State)
		}
	}
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "PHYSICAL DEVICES")
	if len(report.Devices) == 0 {
		fmt.Fprintln(tw, "  none found")
	} else {
		fmt.Fprintln(tw, "NAME\tIDENTIFIER\tUDID\tMODEL\tOS\tSTATE")
		for _, device := range report.Devices {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", device.Name, device.Identifier, device.UDID, device.Model, device.OSVersion, device.State)
		}
	}
	for _, note := range report.Notes {
		fmt.Fprintf(tw, "\nnote: %s\n", note)
	}
	return tw.Flush()
}

// xcrunJSON runs an xcrun command and returns its stdout, including stderr in errors.
func xcrunJSON(ctx context.Context, args ...string) ([]byte, error) {
	var stderr bytes.Buffer
	command := exec.CommandContext(ctx, "xcrun", args...)
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("xcrun %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(stderr.String()), err)
	}
	return output, nil
}

func listIOSSimulators(ctx context.Context) ([]iosSimulator, error) {
	output, err := xcrunJSON(ctx, "simctl", "list", "devices", "available", "-j")
	if err != nil {
		return nil, err
	}
	return parseSimctlDevices(output)
}

func listIOSRuntimes(ctx context.Context) ([]iosRuntime, error) {
	output, err := xcrunJSON(ctx, "simctl", "list", "runtimes", "-j")
	if err != nil {
		return nil, err
	}
	return parseSimctlRuntimes(output)
}

func createIOSSimulator(ctx context.Context, name, deviceTypeID, runtimeID string) (string, error) {
	output, err := xcrunJSON(ctx, "simctl", "create", name, deviceTypeID, runtimeID)
	if err != nil {
		return "", err
	}
	udid := strings.TrimSpace(string(output))
	if udid == "" {
		return "", fmt.Errorf("simctl create %q returned no simulator identifier", name)
	}
	return udid, nil
}

// listIOSPhysicalDevices asks devicectl for paired devices. Callers decide
// whether a failure is fatal; `ios devices` reports it as a note.
func listIOSPhysicalDevices(ctx context.Context) ([]iosPhysicalDevice, error) {
	file, err := os.CreateTemp("", "wails-devicectl-*.json")
	if err != nil {
		return nil, err
	}
	path := file.Name()
	_ = file.Close()
	defer os.Remove(path)
	var stderr bytes.Buffer
	command := exec.CommandContext(ctx, "xcrun", "devicectl", "list", "devices", "--json-output", path)
	command.Stdout = &stderr
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("xcrun devicectl list devices: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseDevicectlDevices(data)
}

func parseSimctlDevices(data []byte) ([]iosSimulator, error) {
	var list struct {
		Devices map[string][]struct {
			Name  string `json:"name"`
			UDID  string `json:"udid"`
			State string `json:"state"`
		} `json:"devices"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parse simctl devices: %w", err)
	}
	result := []iosSimulator{}
	for runtimeID, devices := range list.Devices {
		for _, device := range devices {
			result = append(result, iosSimulator{
				Name: device.Name, UDID: device.UDID, Runtime: simRuntimeDisplay(runtimeID), RuntimeID: runtimeID, State: device.State,
			})
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if iosI, iosJ := isIOSRuntimeID(result[i].RuntimeID), isIOSRuntimeID(result[j].RuntimeID); iosI != iosJ {
			return iosI
		}
		if c := compareSimRuntimeVersions(result[i].RuntimeID, result[j].RuntimeID); c != 0 {
			return c > 0
		}
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		return result[i].UDID < result[j].UDID
	})
	return result, nil
}

func parseSimctlRuntimes(data []byte) ([]iosRuntime, error) {
	var list struct {
		Runtimes []struct {
			Identifier           string `json:"identifier"`
			IsAvailable          bool   `json:"isAvailable"`
			SupportedDeviceTypes []struct {
				Name       string `json:"name"`
				Identifier string `json:"identifier"`
			} `json:"supportedDeviceTypes"`
		} `json:"runtimes"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parse simctl runtimes: %w", err)
	}
	result := make([]iosRuntime, 0, len(list.Runtimes))
	for _, runtime := range list.Runtimes {
		entry := iosRuntime{Identifier: runtime.Identifier, Available: runtime.IsAvailable}
		for _, deviceType := range runtime.SupportedDeviceTypes {
			entry.DeviceTypes = append(entry.DeviceTypes, iosDeviceType{Name: deviceType.Name, Identifier: deviceType.Identifier})
		}
		result = append(result, entry)
	}
	return result, nil
}

func parseDevicectlDevices(data []byte) ([]iosPhysicalDevice, error) {
	var list struct {
		Result struct {
			Devices []struct {
				Identifier           string `json:"identifier"`
				ConnectionProperties struct {
					PairingState string `json:"pairingState"`
					TunnelState  string `json:"tunnelState"`
				} `json:"connectionProperties"`
				DeviceProperties struct {
					Name            string `json:"name"`
					OSVersionNumber string `json:"osVersionNumber"`
				} `json:"deviceProperties"`
				HardwareProperties struct {
					UDID          string `json:"udid"`
					MarketingName string `json:"marketingName"`
					ProductType   string `json:"productType"`
				} `json:"hardwareProperties"`
			} `json:"devices"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parse devicectl devices: %w", err)
	}
	result := make([]iosPhysicalDevice, 0, len(list.Result.Devices))
	for _, device := range list.Result.Devices {
		state := device.ConnectionProperties.TunnelState
		if state == "" {
			state = device.ConnectionProperties.PairingState
		}
		model := device.HardwareProperties.MarketingName
		if model == "" {
			model = device.HardwareProperties.ProductType
		}
		result = append(result, iosPhysicalDevice{
			Name:       device.DeviceProperties.Name,
			Identifier: device.Identifier,
			UDID:       device.HardwareProperties.UDID,
			Model:      model,
			OSVersion:  device.DeviceProperties.OSVersionNumber,
			State:      state,
		})
	}
	return result, nil
}

// chooseIOSRunDevice picks the only connected physical device for a run that
// did not name one. Connected means devicectl reports the tunnel as connected.
func chooseIOSRunDevice(devices []iosPhysicalDevice) (string, error) {
	var connected []iosPhysicalDevice
	for _, device := range devices {
		if device.State == "connected" {
			connected = append(connected, device)
		}
	}
	switch len(connected) {
	case 0:
		return "", fmt.Errorf("physical iOS run requires --device; no connected iOS device was found (see: wails3 ios devices)")
	case 1:
		if connected[0].Identifier != "" {
			return connected[0].Identifier, nil
		}
		return connected[0].UDID, nil
	}
	names := make([]string, 0, len(connected))
	for _, device := range connected {
		names = append(names, fmt.Sprintf("%s (%s)", device.Name, device.Identifier))
	}
	return "", fmt.Errorf("physical iOS run requires --device; %d connected iOS devices found: %s", len(connected), strings.Join(names, ", "))
}

// parseSimRuntimeID splits "com.apple.CoreSimulator.SimRuntime.iOS-18-2" into
// the platform ("iOS") and numeric version (18, 2).
func parseSimRuntimeID(id string) (string, []int) {
	tail := id
	if index := strings.LastIndex(id, "."); index >= 0 {
		tail = id[index+1:]
	}
	parts := strings.Split(tail, "-")
	var version []int
	for _, part := range parts[1:] {
		number, err := strconv.Atoi(part)
		if err != nil {
			break
		}
		version = append(version, number)
	}
	return parts[0], version
}

func simRuntimeDisplay(id string) string {
	platform, version := parseSimRuntimeID(id)
	parts := make([]string, len(version))
	for i, number := range version {
		parts[i] = strconv.Itoa(number)
	}
	if len(parts) == 0 {
		return platform
	}
	return platform + " " + strings.Join(parts, ".")
}

// compareSimRuntimeVersions compares numerically, so iOS-18-2 is newer than iOS-9-3.
func compareSimRuntimeVersions(a, b string) int {
	_, av := parseSimRuntimeID(a)
	_, bv := parseSimRuntimeID(b)
	for i := 0; i < len(av) || i < len(bv); i++ {
		ai, bi := 0, 0
		if i < len(av) {
			ai = av[i]
		}
		if i < len(bv) {
			bi = bv[i]
		}
		if ai != bi {
			if ai < bi {
				return -1
			}
			return 1
		}
	}
	return 0
}

func isIOSRuntimeID(id string) bool {
	platform, _ := parseSimRuntimeID(id)
	return platform == "iOS"
}
