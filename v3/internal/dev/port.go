package dev

import "strconv"

// DefaultVitePort is the development server port used when neither the
// command line, the environment nor the manifest selects one.
const DefaultVitePort = 9245

// ResolveVitePort applies development server port precedence: an explicit
// --port flag, then a valid WAILS_VITE_PORT value, then dev.port from the
// manifest. Environment values that are not integers in 1-65535 are ignored,
// matching the legacy Taskfile behaviour.
func ResolveVitePort(flag int, environment string, configured int) int {
	if flag != 0 {
		return flag
	}
	if value, err := strconv.Atoi(environment); err == nil && value >= 1 && value <= 65535 {
		return value
	}
	if configured != 0 {
		return configured
	}
	return DefaultVitePort
}
