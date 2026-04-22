package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// rawFeedResponse mirrors the JSON envelope returned by /api/recommend/item_list/
// and similar feed endpoints.
type rawFeedResponse struct {
	StatusCode  int               `json:"statusCode"`
	Status_Code int               `json:"status_code"`
	HasMore     bool              `json:"hasMore"`
	ItemList    []json.RawMessage `json:"itemList"`
	Extra       json.RawMessage   `json:"extra"`
}

// ForYouFeed fetches a page of For You Page videos.
// cursor 0 returns the first page; use the returned FeedPage.Cursor for subsequent pages.
// count is capped at 35 per TikTok limits (16 is a safe default).
func (c *Client) ForYouFeed(ctx context.Context, count int, cursor int64) (*FeedPage, error) {
	if count <= 0 || count > 35 {
		count = 16
	}
	params := url.Values{}
	params.Set("count", strconv.Itoa(count))
	params.Set("cursor", strconv.FormatInt(cursor, 10))

	body, err := c.apiGETNoSign(ctx, "/api/recommend/item_list/",
		params, baseURL+"/foryou")
	if err != nil {
		return nil, fmt.Errorf("ForYouFeed: %w", err)
	}

	var raw rawFeedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("ForYouFeed: %w: %v", ErrParseFailed, err)
	}

	videos, err := parseItemList(raw.ItemList)
	if err != nil {
		return nil, fmt.Errorf("ForYouFeed: %w", err)
	}

	return &FeedPage{
		Videos:  videos,
		HasMore: raw.HasMore,
		Cursor:  cursor + int64(len(videos)),
	}, nil
}

// FollowingFeed fetches a page of videos from accounts the authenticated user follows.
// Requires X-Bogus; returns ErrEmptyResponse if signing is broken.
func (c *Client) FollowingFeed(ctx context.Context, count int, cursor int64) (*FeedPage, error) {
	if count <= 0 || count > 35 {
		count = 16
	}
	params := url.Values{}
	params.Set("count", strconv.Itoa(count))
	params.Set("cursor", strconv.FormatInt(cursor, 10))

	body, err := c.apiGET(ctx, "/api/follow/item_list/",
		params, baseURL+"/following")
	if err != nil {
		return nil, fmt.Errorf("FollowingFeed: %w", err)
	}

	var raw rawFeedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("FollowingFeed: %w: %v", ErrParseFailed, err)
	}

	videos, err := parseItemList(raw.ItemList)
	if err != nil {
		return nil, fmt.Errorf("FollowingFeed: %w", err)
	}

	return &FeedPage{
		Videos:  videos,
		HasMore: raw.HasMore,
		Cursor:  cursor + int64(len(videos)),
	}, nil
}

// TrendingFeed fetches a page of Explore/Trending videos.
// Requires X-Bogus.
func (c *Client) TrendingFeed(ctx context.Context, count int, cursor int64) (*FeedPage, error) {
	if count <= 0 || count > 35 {
		count = 16
	}
	params := url.Values{}
	params.Set("count", strconv.Itoa(count))
	params.Set("cursor", strconv.FormatInt(cursor, 10))

	body, err := c.apiGET(ctx, "/api/trending/feed/",
		params, baseURL+"/explore")
	if err != nil {
		return nil, fmt.Errorf("TrendingFeed: %w", err)
	}

	var raw rawFeedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("TrendingFeed: %w: %v", ErrParseFailed, err)
	}

	videos, err := parseItemList(raw.ItemList)
	if err != nil {
		return nil, fmt.Errorf("TrendingFeed: %w", err)
	}

	return &FeedPage{
		Videos:  videos,
		HasMore: raw.HasMore,
		Cursor:  cursor + int64(len(videos)),
	}, nil
}

// parseItemList decodes a slice of raw JSON video items into []Video.
func parseItemList(raws []json.RawMessage) ([]Video, error) {
	videos := make([]Video, 0, len(raws))
	for _, r := range raws {
		v, err := parseRawVideo(r)
		if err != nil {
			continue // skip unparseable items
		}
		videos = append(videos, v)
	}
	return videos, nil
}

// rawItem is the wire-format video object from TikTok's feed APIs.
type rawItem struct {
	ID            string          `json:"id"`
	Desc          string          `json:"desc"`
	CreateTime    flexInt64       `json:"createTime"` // int in API, string in SSR
	ScheduleTime  flexInt64       `json:"scheduleTime"`
	Author        Author          `json:"author"`
	AuthorStats   UserStats       `json:"authorStats"`
	AuthorStatsV2 json.RawMessage `json:"authorStatsV2"` // string-typed on the wire — ignored, use AuthorStats
	Music         Music           `json:"music"`
	Stats         rawVideoStats   `json:"stats"`
	StatsV2       VideoStatsV2    `json:"statsV2"`
	Video         VideoDetail     `json:"video"`
	Challenges    []Challenge     `json:"challenges"`
	TextExtra     []TextExtra     `json:"textExtra"`
	Contents      []Content       `json:"contents"`
	POI           *POI            `json:"poi"`

	Digged            bool   `json:"digged"`
	Collected         bool   `json:"collected"`
	ShareEnabled      bool   `json:"shareEnabled"`
	DuetEnabled       bool   `json:"duetEnabled"`
	StitchEnabled     bool   `json:"stitchEnabled"`
	DuetDisplay       int    `json:"duetDisplay"`
	StitchDisplay     int    `json:"stitchDisplay"`
	ForFriend         bool   `json:"forFriend"`
	PrivateItem       bool   `json:"privateItem"`
	IsAd              bool   `json:"isAd"`
	OfficalItem       bool   `json:"officalItem"`
	OriginalItem      bool   `json:"originalItem"`
	IsReviewing       bool   `json:"isReviewing"`
	Secret            bool   `json:"secret"`
	ItemCommentStatus int    `json:"itemCommentStatus"`
	IndexEnabled      bool   `json:"indexEnabled"`
	CategoryType      int    `json:"CategoryType"`
	DiversificationID int    `json:"diversificationId"`
	TextLanguage      string `json:"textLanguage"`
	TextTranslatable  bool   `json:"textTranslatable"`
	AIGCDescription   string `json:"AIGCDescription"`
	IsAIGC            bool   `json:"IsAigc"`
	IsHDBitrate       bool   `json:"IsHDBitrate"`
	ItemControl       ItemControl `json:"item_control"`
	BackendSourceEventTracking string `json:"backendSourceEventTracking"`
}

func parseRawVideo(raw json.RawMessage) (Video, error) {
	var item rawItem
	if err := json.Unmarshal(raw, &item); err != nil {
		return Video{}, err
	}
	return Video{
		ID:           item.ID,
		Desc:         item.Desc,
		CreateTime:   int64(item.CreateTime),
		ScheduleTime: int64(item.ScheduleTime),
		Author:       item.Author,
		AuthorStats:  item.AuthorStats,
		Music:        item.Music,
		Stats:        item.Stats.toVideoStats(),
		StatsV2:      item.StatsV2,
		Video:        item.Video,
		Challenges:   item.Challenges,
		TextExtra:    item.TextExtra,
		Contents:     item.Contents,
		POI:          item.POI,

		Digged:            item.Digged,
		Collected:         item.Collected,
		ShareEnabled:      item.ShareEnabled,
		DuetEnabled:       item.DuetEnabled,
		StitchEnabled:     item.StitchEnabled,
		DuetDisplay:       item.DuetDisplay,
		StitchDisplay:     item.StitchDisplay,
		ForFriend:         item.ForFriend,
		PrivateItem:       item.PrivateItem,
		IsAd:              item.IsAd,
		OfficalItem:       item.OfficalItem,
		OriginalItem:      item.OriginalItem,
		IsReviewing:       item.IsReviewing,
		Secret:            item.Secret,
		ItemCommentStatus: item.ItemCommentStatus,
		IndexEnabled:      item.IndexEnabled,
		CategoryType:      item.CategoryType,
		DiversificationID: item.DiversificationID,
		TextLanguage:      item.TextLanguage,
		TextTranslatable:  item.TextTranslatable,
		AIGCDescription:   item.AIGCDescription,
		IsAIGC:            item.IsAIGC,
		IsHDBitrate:       item.IsHDBitrate,
		ItemControl:       item.ItemControl,
		BackendSourceEventTracking: item.BackendSourceEventTracking,
	}, nil
}
