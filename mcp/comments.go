package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	tiktok "github.com/teslashibe/tiktok-go"
)

// GetCommentsInput is the typed input for tiktok_get_comments.
type GetCommentsInput struct {
	VideoID string `json:"video_id" jsonschema:"description=numeric TikTok video ID,required"`
	Count   int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	Cursor  int64  `json:"cursor,omitempty" jsonschema:"description=pagination cursor (0 for first page),minimum=0,default=0"`
}

func getComments(ctx context.Context, c *tiktok.Client, in GetCommentsInput) (any, error) {
	res, err := c.GetComments(ctx, in.VideoID, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Comments, formatCursorInt(res.Cursor, res.HasMore), limit), nil
}

// GetRepliesInput is the typed input for tiktok_get_replies.
type GetRepliesInput struct {
	VideoID   string `json:"video_id" jsonschema:"description=numeric TikTok video ID,required"`
	CommentID string `json:"comment_id" jsonschema:"description=parent comment ID to fetch replies for,required"`
	Count     int    `json:"count,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	Cursor    int64  `json:"cursor,omitempty" jsonschema:"description=pagination cursor (0 for first page),minimum=0,default=0"`
}

func getReplies(ctx context.Context, c *tiktok.Client, in GetRepliesInput) (any, error) {
	res, err := c.GetReplies(ctx, in.VideoID, in.CommentID, in.Count, in.Cursor)
	if err != nil {
		return nil, err
	}
	limit := in.Count
	if limit <= 0 {
		limit = 20
	}
	return mcptool.PageOf(res.Comments, formatCursorInt(res.Cursor, res.HasMore), limit), nil
}

// PostCommentInput is the typed input for tiktok_post_comment.
type PostCommentInput struct {
	VideoID string `json:"video_id" jsonschema:"description=numeric TikTok video ID to comment on,required"`
	Text    string `json:"text" jsonschema:"description=comment body (Unicode and emoji allowed),required"`
}

func postComment(ctx context.Context, c *tiktok.Client, in PostCommentInput) (any, error) {
	cm, err := c.PostComment(ctx, in.VideoID, in.Text)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "comment_id": cm.CID, "comment": cm}, nil
}

// ReplyToCommentInput is the typed input for tiktok_reply_to_comment.
type ReplyToCommentInput struct {
	VideoID         string `json:"video_id" jsonschema:"description=numeric TikTok video ID,required"`
	ParentCommentID string `json:"parent_comment_id" jsonschema:"description=ID of the comment being replied to,required"`
	Text            string `json:"text" jsonschema:"description=reply body (Unicode and emoji allowed),required"`
}

func replyToComment(ctx context.Context, c *tiktok.Client, in ReplyToCommentInput) (any, error) {
	cm, err := c.ReplyToComment(ctx, in.VideoID, in.ParentCommentID, in.Text)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "comment_id": cm.CID, "comment": cm}, nil
}

// DeleteCommentInput is the typed input for tiktok_delete_comment.
type DeleteCommentInput struct {
	VideoID   string `json:"video_id" jsonschema:"description=numeric TikTok video ID,required"`
	CommentID string `json:"comment_id" jsonschema:"description=ID of the comment to delete,required"`
}

func deleteComment(ctx context.Context, c *tiktok.Client, in DeleteCommentInput) (any, error) {
	if err := c.DeleteComment(ctx, in.VideoID, in.CommentID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "video_id": in.VideoID, "comment_id": in.CommentID}, nil
}

// LikeCommentInput is the typed input for tiktok_like_comment.
type LikeCommentInput struct {
	VideoID   string `json:"video_id" jsonschema:"description=numeric TikTok video ID,required"`
	CommentID string `json:"comment_id" jsonschema:"description=comment ID to like or unlike,required"`
	Like      bool   `json:"like,omitempty" jsonschema:"description=true to like (default); false to unlike,default=true"`
}

func likeComment(ctx context.Context, c *tiktok.Client, in LikeCommentInput) (any, error) {
	if err := c.LikeComment(ctx, in.VideoID, in.CommentID, in.Like); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "video_id": in.VideoID, "comment_id": in.CommentID, "liked": in.Like}, nil
}

var commentTools = []mcptool.Tool{
	mcptool.Define[*tiktok.Client, GetCommentsInput](
		"tiktok_get_comments",
		"Fetch a paginated list of top-level comments on a TikTok video",
		"GetComments",
		getComments,
	),
	mcptool.Define[*tiktok.Client, GetRepliesInput](
		"tiktok_get_replies",
		"Fetch a paginated list of replies to a TikTok comment",
		"GetReplies",
		getReplies,
	),
	mcptool.Define[*tiktok.Client, PostCommentInput](
		"tiktok_post_comment",
		"Post a top-level comment on a TikTok video",
		"PostComment",
		postComment,
	),
	mcptool.Define[*tiktok.Client, ReplyToCommentInput](
		"tiktok_reply_to_comment",
		"Reply to an existing TikTok comment",
		"ReplyToComment",
		replyToComment,
	),
	mcptool.Define[*tiktok.Client, DeleteCommentInput](
		"tiktok_delete_comment",
		"Delete a TikTok comment by ID (must be authored by the authenticated user)",
		"DeleteComment",
		deleteComment,
	),
	mcptool.Define[*tiktok.Client, LikeCommentInput](
		"tiktok_like_comment",
		"Like or unlike a TikTok comment",
		"LikeComment",
		likeComment,
	),
}
