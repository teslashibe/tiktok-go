package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// GetHashtag fetches metadata and stats for a hashtag by name (without #).
// Requires X-Bogus.
func (c *Client) GetHashtag(ctx context.Context, name string) (*ChallengeInfo, error) {
	params := url.Values{}
	params.Set("challengeName", name)

	body, err := c.apiGET(ctx, "/api/challenge/detail/",
		params, baseURL+"/tag/"+url.QueryEscape(name))
	if err != nil {
		return nil, fmt.Errorf("GetHashtag %q: %w", name, err)
	}

	var raw struct {
		StatusCode    int `json:"statusCode"`
		ChallengeInfo struct {
			Challenge Challenge      `json:"challenge"`
			Stats     ChallengeStats `json:"stats"`
		} `json:"challengeInfo"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("GetHashtag %q: %w: %v", name, ErrParseFailed, err)
	}
	if raw.ChallengeInfo.Challenge.ID == "" {
		return nil, fmt.Errorf("GetHashtag %q: %w", name, ErrNotFound)
	}

	return &ChallengeInfo{
		Challenge: raw.ChallengeInfo.Challenge,
		Stats:     raw.ChallengeInfo.Stats,
	}, nil
}

// GetHashtagVideos fetches a paginated list of videos for a hashtag by challenge ID.
// Use GetHashtag first to resolve a name → challenge ID.
// Requires X-Bogus.
func (c *Client) GetHashtagVideos(ctx context.Context, challengeID string, count int, cursor int64) (*VideoPage, error) {
	if count <= 0 || count > 35 {
		count = 20
	}
	params := url.Values{}
	params.Set("challengeID", challengeID)
	params.Set("count", strconv.Itoa(count))
	params.Set("cursor", strconv.FormatInt(cursor, 10))

	body, err := c.apiGET(ctx, "/api/challenge/item_list/", params, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GetHashtagVideos %q: %w", challengeID, err)
	}

	return parseVideoListResponse(body, cursor)
}
