package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	tiktok "github.com/teslashibe/tiktok-go"
)

// GetHashtagInput is the typed input for tiktok_get_hashtag.
type GetHashtagInput struct {
	Name string `json:"name" jsonschema:"description=hashtag name without the leading #,required"`
}

func getHashtag(ctx context.Context, c *tiktok.Client, in GetHashtagInput) (any, error) {
	return c.GetHashtag(ctx, in.Name)
}

// GetHashtagVideosInput is the typed input for tiktok_get_hashtag_videos.
type GetHashtagVideosInput struct {
	ChallengeID string `json:"challenge_id" jsonschema:"description=numeric hashtag/challenge ID (use tiktok_get_hashtag to resolve a name),required"`
	Count       int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=35,default=20"`
	Cursor      int64  `json:"cursor,omitempty" jsonschema:"description=pagination cursor (0 for first page),minimum=0,default=0"`
}

func getHashtagVideos(ctx context.Context, c *tiktok.Client, in GetHashtagVideosInput) (any, error) {
	res, err := c.GetHashtagVideos(ctx, in.ChallengeID, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Videos, formatCursorInt(res.Cursor, res.HasMore), limit), nil
}

var hashtagTools = []mcptool.Tool{
	mcptool.Define[*tiktok.Client, GetHashtagInput](
		"tiktok_get_hashtag",
		"Fetch a TikTok hashtag's metadata and stats by name (without leading #)",
		"GetHashtag",
		getHashtag,
	),
	mcptool.Define[*tiktok.Client, GetHashtagVideosInput](
		"tiktok_get_hashtag_videos",
		"Fetch a paginated list of videos for a TikTok hashtag (by challenge ID)",
		"GetHashtagVideos",
		getHashtagVideos,
	),
}
