package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// SearchLive searches live rooms by keyword.
// This is a Tier 1 endpoint — no X-Bogus required.
func (c *Client) SearchLive(ctx context.Context, keyword string, count int, cursor string) (*LiveSearchPage, error) {
	if count <= 0 || count > 20 {
		count = 10
	}
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("count", strconv.Itoa(count))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.apiGETNoSign(ctx, "/api/search/live/full/",
		params, baseURL+"/search?q="+url.QueryEscape(keyword)+"&type=live")
	if err != nil {
		return nil, fmt.Errorf("SearchLive %q: %w", keyword, err)
	}

	return parseLiveSearchResponse(body)
}

// SearchVideos searches for videos matching a keyword.
// Requires X-Bogus.
func (c *Client) SearchVideos(ctx context.Context, keyword string, count int, cursor string) (*VideoPage, error) {
	if count <= 0 || count > 20 {
		count = 10
	}
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("count", strconv.Itoa(count))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.apiGET(ctx, "/api/search/item/full/",
		params, baseURL+"/search?q="+url.QueryEscape(keyword)+"&type=video")
	if err != nil {
		return nil, fmt.Errorf("SearchVideos %q: %w", keyword, err)
	}

	var raw struct {
		Status_Code int               `json:"status_code"`
		HasMore     int               `json:"has_more"` // TikTok returns 0/1 not bool
		Cursor      string            `json:"cursor"`
		Data        []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("SearchVideos: %w: %v", ErrParseFailed, err)
	}

	var videos []Video
	for _, item := range raw.Data {
		// search results wrap items in {type:1, item:{...}} objects
		var wrapper struct {
			Type int             `json:"type"`
			Item json.RawMessage `json:"item"`
		}
		if err := json.Unmarshal(item, &wrapper); err != nil {
			continue
		}
		if len(wrapper.Item) == 0 {
			continue
		}
		v, err := parseRawVideo(wrapper.Item)
		if err != nil {
			continue
		}
		videos = append(videos, v)
	}

	nextCursor := raw.Cursor
	var nextCursorInt int64
	if nextCursor != "" {
		if n, err := strconv.ParseInt(nextCursor, 10, 64); err == nil {
			nextCursorInt = n
		}
	}

	return &VideoPage{
		Videos:  videos,
		HasMore: raw.HasMore != 0,
		Cursor:  nextCursorInt,
	}, nil
}

// SearchUsers searches for users matching a keyword.
// Requires X-Bogus.
func (c *Client) SearchUsers(ctx context.Context, keyword string, count int, cursor string) (*UserPage, error) {
	if count <= 0 || count > 20 {
		count = 10
	}
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("count", strconv.Itoa(count))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.apiGET(ctx, "/api/search/user/full/",
		params, baseURL+"/search?q="+url.QueryEscape(keyword)+"&type=user")
	if err != nil {
		return nil, fmt.Errorf("SearchUsers %q: %w", keyword, err)
	}

	var raw struct {
		Status_Code int               `json:"status_code"`
		HasMore     int               `json:"has_more"` // TikTok returns 0/1 not bool
		Cursor      string            `json:"cursor"`
		Data        []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("SearchUsers: %w: %v", ErrParseFailed, err)
	}

	var users []UserInfo
	for _, item := range raw.Data {
		var wrapper struct {
			Type     int `json:"type"`
			UserInfo struct {
				User  User      `json:"user"`
				Stats UserStats `json:"stats"`
			} `json:"userInfo"`
		}
		if err := json.Unmarshal(item, &wrapper); err != nil {
			continue
		}
		if wrapper.UserInfo.User.UniqueID == "" {
			continue
		}
		users = append(users, UserInfo{User: wrapper.UserInfo.User, Stats: wrapper.UserInfo.Stats})
	}

	return &UserPage{
		Users:     users,
		HasMore:   raw.HasMore != 0,
		MinCursor: raw.Cursor,
	}, nil
}

// ─── Live search parsing ──────────────────────────────────────────────────────

// rawLiveResult is the wire format for a single live search result.
type rawLiveResult struct {
	LiveInfo struct {
		RawData  string          `json:"raw_data"`
		RoomInfo json.RawMessage `json:"room_info"`
	} `json:"live_info"`
}

type rawLiveRoom struct {
	ID            int64  `json:"id"`
	IDStr         string `json:"id_str"`
	Title         string `json:"title"`
	Status        int    `json:"status"`
	UserCount     int    `json:"user_count"`
	LikeCount     int64  `json:"like_count"`
	StartTime     int64  `json:"start_time"`
	LiveRoomMode  int    `json:"live_room_mode"`
	AgeRestricted bool   `json:"age_restricted"`

	Owner struct {
		ID             string `json:"id"`
		IDStr          string `json:"id_str"`
		Nickname       string `json:"nickname"`
		DisplayID      string `json:"display_id"`
		SecUID         string `json:"sec_uid"`
		BioDescription string `json:"bio_description"`
		AvatarThumb    struct {
			URLList []string `json:"url_list"`
		} `json:"avatar_thumb"`
		AvatarMedium struct {
			URLList []string `json:"url_list"`
		} `json:"avatar_medium"`
		AvatarLarge struct {
			URLList []string `json:"url_list"`
		} `json:"avatar_large"`
	} `json:"owner"`

	Stats struct {
		TotalUser    int `json:"total_user"`
		EnterCount   int `json:"enter_count"`
		ShareCount   int `json:"share_count"`
		CommentCount int `json:"comment_count"`
	} `json:"stats"`

	Hashtag struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	} `json:"hashtag"`

	Cover struct {
		URLList []string `json:"url_list"`
	} `json:"cover"`

	StreamURL LiveStreamURL `json:"stream_url"`
}

func parseLiveSearchResponse(body []byte) (*LiveSearchPage, error) {
	var raw struct {
		Status_Code int             `json:"status_code"`
		HasMore     int             `json:"has_more"` // TikTok returns 0/1 not bool
		Cursor      json.RawMessage `json:"cursor"`   // may be int or string
		Data        []rawLiveResult `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}

	// Normalize cursor: accept both "123" and 123
	var cursorStr string
	if len(raw.Cursor) > 0 {
		if raw.Cursor[0] == '"' {
			json.Unmarshal(raw.Cursor, &cursorStr) //nolint:errcheck
		} else {
			var n int64
			if err := json.Unmarshal(raw.Cursor, &n); err == nil {
				cursorStr = strconv.FormatInt(n, 10)
			}
		}
	}

	rooms := make([]LiveRoom, 0, len(raw.Data))
	for _, res := range raw.Data {
		if res.LiveInfo.RawData == "" {
			continue
		}
		var rl rawLiveRoom
		if err := json.Unmarshal([]byte(res.LiveInfo.RawData), &rl); err != nil {
			continue
		}
		room := LiveRoom{
			ID:            rl.ID,
			IDStr:         rl.IDStr,
			Title:         rl.Title,
			Status:        rl.Status,
			UserCount:     rl.UserCount,
			LikeCount:     rl.LikeCount,
			StartTime:     rl.StartTime,
			LiveRoomMode:  rl.LiveRoomMode,
			AgeRestricted: rl.AgeRestricted,
			Stats: LiveStats{
				TotalUser:    rl.Stats.TotalUser,
				EnterCount:   rl.Stats.EnterCount,
				ShareCount:   rl.Stats.ShareCount,
				CommentCount: rl.Stats.CommentCount,
			},
			Hashtag: LiveHashtag{
				ID:    rl.Hashtag.ID,
				Title: rl.Hashtag.Title,
			},
			StreamURL: rl.StreamURL,
		}
		if len(rl.Cover.URLList) > 0 {
			room.Cover = rl.Cover.URLList[0]
		}
		room.Owner = LiveOwner{
			ID:             rl.Owner.ID,
			IDStr:          rl.Owner.IDStr,
			Nickname:       rl.Owner.Nickname,
			DisplayID:      rl.Owner.DisplayID,
			SecUID:         rl.Owner.SecUID,
			BioDescription: rl.Owner.BioDescription,
		}
		if len(rl.Owner.AvatarThumb.URLList) > 0 {
			room.Owner.AvatarThumb = rl.Owner.AvatarThumb.URLList[0]
		}
		if len(rl.Owner.AvatarMedium.URLList) > 0 {
			room.Owner.AvatarMedium = rl.Owner.AvatarMedium.URLList[0]
		}
		if len(rl.Owner.AvatarLarge.URLList) > 0 {
			room.Owner.AvatarLarge = rl.Owner.AvatarLarge.URLList[0]
		}
		rooms = append(rooms, room)
	}

	return &LiveSearchPage{
		Rooms:   rooms,
		HasMore: raw.HasMore != 0,
		Cursor:  cursorStr,
	}, nil
}
