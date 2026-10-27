package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMSIXProcessorArchitectureAcceptsGoAndMSIXNames(t *testing.T) {
	for input, want := range map[string]string{
		"":      "",
		"x64":   "x64",
		"x86":   "x86",
		"arm64": "arm64",
		"amd64": "x64",
		"386":   "x86",
	} {
		got, err := msixProcessorArchitecture(input)
		assert.NoError(t, err, input)
		assert.Equal(t, want, got, input)
	}
}

func TestMSIXProcessorArchitectureRejectsUnknownNames(t *testing.T) {
	for _, input := range []string{"riscv64", "AMD64", "x86_64"} {
		_, err := msixProcessorArchitecture(input)
		assert.Error(t, err, input)
	}
}
