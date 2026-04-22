package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// LikeVideo likes or unlikes a video.
// like=true to like, false to unlike.
// Requires X-Bogus.
func (c *Client) LikeVideo(ctx context.Context, videoID string, like bool) error {
	typeVal := "0"
	if like {
		typeVal = "1"
	}
	form := url.Values{}
	form.Set("aweme_id", videoID)
	form.Set("type", typeVal)

	body, err := c.apiPOST(ctx, "/api/commit/item/digg/", form, baseURL)
	if err != nil {
		return fmt.Errorf("LikeVideo: %w", err)
	}

	return checkStatusCode(body, "LikeVideo")
}

// FollowUser follows or unfollows a user by their numeric user ID.
// follow=true to follow, false to unfollow.
// Requires X-Bogus.
func (c *Client) FollowUser(ctx context.Context, userID string, follow bool) error {
	typeVal := "0"
	if follow {
		typeVal = "1"
	}
	form := url.Values{}
	form.Set("user_id", userID)
	form.Set("type", typeVal)
	form.Set("from", "0")
	form.Set("from_pre", "0")

	body, err := c.apiPOST(ctx, "/api/commit/follow/user/", form, baseURL)
	if err != nil {
		return fmt.Errorf("FollowUser: %w", err)
	}

	return checkStatusCode(body, "FollowUser")
}

// CollectVideo saves or unsaves a video.
// collect=true to save, false to unsave.
// Requires X-Bogus.
func (c *Client) CollectVideo(ctx context.Context, videoID string, collect bool) error {
	typeVal := "0"
	if collect {
		typeVal = "1"
	}
	form := url.Values{}
	form.Set("aweme_id", videoID)
	form.Set("type", typeVal)

	body, err := c.apiPOST(ctx, "/api/commit/item/collect/", form, baseURL)
	if err != nil {
		return fmt.Errorf("CollectVideo: %w", err)
	}

	return checkStatusCode(body, "CollectVideo")
}

// RepostVideo reposts a video to the authenticated user's profile.
// Requires X-Bogus.
func (c *Client) RepostVideo(ctx context.Context, videoID string) error {
	form := url.Values{}
	form.Set("aweme_id", videoID)

	body, err := c.apiPOST(ctx, "/api/repost/", form, baseURL)
	if err != nil {
		return fmt.Errorf("RepostVideo: %w", err)
	}

	return checkStatusCode(body, "RepostVideo")
}

// DeleteRepost removes a repost from the authenticated user's profile.
// Requires X-Bogus.
func (c *Client) DeleteRepost(ctx context.Context, videoID string) error {
	form := url.Values{}
	form.Set("aweme_id", videoID)

	body, err := c.apiPOST(ctx, "/api/repost/delete/", form, baseURL)
	if err != nil {
		return fmt.Errorf("DeleteRepost: %w", err)
	}

	return checkStatusCode(body, "DeleteRepost")
}

// BlockUser blocks or unblocks a user.
// block=true to block, false to unblock.
// Requires X-Bogus.
func (c *Client) BlockUser(ctx context.Context, userID string, block bool) error {
	typeVal := "0"
	if block {
		typeVal = "1"
	}
	form := url.Values{}
	form.Set("user_id", userID)
	form.Set("type", typeVal)

	body, err := c.apiPOST(ctx, "/api/commit/follow/block/", form, baseURL)
	if err != nil {
		return fmt.Errorf("BlockUser: %w", err)
	}

	return checkStatusCode(body, "BlockUser")
}

// MuteUser mutes or unmutes a user.
// mute=true to mute, false to unmute.
// Requires X-Bogus.
func (c *Client) MuteUser(ctx context.Context, userID string, mute bool) error {
	typeVal := "0"
	if mute {
		typeVal = "1"
	}
	form := url.Values{}
	form.Set("user_id", userID)
	form.Set("type", typeVal)

	body, err := c.apiPOST(ctx, "/api/commit/mute/user/", form, baseURL)
	if err != nil {
		return fmt.Errorf("MuteUser: %w", err)
	}

	return checkStatusCode(body, "MuteUser")
}

// checkStatusCode parses a generic TikTok status response and returns an error
// if status_code is non-zero or the response indicates failure.
func checkStatusCode(body []byte, op string) error {
	if len(body) == 0 {
		return fmt.Errorf("%s: %w", op, ErrEmptyResponse)
	}
	var resp struct {
		StatusCode  int    `json:"status_code"`
		StatusCode2 int    `json:"statusCode"`
		StatusMsg   string `json:"status_msg"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil // non-JSON success responses are ok
	}
	sc := resp.StatusCode
	if sc == 0 {
		sc = resp.StatusCode2
	}
	if sc != 0 {
		msg := resp.StatusMsg
		if msg == "" {
			msg = "unknown error"
		}
		return fmt.Errorf("%s: API error %d: %s", op, sc, msg)
	}
	return nil
}
