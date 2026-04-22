# tiktok-go

[![Go Reference](https://pkg.go.dev/badge/github.com/teslashibe/tiktok-go.svg)](https://pkg.go.dev/github.com/teslashibe/tiktok-go)

Authenticated Go client for TikTok's private web API. Zero external dependencies — stdlib only.

Mirrors the conventions of [x-go](https://github.com/teslashibe/x-go), [reddit-go](https://github.com/teslashibe/reddit-go), and [linkedin-go](https://github.com/teslashibe/linkedin-go).

## Auth

Export browser cookies from an authenticated TikTok session:

| Cookie | Field | Required |
|--------|-------|----------|
| `sessionid` | `SessionID` | ✅ Required |
| `tt_csrf_token` | `CSRFToken` | ✅ Required |
| `msToken` | `MsToken` | Strongly recommended |
| `sid_tt` | `SIDtt` | Recommended |
| `ttwid` | `TTWid` | Recommended |
| `odin_tt` | `OdinTT` | Recommended |
| `sid_ucp_v1` | `SIDUcpV1` | Recommended |
| `uid_tt` | `UIDtt` | Recommended |

> **msToken rotation:** TikTok rotates `msToken` on every response via `Set-Cookie`. The client tracks and applies this automatically.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    tiktok "github.com/teslashibe/tiktok-go"
)

func main() {
    client, _ := tiktok.New(tiktok.Cookies{
        SessionID: "your-sessionid",
        CSRFToken: "your-csrf-token",
        MsToken:   "your-ms-token",
    })

    // Stream FYP feed
    it := tiktok.NewFYPIterator(client, 10, tiktok.WithMaxVideos(50))
    for it.Next(context.Background()) {
        for _, v := range it.Page() {
            fmt.Printf("@%s — %s (plays: %d)\n",
                v.Author.UniqueID, v.Desc, v.Stats.PlayCount)
        }
    }
}
```

## Integration Test Results

All Tier 1 tests pass with a valid session:

```
--- PASS: TestForYouFeed    (0.70s)   Got 5 FYP videos (hasMore=true)
--- PASS: TestFeedIterator  (1.59s)   Iterated 15 FYP videos
--- PASS: TestGetUser       (0.55s)   @khaby.lame: followers=160800000 videos=1317
--- PASS: TestGetVideo      (1.35s)   Video 7628338924133436692: plays=16800000 duration=62s
--- PASS: TestSearchLive    (0.69s)   (0 active rooms at time of test)
--- PASS: TestPostComment   (1.34s)   Posted comment on video
--- PASS: TestGetUserVideos (0.76s)   ErrEmptyResponse (expected — needs X-Bogus)
--- PASS: TestXBogusFormat  (0.00s)   X-Bogus: 25-char string verified
```

## API Surface

### Tier 1 — Works Without X-Bogus ✅

These endpoints work with cookies alone.

| Method | Description | Function |
|--------|-------------|----------|
| FYP feed | For You Page videos | `ForYouFeed(ctx, count, cursor)` |
| User profile | Full user + stats | `GetUser(ctx, username)` |
| Video detail | Full video struct | `GetVideo(ctx, username, videoID)` |
| Live search | Search live rooms | `SearchLive(ctx, keyword, count, cursor)` |
| Post comment | Publish a comment | `PostComment(ctx, videoID, text)` |
| Reply to comment | Reply to a comment | `ReplyToComment(ctx, videoID, commentID, text)` |

### Tier 2 — Requires X-Bogus ⚠️

| Method | Description | Function |
|--------|-------------|----------|
| User videos | Paginated user posts | `GetUserVideos(ctx, secUID, count, cursor)` |
| Liked videos | User's liked videos | `GetLikedVideos(ctx, secUID, count, cursor)` |
| Saved videos | Authenticated user's saves | `GetSavedVideos(ctx, count, cursor)` |
| Followers | User follower list | `GetFollowers(ctx, secUID, count, minCursor)` |
| Following | User following list | `GetFollowing(ctx, secUID, count, minCursor)` |
| Search videos | Video search | `SearchVideos(ctx, keyword, count, cursor)` |
| Search users | User search | `SearchUsers(ctx, keyword, count, cursor)` |
| Get comments | Video comment list | `GetComments(ctx, videoID, count, cursor)` |
| Get replies | Comment reply list | `GetReplies(ctx, videoID, commentID, count, cursor)` |
| Get hashtag | Hashtag detail + stats | `GetHashtag(ctx, name)` |
| Hashtag videos | Videos under hashtag | `GetHashtagVideos(ctx, challengeID, count, cursor)` |
| Following feed | Videos from followed accounts | `FollowingFeed(ctx, count, cursor)` |
| Trending feed | Explore/trending videos | `TrendingFeed(ctx, count, cursor)` |
| Delete comment | Delete own comment | `DeleteComment(ctx, videoID, commentID)` |
| Like comment | Like/unlike a comment | `LikeComment(ctx, videoID, commentID, like)` |

### Social Actions (all require X-Bogus) ⚠️

| Function | Description |
|----------|-------------|
| `LikeVideo(ctx, videoID, like)` | Like or unlike a video |
| `FollowUser(ctx, userID, follow)` | Follow or unfollow a user |
| `CollectVideo(ctx, videoID, collect)` | Save or unsave a video |
| `RepostVideo(ctx, videoID)` | Repost to own profile |
| `DeleteRepost(ctx, videoID)` | Remove a repost |
| `BlockUser(ctx, userID, block)` | Block or unblock a user |
| `MuteUser(ctx, userID, mute)` | Mute or unmute a user |

## Iterators

```go
// FYP iterator — streams indefinitely (FYP never ends)
it := tiktok.NewFYPIterator(client, 10, tiktok.WithMaxVideos(100))
for it.Next(ctx) {
    for _, v := range it.Page() { /* ... */ }
}

// Resume from checkpoint
cp := it.Checkpoint()
data, _ := cp.Marshal()
// ... later ...
cp2, _ := tiktok.UnmarshalCheckpoint(data)
it2 := tiktok.NewFYPIterator(client, 10, tiktok.WithFeedCheckpoint(cp2))

// User videos iterator
it := tiktok.NewUserVideosIterator(client, user.SecUID, 20)

// Comment iterator
cit := tiktok.NewCommentIterator(client, videoID)
for cit.Next(ctx) {
    for _, c := range cit.Page() { /* ... */ }
}
```

## X-Bogus

TikTok requires an `X-Bogus` query parameter on most `/api/*` endpoints. Without it, the server returns HTTP 200 with an empty body.

The algorithm is implemented in `xbogus.go`:
1. CRC32(IEEE) of the URL query string
2. CRC32(IEEE) of the User-Agent
3. Current Unix timestamp (seconds, lower 24 bits)
4. 19-byte array with magic header `[0x02, 0x01, 0x01, 0x01, 0xB8, 0x01, 0x40]`
5. XOR checksum byte
6. Base64-like encoding with `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/`
7. First 25 characters of result

## Error Handling

| Error | Meaning |
|-------|---------|
| `ErrInvalidAuth` | Missing SessionID or CSRFToken |
| `ErrUnauthorized` | Session expired or invalid |
| `ErrForbidden` | Private account or restricted content |
| `ErrNotFound` | User or video does not exist |
| `ErrRateLimited` | Too many requests |
| `ErrEmptyResponse` | Empty body — X-Bogus signature likely wrong |
| `ErrParseFailed` | Unexpected response format |

## Wire-Format Notes

TikTok's JSON encoding is inconsistent between the REST API and SSR page data:

| Field | API response | SSR page data |
|-------|-------------|---------------|
| `createTime` | `int64` | `"string"` |
| `video.size` | `int64` | `"string"` |
| `music.preciseDuration` | `float64` | `{"preciseDuration": 5.9, ...}` |
| `authorStatsV2.*` | `string` (skip) | `string` (skip) |
| `comment.user.avatar_*` | `string` | `{"url_list": [...]}` |
| `has_more` (search) | `int` 0/1 | `int` 0/1 |
| `cursor` (live search) | `int64` | `int64` |

The SDK handles all these transparently via `flexInt64`, `flexFloat64`, and `commentAvatar` custom unmarshalers.

## Rate Limiting

TikTok does not expose rate limit headers. Recommended defaults:
- **Min request gap**: 500ms (configurable via `WithMinRequestGap`)
- **msToken**: Must be kept fresh (auto-updated by the client)
- Aggressive polling triggers captcha/bot detection

```go
client, _ := tiktok.New(cookies,
    tiktok.WithMinRequestGap(time.Second),
)
```

## Integration Tests

```bash
TT_SESSION_ID="..." TT_CSRF_TOKEN="..." TT_MS_TOKEN="..." \
  go test -v -tags=integration ./...
```

## Package Structure

```
tiktok-go/
├── tiktok.go          Client, Cookies, New(), options
├── client.go          HTTP transport, apiGET/POST, msToken rotation
├── xbogus.go          X-Bogus signing algorithm
├── types.go           User, Video, Music, LiveRoom, Comment, Challenge, POI...
├── sigi.go            SSR page parser (__UNIVERSAL_DATA_FOR_REHYDRATION__)
├── feed.go            ForYouFeed, FollowingFeed, TrendingFeed
├── profile.go         GetUser, GetUserVideos, GetFollowers, GetFollowing...
├── video.go           GetVideo (HTML scrape)
├── search.go          SearchLive, SearchVideos, SearchUsers
├── hashtag.go         GetHashtag, GetHashtagVideos
├── comments.go        GetComments, GetReplies, PostComment, ReplyToComment...
├── social.go          LikeVideo, FollowUser, CollectVideo, BlockUser...
├── iterator.go        FeedIterator, CommentIterator, Checkpoint
├── errors.go          Sentinel errors
└── examples/
    ├── fyp/           Stream For You Page
    ├── get_user/      Fetch user profile
    ├── get_video/     Fetch video detail
    ├── post_comment/  Post + reply to comments
    └── search_live/   Search live rooms
```
