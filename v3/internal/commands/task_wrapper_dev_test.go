package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/internal/flags"
	"github.com/wailsapp/wails/v3/internal/wake/pipeline"
)

func TestBuildDevRoutesDevelopmentPublicationThroughManifest(t *testing.T) {
	var got manifestRunOptions
	err := buildWithOperations(&flags.Build{Dev: true, Targets: "darwin/arm64"}, nil, manifestCommandOperations{
		active: func() (bool, error) { return true, nil },
		run:    func(options manifestRunOptions) error { got = options; return nil },
		task:   func(string, []string) error { t.Fatal("task path must not run"); return nil },
	})
	require.NoError(t, err)
	assert.True(t, got.Development)
	assert.True(t, got.DevBuild)
	assert.Equal(t, "build", got.Verb)
	assert.Equal(t, []pipeline.Target{{OS: "darwin", Arch: "arm64"}}, got.Targets)
}

func TestBuildWithoutDevLeavesProductionRequest(t *testing.T) {
	var got manifestRunOptions
	err := buildWithOperations(&flags.Build{}, nil, manifestCommandOperations{
		active: func() (bool, error) { return true, nil },
		run:    func(options manifestRunOptions) error { got = options; return nil },
		task:   func(string, []string) error { return nil },
	})
	require.NoError(t, err)
	assert.False(t, got.Development)
	assert.False(t, got.DevBuild)
}

func TestBuildDevRejectsUnsupportedCombinations(t *testing.T) {
	tests := []struct {
		name    string
		flags   flags.Build
		args    []string
		message string
	}{
		{"formats", flags.Build{Dev: true, Formats: "app"}, nil, "--formats"},
		{"profile flag", flags.Build{Dev: true, Profile: "release"}, nil, "profile"},
		{"positional profile", flags.Build{Dev: true}, []string{"release"}, "profile"},
		{"server", flags.Build{Dev: true, Server: true}, nil, "--server"},
		{"mobile target", flags.Build{Dev: true, Targets: "ios/arm64"}, nil, "desktop targets"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buildFlags := test.flags
			err := buildWithOperations(&buildFlags, test.args, manifestCommandOperations{
				active: func() (bool, error) { return true, nil },
				run:    func(manifestRunOptions) error { t.Fatal("run must not start"); return nil },
				task:   func(string, []string) error { return nil },
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.message)
		})
	}
}

func TestBuildDevAllowsObfuscation(t *testing.T) {
	var got manifestRunOptions
	buildFlags := flags.Build{Dev: true, Obfuscated: true}
	err := buildWithOperations(&buildFlags, nil, manifestCommandOperations{
		active: func() (bool, error) { return true, nil },
		run:    func(options manifestRunOptions) error { got = options; return nil },
		task:   func(string, []string) error { return nil },
	})
	require.NoError(t, err)
	assert.True(t, got.Obfuscated)
	assert.True(t, got.DevBuild)
}

func TestLegacyBuildDevPassesTaskfileVariable(t *testing.T) {
	var gotArgs []string
	buildFlags := flags.Build{Dev: true}
	err := buildWithOperations(&buildFlags, nil, manifestCommandOperations{
		active: func() (bool, error) { return false, nil },
		run:    func(manifestRunOptions) error { t.Fatal("manifest path must not run"); return nil },
		task: func(action string, args []string) error {
			assert.Equal(t, "build", action)
			gotArgs = args
			return nil
		},
	})
	require.NoError(t, err)
	assert.Contains(t, gotArgs, "DEV=true")
}
