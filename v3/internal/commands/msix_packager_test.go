package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

func TestMSIXPackagerOptionsMapsManifestTool(t *testing.T) {
	for _, test := range []struct {
		tool       string
		wantTool   bool
		wantMakeOK bool
	}{
		{tool: "", wantTool: false, wantMakeOK: true},
		{tool: manifest.MSIXToolMakeAppx, wantTool: false, wantMakeOK: true},
		{tool: manifest.MSIXToolPackagingTool, wantTool: true, wantMakeOK: false},
	} {
		options := msixPackagerOptions(test.tool)
		assert.Equal(t, test.wantTool, options.UseMsixPackagingTool, test.tool)
		assert.Equal(t, test.wantMakeOK, options.UseMakeAppx, test.tool)
	}
}
