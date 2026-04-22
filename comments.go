package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// GetComments fetches a paginated list of comments on a video.
// Requires X-Bogus.
func (c *Client) GetComments(ctx context.Context, videoID string, count int, cursor int64) (*CommentPage, error) {
	if count <= 0 || count > 50 {
		count = 20
	}
	params := url.Values{}
	params.Set("aweme_id", videoID)
	params.Set("count", strconv.Itoa(count))
	params.Set("cursor", strconv.FormatInt(cursor, 10))

	body, err := c.apiGET(ctx, "/api/comment/list/", params, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GetComments %q: %w", videoID, err)
	}

	return parseCommentListResponse(body)
}

// GetReplies fetches a paginated list of replies to a specific comment.
// Requires X-Bogus.
func (c *Client) GetReplies(ctx context.Context, videoID, commentID string, count int, cursor int64) (*CommentPage, error) {
	if count <= 0 || count > 50 {
		count = 20
	}
	params := url.Values{}
	params.Set("aweme_id", videoID)
	params.Set("comment_id", commentID)
	params.Set("count", strconv.Itoa(count))
	params.Set("cursor", strconv.FormatInt(cursor, 10))

	body, err := c.apiGET(ctx, "/api/comment/list/reply/", params, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GetReplies %q: %w", commentID, err)
	}

	return parseCommentListResponse(body)
}

// PostComment posts a top-level comment on a video.
// Does NOT require X-Bogus — confirmed working with cookies only.
// text supports Unicode, emoji, and plain mentions (@username).
// For structured @mentions, use PostCommentWithMentions.
func (c *Client) PostComment(ctx context.Context, videoID, text string) (*Comment, error) {
	form := url.Values{}
	form.Set("aweme_id", videoID)
	form.Set("text", text)
	form.Set("text_extra", "[]")
	form.Set("channel_id", "0")

	body, err := c.apiPOSTNoSign(ctx, "/api/comment/publish/", form, baseURL)
	if err != nil {
		return nil, fmt.Errorf("PostComment: %w", err)
	}

	return parseCommentResponse(body)
}

// ReplyToComment posts a reply to an existing comment.
// Does NOT require X-Bogus.
func (c *Client) ReplyToComment(ctx context.Context, videoID, parentCommentID, text string) (*Comment, error) {
	form := url.Values{}
	form.Set("aweme_id", videoID)
	form.Set("text", text)
	form.Set("text_extra", "[]")
	form.Set("channel_id", "0")
	form.Set("comment_id", parentCommentID)
	form.Set("reply_comment_id", parentCommentID)
	form.Set("reply_type", "2")

	body, err := c.apiPOSTNoSign(ctx, "/api/comment/publish/", form, baseURL)
	if err != nil {
		return nil, fmt.Errorf("ReplyToComment: %w", err)
	}

	return parseCommentResponse(body)
}

// DeleteComment deletes a comment by ID.
// Requires X-Bogus.
func (c *Client) DeleteComment(ctx context.Context, videoID, commentID string) error {
	form := url.Values{}
	form.Set("aweme_id", videoID)
	form.Set("comment_id", commentID)

	_, err := c.apiPOST(ctx, "/api/comment/delete/", form, baseURL)
	if err != nil {
		return fmt.Errorf("DeleteComment: %w", err)
	}
	return nil
}

// LikeComment likes or unlikes a comment.
// like=true to like, false to unlike.
// Requires X-Bogus.
func (c *Client) LikeComment(ctx context.Context, videoID, commentID string, like bool) error {
	typeVal := "0"
	if like {
		typeVal = "1"
	}
	form := url.Values{}
	form.Set("aweme_id", videoID)
	form.Set("comment_id", commentID)
	form.Set("type", typeVal)

	_, err := c.apiPOST(ctx, "/api/comment/digg/", form, baseURL)
	if err != nil {
		return fmt.Errorf("LikeComment: %w", err)
	}
	return nil
}

// ─── comment parsing helpers ──────────────────────────────────────────────────

func parseCommentResponse(body []byte) (*Comment, error) {
	var raw struct {
		StatusCode int     `json:"status_code"`
		StatusMsg  string  `json:"status_msg"`
		Comment    Comment `json:"comment"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}
	if raw.Comment.CID == "" {
		return nil, fmt.Errorf("%w: empty comment in response (msg: %s)", ErrParseFailed, raw.StatusMsg)
	}
	return &raw.Comment, nil
}

func parseCommentListResponse(body []byte) (*CommentPage, error) {
	var raw struct {
		StatusCode  int       `json:"status_code"`
		StatusCode2 int       `json:"statusCode"`
		HasMore     int       `json:"has_more"` // TikTok returns 0/1
		Cursor      int64     `json:"cursor"`
		Total       int64     `json:"total"`
		Comments    []Comment `json:"comments"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}
	return &CommentPage{
		Comments: raw.Comments,
		HasMore:  raw.HasMore == 1,
		Cursor:   raw.Cursor,
		Total:    raw.Total,
	}, nil
}
