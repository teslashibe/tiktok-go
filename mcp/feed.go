package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	tiktok "github.com/teslashibe/tiktok-go"
)

// ForYouFeedInput is the typed input for tiktok_for_you_feed.
type ForYouFeedInput struct {
	Count  int   `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=35,default=16"`
	Cursor int64 `json:"cursor,omitempty" jsonschema:"description=pagination cursor (0 for first page),minimum=0,default=0"`
}

func forYouFeed(ctx context.Context, c *tiktok.Client, in ForYouFeedInput) (any, error) {
	res, err := c.ForYouFeed(ctx, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 16
	}
	return mcptool.PageOf(res.Videos, formatCursorInt(res.Cursor, res.HasMore), limit), nil
}

// FollowingFeedInput is the typed input for tiktok_following_feed.
type FollowingFeedInput struct {
	Count  int   `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=35,default=16"`
	Cursor int64 `json:"cursor,omitempty" jsonschema:"description=pagination cursor (0 for first page),minimum=0,default=0"`
}

func followingFeed(ctx context.Context, c *tiktok.Client, in FollowingFeedInput) (any, error) {
	res, err := c.FollowingFeed(ctx, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 16
	}
	return mcptool.PageOf(res.Videos, formatCursorInt(res.Cursor, res.HasMore), limit), nil
}

// TrendingFeedInput is the typed input for tiktok_trending_feed.
type TrendingFeedInput struct {
	Count  int   `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=35,default=16"`
	Cursor int64 `json:"cursor,omitempty" jsonschema:"description=pagination cursor (0 for first page),minimum=0,default=0"`
}

func trendingFeed(ctx context.Context, c *tiktok.Client, in TrendingFeedInput) (any, error) {
	res, err := c.TrendingFeed(ctx, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 16
	}
	return mcptool.PageOf(res.Videos, formatCursorInt(res.Cursor, res.HasMore), limit), nil
}

var feedTools = []mcptool.Tool{
	mcptool.Define[*tiktok.Client, ForYouFeedInput](
		"tiktok_for_you_feed",
		"Fetch a page of TikTok For You Page (FYP) videos",
		"ForYouFeed",
		forYouFeed,
	),
	mcptool.Define[*tiktok.Client, FollowingFeedInput](
		"tiktok_following_feed",
		"Fetch a page of videos from accounts the authenticated TikTok user follows",
		"FollowingFeed",
		followingFeed,
	),
	mcptool.Define[*tiktok.Client, TrendingFeedInput](
		"tiktok_trending_feed",
		"Fetch a page of TikTok Explore/Trending videos",
		"TrendingFeed",
		trendingFeed,
	),
}
