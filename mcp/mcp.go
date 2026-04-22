// Package mcp exposes the tiktok-go [tiktok.Client] surface as a set of
// MCP (Model Context Protocol) tools that any host application can mount on
// its own MCP server.
//
// All tools wrap exported methods on *tiktok.Client. Each tool is defined
// via [mcptool.Define] so the JSON input schema is reflected from the typed
// input struct — no hand-maintained schemas, no drift.
//
// Usage from a host application:
//
//	import (
//	    "github.com/teslashibe/mcptool"
//	    tiktok "github.com/teslashibe/tiktok-go"
//	    ttmcp "github.com/teslashibe/tiktok-go/mcp"
//	)
//
//	client, _ := tiktok.New(tiktok.Cookies{...})
//	for _, tool := range ttmcp.Provider{}.Tools() {
//	    // register tool with your MCP server, passing client as the client arg
//	    // when invoking
//	}
//
// The [Excluded] map documents methods on *Client that are intentionally not
// exposed via MCP, with a one-line reason. The coverage test in mcp_test.go
// fails if a new exported method is added without either being wrapped by a
// tool or appearing in [Excluded].
package mcp

import "github.com/teslashibe/mcptool"

// Provider implements [mcptool.Provider] for tiktok-go. The zero value is
// ready to use.
type Provider struct{}

// Platform returns "tiktok".
func (Provider) Platform() string { return "tiktok" }

// Tools returns every tiktok-go MCP tool, in registration order.
func (Provider) Tools() []mcptool.Tool {
	out := make([]mcptool.Tool, 0,
		len(userTools)+len(videoTools)+len(searchTools)+
			len(hashtagTools)+len(feedTools)+len(commentTools)+len(socialTools))
	out = append(out, userTools...)
	out = append(out, videoTools...)
	out = append(out, searchTools...)
	out = append(out, hashtagTools...)
	out = append(out, feedTools...)
	out = append(out, commentTools...)
	out = append(out, socialTools...)
	return out
}
