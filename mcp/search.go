package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	tiktok "github.com/teslashibe/tiktok-go"
)

// SearchVideosInput is the typed input for tiktok_search_videos.
type SearchVideosInput struct {
	Keyword string `json:"keyword" jsonschema:"description=search query,required"`
	Count   int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=20,default=10"`
	Cursor  string `json:"cursor,omitempty" jsonschema:"description=pagination cursor returned as next_cursor by a prior call"`
}

func searchVideos(ctx context.Context, c *tiktok.Client, in SearchVideosInput) (any, error) {
	res, err := c.SearchVideos(ctx, in.Keyword, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 10
	}
	return mcptool.PageOf(res.Videos, formatCursorInt(res.Cursor, res.HasMore), limit), nil
}

// SearchUsersInput is the typed input for tiktok_search_users.
type SearchUsersInput struct {
	Keyword string `json:"keyword" jsonschema:"description=search query,required"`
	Count   int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=20,default=10"`
	Cursor  string `json:"cursor,omitempty" jsonschema:"description=pagination cursor returned as next_cursor by a prior call"`
}

func searchUsers(ctx context.Context, c *tiktok.Client, in SearchUsersInput) (any, error) {
	res, err := c.SearchUsers(ctx, in.Keyword, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 10
	}
	return mcptool.PageOf(res.Users, nextStringCursor(res.MinCursor, res.HasMore), limit), nil
}

// SearchLiveInput is the typed input for tiktok_search_live.
type SearchLiveInput struct {
	Keyword string `json:"keyword" jsonschema:"description=search query,required"`
	Count   int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=20,default=10"`
	Cursor  string `json:"cursor,omitempty" jsonschema:"description=pagination cursor returned as next_cursor by a prior call"`
}

func searchLive(ctx context.Context, c *tiktok.Client, in SearchLiveInput) (any, error) {
	res, err := c.SearchLive(ctx, in.Keyword, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 10
	}
	return mcptool.PageOf(res.Rooms, nextStringCursor(res.Cursor, res.HasMore), limit), nil
}

var searchTools = []mcptool.Tool{
	mcptool.Define[*tiktok.Client, SearchVideosInput](
		"tiktok_search_videos",
		"Search TikTok videos by keyword",
		"SearchVideos",
		searchVideos,
	),
	mcptool.Define[*tiktok.Client, SearchUsersInput](
		"tiktok_search_users",
		"Search TikTok users by keyword",
		"SearchUsers",
		searchUsers,
	),
	mcptool.Define[*tiktok.Client, SearchLiveInput](
		"tiktok_search_live",
		"Search active TikTok LIVE rooms by keyword",
		"SearchLive",
		searchLive,
	),
}
