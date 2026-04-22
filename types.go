package tiktok

import (
	"encoding/json"
	"strconv"
)

// flexInt64 unmarshals from both a JSON number (1234) and a JSON string ("1234").
// TikTok's SSR page data encodes timestamps as strings.
type flexInt64 int64

func (f *flexInt64) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*f = flexInt64(n)
		return nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*f = flexInt64(n)
	return nil
}

// commentAvatar unmarshals from either a bare string URL or {"url_list":[...],"url":"..."}.
// TikTok's comment endpoints return avatar fields as nested objects.
type commentAvatar string

func (a *commentAvatar) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*a = commentAvatar(s)
		return nil
	}
	var obj struct {
		URLList []string `json:"url_list"`
		URL     string   `json:"url"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	if len(obj.URLList) > 0 {
		*a = commentAvatar(obj.URLList[0])
	} else {
		*a = commentAvatar(obj.URL)
	}
	return nil
}

// String returns the URL string value.
func (a commentAvatar) String() string { return string(a) }

// flexFloat64 unmarshals from either a JSON number (5.9) or an object
// {"preciseDuration": 5.9, ...} as TikTok's SSR page data uses for music duration.
type flexFloat64 float64

func (f *flexFloat64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	first := data[0]
	if first == '-' || (first >= '0' && first <= '9') {
		var n float64
		if err := json.Unmarshal(data, &n); err != nil {
			return err
		}
		*f = flexFloat64(n)
		return nil
	}
	if first == '{' {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		for _, key := range []string{"preciseDuration", "value"} {
			if raw, ok := obj[key]; ok {
				var n float64
				if err := json.Unmarshal(raw, &n); err == nil {
					*f = flexFloat64(n)
					return nil
				}
			}
		}
		return nil // leave at zero if no recognisable key
	}
	var n float64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*f = flexFloat64(n)
	return nil
}

// ─── User ────────────────────────────────────────────────────────────────────

// User is a TikTok user profile.
type User struct {
	ID                   string `json:"id"`
	ShortID              string `json:"shortId"`
	UniqueID             string `json:"uniqueId"`
	Nickname             string `json:"nickname"`
	Signature            string `json:"signature"`
	AvatarLarger         string `json:"avatarLarger"`
	AvatarMedium         string `json:"avatarMedium"`
	AvatarThumb          string `json:"avatarThumb"`
	CreateTime           int64  `json:"createTime"`
	Verified             bool   `json:"verified"`
	SecUID               string `json:"secUid"`
	PrivateAccount       bool   `json:"privateAccount"`
	Secret               bool   `json:"secret"`
	FTC                  bool   `json:"ftc"`
	Relation             int    `json:"relation"` // 0=none,1=following,2=follower,3=mutual
	OpenFavorite         bool   `json:"openFavorite"`
	CommentSetting       int    `json:"commentSetting"`
	DuetSetting          int    `json:"duetSetting"`
	StitchSetting        int    `json:"stitchSetting"`
	DownloadSetting      int    `json:"downloadSetting"`
	IsADVirtual          bool   `json:"isADVirtual"`
	IsEmbedBanned        bool   `json:"isEmbedBanned"`
	CanExpPlaylist       bool   `json:"canExpPlaylist"`
	TtSeller             bool   `json:"ttSeller"`
	RoomID               string `json:"roomId"` // non-empty if currently live
	UniqueIDModifyTime   int64  `json:"uniqueIdModifyTime"`
	NickNameModifyTime   int64  `json:"nickNameModifyTime"`
	RecommendReason      string `json:"recommendReason"`
	NowInvitationCardURL string `json:"nowInvitationCardUrl"`
	SuggestAccountBind   bool   `json:"suggestAccountBind"`
	UserStoryStatus      int    `json:"UserStoryStatus"`
}

// UserStats holds engagement metrics for a user.
type UserStats struct {
	FollowerCount  int64 `json:"followerCount"`
	FollowingCount int64 `json:"followingCount"`
	Heart          int64 `json:"heart"`      // total likes received
	HeartCount     int64 `json:"heartCount"`
	VideoCount     int64 `json:"videoCount"`
	DiggCount      int64 `json:"diggCount"`
	FriendCount    int64 `json:"friendCount"`
}

// UserInfo bundles a User with their stats (returned from profile pages).
type UserInfo struct {
	User  User      `json:"user"`
	Stats UserStats `json:"stats"`
}

// UserPage is a paginated list of users.
type UserPage struct {
	Users     []UserInfo `json:"users"`
	HasMore   bool       `json:"hasMore"`
	MinCursor string     `json:"minCursor"`
	MaxCursor string     `json:"maxCursor"`
}

// ─── Video ───────────────────────────────────────────────────────────────────

// Video is a TikTok video (also called an "item" or "aweme").
type Video struct {
	ID           string       `json:"id"`
	Desc         string       `json:"desc"`
	CreateTime   int64        `json:"createTime"`
	ScheduleTime int64        `json:"scheduleTime"`
	Author       Author       `json:"author"`
	AuthorStats  UserStats    `json:"authorStats"`
	Music        Music        `json:"music"`
	Stats        VideoStats   `json:"stats"`
	StatsV2      VideoStatsV2 `json:"statsV2"`
	Video        VideoDetail  `json:"video"`
	Challenges   []Challenge  `json:"challenges"`
	TextExtra    []TextExtra  `json:"textExtra"`
	Contents     []Content    `json:"contents"`
	POI          *POI         `json:"poi,omitempty"`

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

	AIGCDescription string `json:"AIGCDescription"`
	IsAIGC          bool   `json:"IsAigc"`
	IsHDBitrate     bool   `json:"IsHDBitrate"`

	ItemControl ItemControl `json:"item_control"`

	BackendSourceEventTracking string `json:"backendSourceEventTracking"`
}

// Author is the compact user embedded on video objects.
type Author struct {
	ID              string `json:"id"`
	UniqueID        string `json:"uniqueId"`
	Nickname        string `json:"nickname"`
	Signature       string `json:"signature"`
	SecUID          string `json:"secUid"`
	AvatarLarger    string `json:"avatarLarger"`
	AvatarMedium    string `json:"avatarMedium"`
	AvatarThumb     string `json:"avatarThumb"`
	Verified        bool   `json:"verified"`
	PrivateAccount  bool   `json:"privateAccount"`
	Relation        int    `json:"relation"`
	OpenFavorite    bool   `json:"openFavorite"`
	CommentSetting  int    `json:"commentSetting"`
	DuetSetting     int    `json:"duetSetting"`
	StitchSetting   int    `json:"stitchSetting"`
	DownloadSetting int    `json:"downloadSetting"`
	FTC             bool   `json:"ftc"`
	IsADVirtual     bool   `json:"isADVirtual"`
	IsEmbedBanned   bool   `json:"isEmbedBanned"`
	RoomID          string `json:"roomId"`
	UserStoryStatus int    `json:"UserStoryStatus"`
}

// VideoStats holds engagement counts for a video.
type VideoStats struct {
	DiggCount    int64 `json:"diggCount"`
	ShareCount   int64 `json:"shareCount"`
	CommentCount int64 `json:"commentCount"`
	PlayCount    int64 `json:"playCount"`
	CollectCount int64 `json:"-"` // filled from string statsV2 or raw
}

// rawVideoStats is used for JSON unmarshalling where collectCount is a string.
type rawVideoStats struct {
	DiggCount    int64       `json:"diggCount"`
	ShareCount   int64       `json:"shareCount"`
	CommentCount int64       `json:"commentCount"`
	PlayCount    int64       `json:"playCount"`
	CollectCount interface{} `json:"collectCount"` // may be int or string
}

func (r rawVideoStats) toVideoStats() VideoStats {
	vs := VideoStats{
		DiggCount:    r.DiggCount,
		ShareCount:   r.ShareCount,
		CommentCount: r.CommentCount,
		PlayCount:    r.PlayCount,
	}
	switch v := r.CollectCount.(type) {
	case float64:
		vs.CollectCount = int64(v)
	case string:
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			vs.CollectCount = n
		}
	}
	return vs
}

// VideoStatsV2 is the string-typed version of VideoStats (includes repostCount).
type VideoStatsV2 struct {
	DiggCount    string `json:"diggCount"`
	ShareCount   string `json:"shareCount"`
	CommentCount string `json:"commentCount"`
	PlayCount    string `json:"playCount"`
	RepostCount  string `json:"repostCount"`
	CollectCount string `json:"collectCount"`
}

// VideoDetail holds the media-level attributes of a video.
type VideoDetail struct {
	ID           string            `json:"id"`
	Height       int               `json:"height"`
	Width        int               `json:"width"`
	Duration     int               `json:"duration"`
	Ratio        string            `json:"ratio"`
	Cover        string            `json:"cover"`
	OriginCover  string            `json:"originCover"`
	DynamicCover string            `json:"dynamicCover"`
	PlayAddr     string            `json:"playAddr"`
	DownloadAddr string            `json:"downloadAddr"`
	ReflowCover  string            `json:"reflowCover"`
	Bitrate      int               `json:"bitrate"`
	EncodedType  string            `json:"encodedType"`
	Format       string            `json:"format"`
	VideoQuality string            `json:"videoQuality"`
	CodecType    string            `json:"codecType"`
	Definition   string            `json:"definition"`
	Size         flexInt64         `json:"size"` // int in API response, string in SSR page data
	VideoID      string            `json:"videoID"`
	ZoomCover    map[string]string `json:"zoomCover"`
}

// VideoPage is a paginated list of videos.
type VideoPage struct {
	Videos  []Video `json:"videos"`
	HasMore bool    `json:"hasMore"`
	Cursor  int64   `json:"cursor"`
}

// ─── Music / Sound ───────────────────────────────────────────────────────────

// Music represents the sound attached to a video.
type Music struct {
	ID               string      `json:"id"`
	Title            string      `json:"title"`
	AuthorName       string      `json:"authorName"`
	PlayURL          string      `json:"playUrl"`
	CoverLarge       string      `json:"coverLarge"`
	CoverMedium      string      `json:"coverMedium"`
	CoverThumb       string      `json:"coverThumb"`
	Duration         int         `json:"duration"`
	Original         bool        `json:"original"`
	Private          bool        `json:"private"`
	IsCopyrighted    bool        `json:"isCopyrighted"`
	IsCommerceMusic  bool        `json:"is_commerce_music"`
	IsUnlimitedMusic bool        `json:"is_unlimited_music"`
	ShootDuration    int         `json:"shoot_duration"`
	PreciseDuration  flexFloat64 `json:"preciseDuration"` // float in API, nested object in SSR
	Collected        bool        `json:"collected"`
}

// ─── Hashtag / Challenge ─────────────────────────────────────────────────────

// Challenge is a TikTok hashtag/challenge.
type Challenge struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Desc          string `json:"desc"`
	ProfileLarger string `json:"profileLarger"`
	ProfileMedium string `json:"profileMedium"`
	ProfileThumb  string `json:"profileThumb"`
	CoverLarger   string `json:"coverLarger"`
	CoverMedium   string `json:"coverMedium"`
	CoverThumb    string `json:"coverThumb"`
	IsCommerce    bool   `json:"isCommerce"`
}

// ChallengeInfo bundles a Challenge with its video/view statistics.
type ChallengeInfo struct {
	Challenge Challenge      `json:"challenge"`
	Stats     ChallengeStats `json:"stats"`
}

// ChallengeStats holds engagement stats for a hashtag.
type ChallengeStats struct {
	VideoCount int64 `json:"videoCount"`
	ViewCount  int64 `json:"viewCount"`
}

// ─── Location / POI ──────────────────────────────────────────────────────────

// POI is a TikTok location tag.
type POI struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Address          string `json:"address"`
	City             string `json:"city"`
	Province         string `json:"province"`
	Country          string `json:"country"`
	CityCode         string `json:"cityCode"`
	CountryCode      string `json:"countryCode"`
	FatherPOIID      string `json:"fatherPoiId"`
	FatherPOIName    string `json:"fatherPoiName"`
	TTTypeCode       string `json:"ttTypeCode"`
	TTTypeNameTiny   string `json:"ttTypeNameTiny"`
	TTTypeNameMedium string `json:"ttTypeNameMedium"`
	TTTypeNameSuper  string `json:"ttTypeNameSuper"`
	Type             int    `json:"type"`
	TypeCode         string `json:"typeCode"`
}

// ─── Text Entities ───────────────────────────────────────────────────────────

// TextExtra is an inline entity in a video description.
type TextExtra struct {
	Start        int    `json:"start"`
	End          int    `json:"end"`
	Type         int    `json:"type"` // 1=hashtag, 2=mention, 3=URL
	HashtagName  string `json:"hashtagName"`
	HashtagID    string `json:"hashtagId"`
	UserID       string `json:"userId"`
	UserUniqueID string `json:"userUniqueId"`
	SecUID       string `json:"secUid"`
}

// Content is a parsed segment of a video description.
type Content struct {
	Desc string `json:"desc"`
}

// ItemControl holds flags about what actions are permitted on a video.
type ItemControl struct {
	CanRepost bool `json:"can_repost"`
}

// ─── Feed ────────────────────────────────────────────────────────────────────

// FeedPage is a page of videos from a feed (FYP, Following, Trending).
type FeedPage struct {
	Videos  []Video `json:"videos"`
	HasMore bool    `json:"hasMore"`
	Cursor  int64   `json:"cursor"`
}

// ─── Comments ────────────────────────────────────────────────────────────────

// Comment is a TikTok video comment.
type Comment struct {
	AwemeID            string      `json:"aweme_id"`
	CID                string      `json:"cid"`
	CreateTime         int64       `json:"create_time"`
	DiggCount          int         `json:"digg_count"`
	Status             int         `json:"status"` // 2=published, 4=under review
	Text               string      `json:"text"`
	TextExtra          []TextExtra `json:"text_extra"`
	ReplyID            string      `json:"reply_id"`
	ReplyToReplyID     string      `json:"reply_to_reply_id"`
	ReplyComment       []Comment   `json:"reply_comment"`
	UserDigged         int         `json:"user_digged"`
	CommentPostItemIDs []string    `json:"comment_post_item_ids"`
	User               CommentUser `json:"user"`
}

// CommentUser is the minimal user object embedded in comment responses.
type CommentUser struct {
	UID           string        `json:"uid"`
	UniqueID      string        `json:"unique_id"`
	Nickname      string        `json:"nickname"`
	Signature     string        `json:"signature"`
	SecUID        string        `json:"sec_uid"`
	Region        string        `json:"region"`
	AvatarLarger  commentAvatar `json:"avatar_larger"`
	AvatarMedium  commentAvatar `json:"avatar_medium"`
	AvatarThumb   commentAvatar `json:"avatar_thumb"`
	Verified      bool          `json:"is_verified"`
	FollowerCount int64         `json:"follower_count"`
	FollowStatus  int           `json:"follow_status"`
}

// CommentPage is a paginated list of comments.
type CommentPage struct {
	Comments []Comment `json:"comments"`
	HasMore  bool      `json:"has_more"`
	Cursor   int64     `json:"cursor"`
	Total    int64     `json:"total"`
}

// ─── Live ─────────────────────────────────────────────────────────────────────

// LiveRoom represents an active TikTok live stream.
type LiveRoom struct {
	ID            int64         `json:"id"`
	IDStr         string        `json:"id_str"`
	Title         string        `json:"title"`
	Status        int           `json:"status"`     // 2 = currently live
	UserCount     int           `json:"user_count"` // concurrent viewers
	LikeCount     int64         `json:"like_count"`
	StartTime     int64         `json:"start_time"`
	Cover         string        // first URL from cover.url_list
	Owner         LiveOwner     `json:"owner"`
	Stats         LiveStats     `json:"stats"`
	Hashtag       LiveHashtag   `json:"hashtag"`
	StreamURL     LiveStreamURL `json:"stream_url"`
	LiveRoomMode  int           `json:"live_room_mode"`
	AgeRestricted bool          `json:"age_restricted"`
}

// LiveOwner is the host of a live room.
type LiveOwner struct {
	ID             string `json:"id"`
	IDStr          string `json:"id_str"`
	Nickname       string `json:"nickname"`
	DisplayID      string `json:"display_id"`
	SecUID         string `json:"sec_uid"`
	BioDescription string `json:"bio_description"`
	AvatarThumb    string // from avatar_thumb.url_list[0]
	AvatarMedium   string
	AvatarLarge    string
}

// LiveStats holds engagement metrics for a live stream.
type LiveStats struct {
	TotalUser    int `json:"total_user"`
	EnterCount   int `json:"enter_count"`
	ShareCount   int `json:"share_count"`
	CommentCount int `json:"comment_count"`
}

// LiveHashtag is the category/hashtag associated with a live stream.
type LiveHashtag struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// LiveStreamURL contains playback URLs for a live stream.
type LiveStreamURL struct {
	RTMPPullURL         string            `json:"rtmp_pull_url"`
	FLVPullURL          map[string]string `json:"flv_pull_url"`
	CandidateResolution []string          `json:"candidate_resolution"`
	StreamSizeWidth     int               `json:"stream_size_width"`
	StreamSizeHeight    int               `json:"stream_size_height"`
}

// LiveSearchPage is a page of live room results.
type LiveSearchPage struct {
	Rooms   []LiveRoom `json:"rooms"`
	HasMore bool       `json:"has_more"`
	Cursor  string     `json:"cursor"`
}

// ─── Search ──────────────────────────────────────────────────────────────────

// SearchPage holds mixed results from a general search.
type SearchPage struct {
	Videos  []Video    `json:"videos"`
	Users   []UserInfo `json:"users"`
	HasMore bool       `json:"has_more"`
	Cursor  string     `json:"cursor"`
}
