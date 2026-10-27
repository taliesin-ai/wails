package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeFindIdentityOutput = `  1) 1111AAAA0000000000000000000000000000AAAA "Apple Development: Other Person (ZZZZZ99999)"
  2) 2222BBBB0000000000000000000000000000BBBB "Apple Distribution: Jane Doe (ABCDE12345)"
  3) 3333CCCC0000000000000000000000000000CCCC "Apple Development: Jane Doe (ABCDE12345)"
     3 valid identities found
`

func TestSelectIOSTeamIdentityPrefersDevelopmentForTeam(t *testing.T) {
	identity, err := selectIOSTeamIdentity(fakeFindIdentityOutput, "ABCDE12345")
	require.NoError(t, err)
	assert.Equal(t, "Apple Development: Jane Doe (ABCDE12345)", identity)
}

func TestSelectIOSTeamIdentityFallsBackToDistribution(t *testing.T) {
	output := `  1) 2222BBBB0000 "Apple Distribution: Jane Doe (ABCDE12345)"`
	identity, err := selectIOSTeamIdentity(output, "ABCDE12345")
	require.NoError(t, err)
	assert.Equal(t, "Apple Distribution: Jane Doe (ABCDE12345)", identity)
}

func TestSelectIOSTeamIdentityRejectsOtherTeams(t *testing.T) {
	_, err := selectIOSTeamIdentity(fakeFindIdentityOutput, "QWERTY1234")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "QWERTY1234")
	assert.Contains(t, err.Error(), "ios.signing.identity")
}

func TestSelectIOSTeamIdentityIgnoresPartialTeamSuffix(t *testing.T) {
	// A team ID that merely ends in the same characters must not match.
	output := `  1) 1 "Apple Development: Jane (XABCDE12345)"`
	_, err := selectIOSTeamIdentity(output, "ABCDE12345")
	require.Error(t, err)
}

func TestRewriteIOSXcodeProjectAddsDevelopmentTeamToEveryConfiguration(t *testing.T) {
	// Use the generated template so the full rewrite path is exercised.
	out := filepath.Join(t.TempDir(), "xcode")
	require.NoError(t, IOSXcodeGen(&IOSXcodeGenOptions{OutDir: out, Config: filepath.Join(t.TempDir(), "missing.yml")}))
	template, err := os.ReadFile(filepath.Join(out, "main.xcodeproj", "project.pbxproj"))
	require.NoError(t, err)
	rewritten, err := rewriteIOSXcodeProject(template, "Wails App.a", "ABCDE12345")
	require.NoError(t, err)
	text := string(rewritten)
	assert.Equal(t, 2, strings.Count(text, "DEVELOPMENT_TEAM = ABCDE12345;"))
	assert.Equal(t, 2, strings.Count(text, "buildSettings = {\n\t\t\t\tDEVELOPMENT_TEAM = ABCDE12345;\n"))

	untouched, err := rewriteIOSXcodeProject(template, "Wails App.a", "")
	require.NoError(t, err)
	assert.NotContains(t, string(untouched), "DEVELOPMENT_TEAM")
}

func TestRewriteIOSXcodeProjectRejectsInvalidTeam(t *testing.T) {
	out := filepath.Join(t.TempDir(), "xcode")
	require.NoError(t, IOSXcodeGen(&IOSXcodeGenOptions{OutDir: out, Config: filepath.Join(t.TempDir(), "missing.yml")}))
	template, err := os.ReadFile(filepath.Join(out, "main.xcodeproj", "project.pbxproj"))
	require.NoError(t, err)
	_, err = rewriteIOSXcodeProject(template, "Wails App.a", "bad team")
	require.Error(t, err)
}

func TestMigrationTranslatesRootTeamID(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "Taskfile.yml"), []byte("version: '3'\nvars:\n  TEAM_ID: ABCDE12345\ntasks: {}\n"), 0o644))

	report, doc, err := analyseMigration(root)
	require.NoError(t, err)
	assert.Equal(t, "ABCDE12345", doc.Signing.IOS.TeamID)
	assert.Contains(t, report.Diagnostics, MigrationDiagnostic{Severity: "info", Code: "translated-team-id", File: "Taskfile.yml", Message: "translated TEAM_ID to ios.signing.team_id"})
}

func TestMigrationRejectsMalformedTeamID(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "Taskfile.yml"), []byte("version: '3'\nvars:\n  TEAM_ID: notvalid\ntasks: {}\n"), 0o644))

	report, doc, err := analyseMigration(root)
	require.NoError(t, err)
	assert.Empty(t, doc.Signing.IOS.TeamID)
	assert.Contains(t, report.Diagnostics, MigrationDiagnostic{Severity: "warning", Code: "invalid-team-id", File: "Taskfile.yml", Message: "TEAM_ID is not a 10-character Apple team identifier and was not migrated"})
}
