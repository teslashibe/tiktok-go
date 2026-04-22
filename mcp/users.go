package mcp

import (
	"context"
	"strconv"

	"github.com/teslashibe/mcptool"
	tiktok "github.com/teslashibe/tiktok-go"
)

// GetUserInput is the typed input for tiktok_get_user.
type GetUserInput struct {
	Username string `json:"username" jsonschema:"description=TikTok @handle without the leading @,required"`
}

func getUser(ctx context.Context, c *tiktok.Client, in GetUserInput) (any, error) {
	return c.GetUser(ctx, in.Username)
}

// GetUserVideosInput is the typed input for tiktok_get_user_videos.
type GetUserVideosInput struct {
	SecUID string `json:"sec_uid" jsonschema:"description=user secUid (obtain via tiktok_get_user),required"`
	Count  int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=35,default=16"`
	Cursor int64  `json:"cursor,omitempty" jsonschema:"description=pagination cursor (0 for first page),minimum=0,default=0"`
}

func getUserVideos(ctx context.Context, c *tiktok.Client, in GetUserVideosInput) (any, error) {
	res, err := c.GetUserVideos(ctx, in.SecUID, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 16
	}
	return mcptool.PageOf(res.Videos, formatCursorInt(res.Cursor, res.HasMore), limit), nil
}

// GetLikedVideosInput is the typed input for tiktok_get_liked_videos.
type GetLikedVideosInput struct {
	SecUID string `json:"sec_uid" jsonschema:"description=user secUid (obtain via tiktok_get_user),required"`
	Count  int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=35,default=16"`
	Cursor int64  `json:"cursor,omitempty" jsonschema:"description=pagination cursor (0 for first page),minimum=0,default=0"`
}

func getLikedVideos(ctx context.Context, c *tiktok.Client, in GetLikedVideosInput) (any, error) {
	res, err := c.GetLikedVideos(ctx, in.SecUID, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 16
	}
	return mcptool.PageOf(res.Videos, formatCursorInt(res.Cursor, res.HasMore), limit), nil
}

// GetSavedVideosInput is the typed input for tiktok_get_saved_videos.
type GetSavedVideosInput struct {
	Count  int   `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=35,default=16"`
	Cursor int64 `json:"cursor,omitempty" jsonschema:"description=pagination cursor (0 for first page),minimum=0,default=0"`
}

func getSavedVideos(ctx context.Context, c *tiktok.Client, in GetSavedVideosInput) (any, error) {
	res, err := c.GetSavedVideos(ctx, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 16
	}
	return mcptool.PageOf(res.Videos, formatCursorInt(res.Cursor, res.HasMore), limit), nil
}

// GetFollowersInput is the typed input for tiktok_get_followers.
type GetFollowersInput struct {
	SecUID    string `json:"sec_uid" jsonschema:"description=user secUid (obtain via tiktok_get_user),required"`
	Count     int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=30"`
	MinCursor string `json:"min_cursor,omitempty" jsonschema:"description=pagination cursor returned as next_cursor by a prior call (empty for first page)"`
}

func getFollowers(ctx context.Context, c *tiktok.Client, in GetFollowersInput) (any, error) {
	res, err := c.GetFollowers(ctx, in.SecUID, in.Count, in.MinCursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 30
	}
	return mcptool.PageOf(res.Users, nextStringCursor(res.MinCursor, res.HasMore), limit), nil
}

// GetFollowingInput is the typed input for tiktok_get_following.
type GetFollowingInput struct {
	SecUID    string `json:"sec_uid" jsonschema:"description=user secUid (obtain via tiktok_get_user),required"`
	Count     int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=30"`
	MinCursor string `json:"min_cursor,omitempty" jsonschema:"description=pagination cursor returned as next_cursor by a prior call (empty for first page)"`
}

func getFollowing(ctx context.Context, c *tiktok.Client, in GetFollowingInput) (any, error) {
	res, err := c.GetFollowing(ctx, in.SecUID, in.Count, in.MinCursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 30
	}
	return mcptool.PageOf(res.Users, nextStringCursor(res.MinCursor, res.HasMore), limit), nil
}

// formatCursorInt returns the cursor as a string when more pages exist; an
// empty string signals the agent that pagination is exhausted.
func formatCursorInt(cursor int64, hasMore bool) string {
	if !hasMore {
		return ""
	}
	return strconv.FormatInt(cursor, 10)
}

// nextStringCursor returns the cursor when more pages exist; empty otherwise.
func nextStringCursor(cursor string, hasMore bool) string {
	if !hasMore {
		return ""
	}
	return cursor
}

var userTools = []mcptool.Tool{
	mcptool.Define[*tiktok.Client, GetUserInput](
		"tiktok_get_user",
		"Fetch a TikTok user profile and stats by @handle (without the leading @)",
		"GetUser",
		getUser,
	),
	mcptool.Define[*tiktok.Client, GetUserVideosInput](
		"tiktok_get_user_videos",
		"Fetch a paginated list of videos posted by a TikTok user (by secUid)",
		"GetUserVideos",
		getUserVideos,
	),
	mcptool.Define[*tiktok.Client, GetLikedVideosInput](
		"tiktok_get_liked_videos",
		"Fetch the public liked videos for a TikTok user (by secUid)",
		"GetLikedVideos",
		getLikedVideos,
	),
	mcptool.Define[*tiktok.Client, GetSavedVideosInput](
		"tiktok_get_saved_videos",
		"Fetch the collected/saved videos for the authenticated TikTok user",
		"GetSavedVideos",
		getSavedVideos,
	),
	mcptool.Define[*tiktok.Client, GetFollowersInput](
		"tiktok_get_followers",
		"Fetch a paginated list of a TikTok user's followers (by secUid)",
		"GetFollowers",
		getFollowers,
	),
	mcptool.Define[*tiktok.Client, GetFollowingInput](
		"tiktok_get_following",
		"Fetch a paginated list of users a TikTok user follows (by secUid)",
		"GetFollowing",
		getFollowing,
	),
}
