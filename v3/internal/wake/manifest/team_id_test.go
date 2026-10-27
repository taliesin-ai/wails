package manifest

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func teamIDManifest(platform, team string) string {
	return "version: 3\nproject:\n  name: team-app\n  product_name: Team App\n  identifier: com.example.team\n  version: 1.0.0\n" +
		platform + ":\n  signing:\n    team_id: " + team + "\n"
}

func TestIsValidAppleTeamID(t *testing.T) {
	for _, valid := range []string{"ABCDE12345", "0123456789", "A0B1C2D3E4"} {
		require.True(t, IsValidAppleTeamID(valid), valid)
	}
	for _, invalid := range []string{"", "abcde12345", "ABCDE1234", "ABCDE123456", "ABCDE-1234", " ABCDE12345", "ABCDE12345\n"} {
		require.False(t, IsValidAppleTeamID(invalid), invalid)
	}
}

func TestTeamIDRoundTripsThroughEncoder(t *testing.T) {
	root := t.TempDir()
	loaded, err := decodeYAML(root, filepath.Join(root, Filename), []byte(teamIDManifest("ios", "ABCDE12345")), "")
	require.NoError(t, err)
	require.Equal(t, "ABCDE12345", loaded.Config.Signing.IOS.TeamID)

	encoded, err := EncodeConfig(loaded.Config)
	require.NoError(t, err)
	require.Contains(t, string(encoded), "team_id: ABCDE12345")

	again, err := decodeYAML(root, filepath.Join(root, EjectedFilename), encoded, "")
	require.NoError(t, err)
	require.Equal(t, "ABCDE12345", again.Config.Signing.IOS.TeamID)
}

func TestTeamIDValidationRejectsMalformedValues(t *testing.T) {
	root := t.TempDir()
	for _, team := range []string{"abcde12345", "ABCDE1234", "ABCDE123456", "ABCDE-1234"} {
		_, err := decodeYAML(root, filepath.Join(root, Filename), []byte(teamIDManifest("ios", team)), "")
		require.Error(t, err, team)
		require.Contains(t, err.Error(), "ios.signing.team_id", team)
	}
}

func TestTeamIDIsOnlyAcceptedForIOS(t *testing.T) {
	root := t.TempDir()
	_, err := decodeYAML(root, filepath.Join(root, Filename), []byte(teamIDManifest("windows", "ABCDE12345")), "")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "team_id"), err.Error())
}

func TestTeamIDJSONSchemaPattern(t *testing.T) {
	schema, err := JSONSchema()
	require.NoError(t, err)
	require.Contains(t, string(schema), `"pattern": "^[A-Z0-9]{10}$"`)
}
