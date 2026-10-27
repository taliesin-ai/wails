package commands

import (
	"context"
	"fmt"
	"strings"
)

// selectIOSTeamIdentity picks the codesigning identity for an Apple team from
// the output of `security find-identity -v -p codesigning`. Lines look like
//
//  1. 0123...ABCD "Apple Development: Jane Doe (ABCDE12345)"
//
// Development identities are preferred because device builds for development
// are signed with them; any other identity for the team is accepted otherwise.
func selectIOSTeamIdentity(findOutput, team string) (string, error) {
	suffix := "(" + team + ")"
	var fallback string
	for _, line := range strings.Split(findOutput, "\n") {
		first := strings.Index(line, `"`)
		last := strings.LastIndex(line, `"`)
		if first < 0 || last <= first {
			continue
		}
		name := line[first+1 : last]
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		if strings.HasPrefix(name, "Apple Development: ") {
			return name, nil
		}
		if fallback == "" {
			fallback = name
		}
	}
	if fallback != "" {
		return fallback, nil
	}
	return "", fmt.Errorf("no codesigning identity found for team %s; install a certificate for that team or set ios.signing.identity", team)
}

// iosTeamSigningIdentity asks the keychain for the identity matching team.
// It runs only when a device build has ios.signing.team_id and no explicit identity.
func iosTeamSigningIdentity(ctx context.Context, dir, team string) (string, error) {
	output, err := runManifestCommand(ctx, dir, nil, "security", "find-identity", "-v", "-p", "codesigning")
	if err != nil {
		return "", fmt.Errorf("list codesigning identities: %s: %w", strings.TrimSpace(output), err)
	}
	return selectIOSTeamIdentity(output, team)
}
