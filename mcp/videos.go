package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	tiktok "github.com/teslashibe/tiktok-go"
)

// GetVideoInput is the typed input for tiktok_get_video.
type GetVideoInput struct {
	VideoID  string `json:"video_id" jsonschema:"description=numeric TikTok video ID (the digits in /@user/video/<id>),required"`
	Username string `json:"username,omitempty" jsonschema:"description=author @handle (optional; TikTok resolves by video_id regardless)"`
}

func getVideo(ctx context.Context, c *tiktok.Client, in GetVideoInput) (any, error) {
	return c.GetVideo(ctx, in.Username, in.VideoID)
}

var videoTools = []mcptool.Tool{
	mcptool.Define[*tiktok.Client, GetVideoInput](
		"tiktok_get_video",
		"Fetch a TikTok video's full metadata, stats, music, and author by video ID",
		"GetVideo",
		getVideo,
	),
}
