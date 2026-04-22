package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// GetUser fetches a user's profile and stats by @username.
// Uses HTML page scraping — no X-Bogus required.
func (c *Client) GetUser(ctx context.Context, username string) (*UserInfo, error) {
	path := "/@" + username
	body, err := c.pageGET(ctx, path, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GetUser %q: %w", username, err)
	}

	scope, err := extractSIGI(body)
	if err != nil {
		return nil, fmt.Errorf("GetUser %q: %w", username, err)
	}

	rawDetail, ok := scope["webapp.user-detail"]
	if !ok {
		return nil, fmt.Errorf("GetUser %q: %w: webapp.user-detail not in scope", username, ErrNotFound)
	}

	var detail struct {
		UserInfo struct {
			User  User      `json:"user"`
			Stats UserStats `json:"stats"`
		} `json:"userInfo"`
		StatusCode int    `json:"statusCode"`
		StatusMsg  string `json:"statusMsg"`
	}
	if err := json.Unmarshal(rawDetail, &detail); err != nil {
		return nil, fmt.Errorf("GetUser %q: %w: %v", username, ErrParseFailed, err)
	}
	if detail.UserInfo.User.UniqueID == "" {
		return nil, fmt.Errorf("GetUser %q: %w: user not found (statusCode=%d)", username, ErrNotFound, detail.StatusCode)
	}

	return &UserInfo{
		User:  detail.UserInfo.User,
		Stats: detail.UserInfo.Stats,
	}, nil
}

// GetUserVideos fetches a paginated list of videos posted by a user.
// secUID is the user's secUid string (obtainable from GetUser).
// Requires X-Bogus.
func (c *Client) GetUserVideos(ctx context.Context, secUID string, count int, cursor int64) (*VideoPage, error) {
	if count <= 0 || count > 35 {
		count = 16
	}
	params := url.Values{}
	params.Set("secUid", secUID)
	params.Set("count", strconv.Itoa(count))
	params.Set("cursor", strconv.FormatInt(cursor, 10))

	body, err := c.apiGET(ctx, "/api/post/item_list/", params, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GetUserVideos: %w", err)
	}

	return parseVideoListResponse(body, cursor)
}

// GetLikedVideos fetches the liked videos for a user (public likes only).
// Requires X-Bogus.
func (c *Client) GetLikedVideos(ctx context.Context, secUID string, count int, cursor int64) (*VideoPage, error) {
	if count <= 0 || count > 35 {
		count = 16
	}
	params := url.Values{}
	params.Set("secUid", secUID)
	params.Set("count", strconv.Itoa(count))
	params.Set("cursor", strconv.FormatInt(cursor, 10))

	body, err := c.apiGET(ctx, "/api/like/item_list/", params, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GetLikedVideos: %w", err)
	}

	return parseVideoListResponse(body, cursor)
}

// GetSavedVideos fetches the collected/saved videos for the authenticated user.
// Requires X-Bogus.
func (c *Client) GetSavedVideos(ctx context.Context, count int, cursor int64) (*VideoPage, error) {
	if count <= 0 || count > 35 {
		count = 16
	}
	params := url.Values{}
	params.Set("count", strconv.Itoa(count))
	params.Set("cursor", strconv.FormatInt(cursor, 10))

	body, err := c.apiGET(ctx, "/api/collection/item_list/", params, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GetSavedVideos: %w", err)
	}

	return parseVideoListResponse(body, cursor)
}

// GetFollowers fetches a paginated list of a user's followers.
// Requires X-Bogus.
func (c *Client) GetFollowers(ctx context.Context, secUID string, count int, minCursor string) (*UserPage, error) {
	if count <= 0 || count > 50 {
		count = 30
	}
	params := url.Values{}
	params.Set("secUid", secUID)
	params.Set("count", strconv.Itoa(count))
	if minCursor == "" {
		minCursor = "0"
	}
	params.Set("minCursor", minCursor)
	params.Set("maxCursor", "0")

	body, err := c.apiGET(ctx, "/api/user/follower/list/", params, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GetFollowers: %w", err)
	}

	return parseUserListResponse(body)
}

// GetFollowing fetches a paginated list of users that the given user follows.
// Requires X-Bogus.
func (c *Client) GetFollowing(ctx context.Context, secUID string, count int, minCursor string) (*UserPage, error) {
	if count <= 0 || count > 50 {
		count = 30
	}
	params := url.Values{}
	params.Set("secUid", secUID)
	params.Set("count", strconv.Itoa(count))
	if minCursor == "" {
		minCursor = "0"
	}
	params.Set("minCursor", minCursor)
	params.Set("maxCursor", "0")

	body, err := c.apiGET(ctx, "/api/user/following/list/", params, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GetFollowing: %w", err)
	}

	return parseUserListResponse(body)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func parseVideoListResponse(body []byte, prevCursor int64) (*VideoPage, error) {
	var raw struct {
		StatusCode  int               `json:"statusCode"`
		Status_Code int               `json:"status_code"`
		HasMore     bool              `json:"hasMore"`
		ItemList    []json.RawMessage `json:"itemList"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}
	videos, err := parseItemList(raw.ItemList)
	if err != nil {
		return nil, err
	}
	return &VideoPage{
		Videos:  videos,
		HasMore: raw.HasMore,
		Cursor:  prevCursor + int64(len(videos)),
	}, nil
}

func parseUserListResponse(body []byte) (*UserPage, error) {
	var raw struct {
		StatusCode  int    `json:"statusCode"`
		Status_Code int    `json:"status_code"`
		HasMore     bool   `json:"hasMore"`
		HasMore2    int    `json:"has_more"` // alternate field
		MinCursor   string `json:"minCursor"`
		MaxCursor   string `json:"maxCursor"`
		UserList    []struct {
			User  User      `json:"user"`
			Stats UserStats `json:"stats"`
		} `json:"userList"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}
	users := make([]UserInfo, 0, len(raw.UserList))
	for _, u := range raw.UserList {
		users = append(users, UserInfo{User: u.User, Stats: u.Stats})
	}
	hasMore := raw.HasMore || raw.HasMore2 == 1
	return &UserPage{
		Users:     users,
		HasMore:   hasMore,
		MinCursor: raw.MinCursor,
		MaxCursor: raw.MaxCursor,
	}, nil
}
