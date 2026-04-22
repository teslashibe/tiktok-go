package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	tiktok "github.com/teslashibe/tiktok-go"
)

// LikeVideoInput is the typed input for tiktok_like_video.
type LikeVideoInput struct {
	VideoID string `json:"video_id" jsonschema:"description=numeric TikTok video ID,required"`
	Like    bool   `json:"like,omitempty" jsonschema:"description=true to like (default); false to unlike,default=true"`
}

func likeVideo(ctx context.Context, c *tiktok.Client, in LikeVideoInput) (any, error) {
	if err := c.LikeVideo(ctx, in.VideoID, in.Like); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "video_id": in.VideoID, "liked": in.Like}, nil
}

// FollowUserInput is the typed input for tiktok_follow_user.
type FollowUserInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric TikTok user ID (User.ID from tiktok_get_user),required"`
	Follow bool   `json:"follow,omitempty" jsonschema:"description=true to follow (default); false to unfollow,default=true"`
}

func followUser(ctx context.Context, c *tiktok.Client, in FollowUserInput) (any, error) {
	if err := c.FollowUser(ctx, in.UserID, in.Follow); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID, "following": in.Follow}, nil
}

// CollectVideoInput is the typed input for tiktok_collect_video.
type CollectVideoInput struct {
	VideoID string `json:"video_id" jsonschema:"description=numeric TikTok video ID,required"`
	Collect bool   `json:"collect,omitempty" jsonschema:"description=true to save (default); false to unsave,default=true"`
}

func collectVideo(ctx context.Context, c *tiktok.Client, in CollectVideoInput) (any, error) {
	if err := c.CollectVideo(ctx, in.VideoID, in.Collect); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "video_id": in.VideoID, "collected": in.Collect}, nil
}

// RepostVideoInput is the typed input for tiktok_repost_video.
type RepostVideoInput struct {
	VideoID string `json:"video_id" jsonschema:"description=numeric TikTok video ID to repost,required"`
}

func repostVideo(ctx context.Context, c *tiktok.Client, in RepostVideoInput) (any, error) {
	if err := c.RepostVideo(ctx, in.VideoID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "video_id": in.VideoID}, nil
}

// DeleteRepostInput is the typed input for tiktok_delete_repost.
type DeleteRepostInput struct {
	VideoID string `json:"video_id" jsonschema:"description=numeric TikTok video ID whose repost to remove,required"`
}

func deleteRepost(ctx context.Context, c *tiktok.Client, in DeleteRepostInput) (any, error) {
	if err := c.DeleteRepost(ctx, in.VideoID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "video_id": in.VideoID}, nil
}

// BlockUserInput is the typed input for tiktok_block_user.
type BlockUserInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric TikTok user ID,required"`
	Block  bool   `json:"block,omitempty" jsonschema:"description=true to block (default); false to unblock,default=true"`
}

func blockUser(ctx context.Context, c *tiktok.Client, in BlockUserInput) (any, error) {
	if err := c.BlockUser(ctx, in.UserID, in.Block); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID, "blocked": in.Block}, nil
}

// MuteUserInput is the typed input for tiktok_mute_user.
type MuteUserInput struct {
	UserID string `json:"user_id" jsonschema:"description=numeric TikTok user ID,required"`
	Mute   bool   `json:"mute,omitempty" jsonschema:"description=true to mute (default); false to unmute,default=true"`
}

func muteUser(ctx context.Context, c *tiktok.Client, in MuteUserInput) (any, error) {
	if err := c.MuteUser(ctx, in.UserID, in.Mute); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID, "muted": in.Mute}, nil
}

var socialTools = []mcptool.Tool{
	mcptool.Define[*tiktok.Client, LikeVideoInput](
		"tiktok_like_video",
		"Like or unlike a TikTok video",
		"LikeVideo",
		likeVideo,
	),
	mcptool.Define[*tiktok.Client, FollowUserInput](
		"tiktok_follow_user",
		"Follow or unfollow a TikTok user by numeric user ID",
		"FollowUser",
		followUser,
	),
	mcptool.Define[*tiktok.Client, CollectVideoInput](
		"tiktok_collect_video",
		"Save or unsave a TikTok video to the authenticated user's collection",
		"CollectVideo",
		collectVideo,
	),
	mcptool.Define[*tiktok.Client, RepostVideoInput](
		"tiktok_repost_video",
		"Repost a TikTok video to the authenticated user's profile",
		"RepostVideo",
		repostVideo,
	),
	mcptool.Define[*tiktok.Client, DeleteRepostInput](
		"tiktok_delete_repost",
		"Remove a previously-reposted TikTok video from the authenticated user's profile",
		"DeleteRepost",
		deleteRepost,
	),
	mcptool.Define[*tiktok.Client, BlockUserInput](
		"tiktok_block_user",
		"Block or unblock a TikTok user by numeric user ID",
		"BlockUser",
		blockUser,
	),
	mcptool.Define[*tiktok.Client, MuteUserInput](
		"tiktok_mute_user",
		"Mute or unmute a TikTok user by numeric user ID",
		"MuteUser",
		muteUser,
	),
}
