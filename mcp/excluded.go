package mcp

// Excluded enumerates exported methods on *tiktok.Client that are
// intentionally not exposed via MCP. Each entry must have a non-empty reason.
//
// The coverage test in mcp_test.go fails if any exported method on *Client is
// neither wrapped by a Tool nor present in this map (or vice-versa: if an
// entry here doesn't correspond to a real method).
//
// When the underlying client gains a new method:
//   - prefer to add an MCP tool for it (see users.go / videos.go / etc.)
//   - if the method is unsuitable for an agent (internal observability,
//     auth-only helper, etc.), add it here with a reason
var Excluded = map[string]string{
	"RateLimit": "internal observability; surfaced via the host application's MCP middleware, not as a callable tool",
}
