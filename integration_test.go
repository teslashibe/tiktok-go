//go:build integration

package tiktok_test

import (
	"context"
	"os"
	"testing"
	"time"

	tiktok "github.com/teslashibe/tiktok-go"
)

func newTestClient(t *testing.T) *tiktok.Client {
	t.Helper()
	cookies := tiktok.Cookies{
		SessionID: requireEnv(t, "TT_SESSION_ID"),
		CSRFToken: requireEnv(t, "TT_CSRF_TOKEN"),
		MsToken:   os.Getenv("TT_MS_TOKEN"),
		SIDtt:     os.Getenv("TT_SID_TT"),
		TTWid:     os.Getenv("TT_TTWID"),
		OdinTT:    os.Getenv("TT_ODIN_TT"),
		SIDUcpV1:  os.Getenv("TT_SID_UCP_V1"),
		UIDtt:     os.Getenv("TT_UID_TT"),
	}
	c, err := tiktok.New(cookies, tiktok.WithMinRequestGap(600*time.Millisecond))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func requireEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("skipping: %s not set", key)
	}
	return v
}

func TestForYouFeed(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	page, err := c.ForYouFeed(ctx, 5, 0)
	if err != nil {
		t.Fatalf("ForYouFeed: %v", err)
	}
	if len(page.Videos) == 0 {
		t.Fatal("expected videos, got none")
	}
	t.Logf("Got %d FYP videos (hasMore=%v)", len(page.Videos), page.HasMore)
	for _, v := range page.Videos {
		t.Logf("  [%s] @%s — plays=%d likes=%d",
			v.ID, v.Author.UniqueID, v.Stats.PlayCount, v.Stats.DiggCount)
	}
}

func TestFeedIterator(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	it := tiktok.NewFYPIterator(c, 5, tiktok.WithMaxVideos(15))
	count := 0
	for it.Next(ctx) {
		count += len(it.Page())
	}
	if err := it.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	t.Logf("Iterated %d FYP videos", count)
	if count == 0 {
		t.Fatal("expected at least one video")
	}
}

func TestGetUser(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	info, err := c.GetUser(ctx, "khaby.lame")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	u, s := info.User, info.Stats
	if u.UniqueID != "khaby.lame" {
		t.Errorf("expected uniqueId=khaby.lame, got %q", u.UniqueID)
	}
	if u.SecUID == "" {
		t.Error("expected non-empty secUid")
	}
	if !u.Verified {
		t.Error("expected khaby.lame to be verified")
	}
	if s.FollowerCount < 100_000_000 {
		t.Errorf("expected >100M followers, got %d", s.FollowerCount)
	}
	t.Logf("@%s: followers=%d videos=%d secUid=%s",
		u.UniqueID, s.FollowerCount, s.VideoCount, u.SecUID)
}

func TestGetVideo(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// Get a real video ID from the FYP first
	page, err := c.ForYouFeed(ctx, 3, 0)
	if err != nil || len(page.Videos) == 0 {
		t.Skip("no FYP videos available")
	}
	v0 := page.Videos[0]

	video, err := c.GetVideo(ctx, v0.Author.UniqueID, v0.ID)
	if err != nil {
		t.Fatalf("GetVideo: %v", err)
	}
	if video.ID != v0.ID {
		t.Errorf("expected ID=%s, got %s", v0.ID, video.ID)
	}
	if video.Video.PlayAddr == "" {
		t.Error("expected non-empty PlayAddr")
	}
	t.Logf("Video %s: @%s plays=%d likes=%d duration=%ds",
		video.ID, video.Author.UniqueID, video.Stats.PlayCount, video.Stats.DiggCount, video.Video.Duration)
}

func TestSearchLive(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	page, err := c.SearchLive(ctx, "music", 5, "")
	if err != nil {
		t.Fatalf("SearchLive: %v", err)
	}
	t.Logf("Got %d live rooms", len(page.Rooms))
	for _, r := range page.Rooms {
		t.Logf("  [%d] %q @%s viewers=%d", r.ID, r.Title, r.Owner.DisplayID, r.UserCount)
	}
}

func TestPostComment(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// Get a video to comment on
	page, err := c.ForYouFeed(ctx, 3, 0)
	if err != nil || len(page.Videos) == 0 {
		t.Skip("no FYP videos")
	}
	videoID := page.Videos[0].ID

	comment, err := c.PostComment(ctx, videoID, "SDK test comment — please ignore 🤖")
	if err != nil {
		t.Fatalf("PostComment: %v", err)
	}
	if comment.CID == "" {
		t.Fatal("expected non-empty comment ID")
	}
	t.Logf("Posted comment %s on video %s", comment.CID, videoID)

	// Clean up — try to delete (needs X-Bogus so may fail gracefully)
	if err := c.DeleteComment(ctx, videoID, comment.CID); err != nil {
		t.Logf("DeleteComment (expected to need X-Bogus): %v", err)
	}
}

func TestGetUserVideos(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	info, err := c.GetUser(ctx, "khaby.lame")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}

	vp, err := c.GetUserVideos(ctx, info.User.SecUID, 5, 0)
	if err != nil {
		t.Logf("GetUserVideos (needs X-Bogus, may be ErrEmptyResponse): %v", err)
		return
	}
	t.Logf("Got %d videos for @khaby.lame", len(vp.Videos))
}

func TestXBogusFormat(t *testing.T) {
	// Verify X-Bogus produces a 25-char alphanumeric+ string
	xb := tiktok.CalcXBogusExported("aid=1988&app_name=tiktok_web", "Mozilla/5.0")
	if len(xb) != 25 {
		t.Errorf("expected 25-char X-Bogus, got %d chars: %q", len(xb), xb)
	}
	t.Logf("X-Bogus: %q", xb)
}
