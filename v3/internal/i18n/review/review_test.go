package review

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Two stores on the same cache dir stand in for the workbench and an MCP
// server running concurrently: neither may serve stale reads or clobber the
// other's records with a process-lifetime memo.
func TestStoreSeesOtherStoreWritesAndDoesNotClobberThem(t *testing.T) {
	// Given
	dir := t.TempDir()
	workbench := NewStore(dir)
	agent := NewStore(dir)
	_ = workbench.Get("de", "x.mdx") // would prime a memo if one existed

	// When the agent records a machine translation
	require.NoError(t, agent.SetStatus("de", "x.mdx", "needs_review", Provenance{Method: "mcp", Model: "claude-test"}))

	// Then the workbench sees it immediately
	status, _, ok := workbench.Status("de", "x.mdx")
	require.True(t, ok)
	require.Equal(t, "needs_review", status)

	// And a later workbench write to another file must not clobber it
	require.NoError(t, workbench.SetViewed("de", "y.mdx", true))
	record := agent.Get("de", "x.mdx")
	require.Equal(t, "needs_review", record.Status)
	require.Equal(t, Provenance{Method: "mcp", Model: "claude-test"}, record.Provenance)
}
