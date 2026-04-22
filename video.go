package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetVideo fetches a video by its ID using HTML page scraping.
// username can be any valid TikTok @handle or "@tiktok" as a placeholder —
// TikTok resolves by videoID regardless of the username in the path.
// No X-Bogus required.
func (c *Client) GetVideo(ctx context.Context, username, videoID string) (*Video, error) {
	if username == "" {
		username = "tiktok"
	}
	path := "/@" + username + "/video/" + videoID
	body, err := c.pageGET(ctx, path, baseURL)
	if err != nil {
		return nil, fmt.Errorf("GetVideo %q: %w", videoID, err)
	}

	scope, err := extractSIGI(body)
	if err != nil {
		return nil, fmt.Errorf("GetVideo %q: %w", videoID, err)
	}

	rawDetail, ok := scope["webapp.video-detail"]
	if !ok {
		return nil, fmt.Errorf("GetVideo %q: %w: webapp.video-detail not in scope", videoID, ErrNotFound)
	}

	var detail struct {
		ItemInfo struct {
			ItemStruct json.RawMessage `json:"itemStruct"`
		} `json:"itemInfo"`
		StatusCode int    `json:"statusCode"`
		StatusMsg  string `json:"statusMsg"`
	}
	if err := json.Unmarshal(rawDetail, &detail); err != nil {
		return nil, fmt.Errorf("GetVideo %q: %w: %v", videoID, ErrParseFailed, err)
	}
	if len(detail.ItemInfo.ItemStruct) == 0 || string(detail.ItemInfo.ItemStruct) == "null" {
		return nil, fmt.Errorf("GetVideo %q: %w: itemStruct is empty (statusCode=%d)", videoID, ErrNotFound, detail.StatusCode)
	}

	v, err := parseRawVideo(detail.ItemInfo.ItemStruct)
	if err != nil {
		return nil, fmt.Errorf("GetVideo %q: %w: %v", videoID, ErrParseFailed, err)
	}
	return &v, nil
}
