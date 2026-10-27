package dev

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveVitePortPrecedence(t *testing.T) {
	tests := []struct {
		name        string
		flag        int
		environment string
		configured  int
		want        int
	}{
		{"flag beats everything", 9400, "9500", 9600, 9400},
		{"environment beats manifest", 0, "9500", 9600, 9500},
		{"manifest beats default", 0, "", 9600, 9600},
		{"default when nothing is set", 0, "", 0, DefaultVitePort},
		{"invalid environment falls through to manifest", 0, "not-a-port", 9600, 9600},
		{"out-of-range environment is ignored", 0, "70000", 9600, 9600},
		{"zero environment is ignored", 0, "0", 0, DefaultVitePort},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, ResolveVitePort(test.flag, test.environment, test.configured))
		})
	}
}
