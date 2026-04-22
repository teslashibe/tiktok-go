package mcp_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/teslashibe/mcptool"
	tiktok "github.com/teslashibe/tiktok-go"
	ttmcp "github.com/teslashibe/tiktok-go/mcp"
)

// TestEveryClientMethodIsWrappedOrExcluded fails when a new exported method
// is added to *tiktok.Client without either being wrapped by an MCP tool
// or being added to ttmcp.Excluded with a reason. This is the drift-
// prevention mechanism: keeping the MCP surface in lockstep with the package
// API is enforced by CI rather than convention.
func TestEveryClientMethodIsWrappedOrExcluded(t *testing.T) {
	rep := mcptool.Coverage(
		reflect.TypeOf(&tiktok.Client{}),
		ttmcp.Provider{}.Tools(),
		ttmcp.Excluded,
	)
	if len(rep.Missing) > 0 {
		t.Fatalf("methods missing MCP exposure (add a tool or list in excluded.go): %v", rep.Missing)
	}
	if len(rep.UnknownExclusions) > 0 {
		t.Fatalf("excluded.go references methods that don't exist on *Client (rename?): %v", rep.UnknownExclusions)
	}
	if len(rep.Wrapped)+len(rep.Excluded) == 0 {
		t.Fatal("no wrapped or excluded methods detected — coverage helper is mis-configured")
	}
}

// TestToolsValidate verifies every tool has a non-empty name in canonical
// snake_case form, a description within length limits, and a non-nil Invoke
// + InputSchema.
func TestToolsValidate(t *testing.T) {
	if err := mcptool.ValidateTools(ttmcp.Provider{}.Tools()); err != nil {
		t.Fatal(err)
	}
}

// TestPlatformName guards against accidental rebrands.
func TestPlatformName(t *testing.T) {
	if got := (ttmcp.Provider{}).Platform(); got != "tiktok" {
		t.Errorf("Platform() = %q, want tiktok", got)
	}
}

// TestToolsHaveTiktokPrefix encodes the per-platform naming convention.
func TestToolsHaveTiktokPrefix(t *testing.T) {
	for _, tool := range (ttmcp.Provider{}).Tools() {
		if !strings.HasPrefix(tool.Name, "tiktok_") {
			t.Errorf("tool %q lacks tiktok_ prefix", tool.Name)
		}
	}
}
