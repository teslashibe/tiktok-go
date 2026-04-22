package tiktok

import (
	"context"
	"encoding/json"
	"time"
)

// Checkpoint is a serialisable cursor for resuming paginated iteration.
type Checkpoint struct {
	Cursor    int64     `json:"cursor"`
	CursorStr string    `json:"cursor_str,omitempty"` // for string-cursor endpoints
	Seen      int       `json:"seen"`
	Query     string    `json:"query,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Marshal serialises a Checkpoint to JSON.
func (cp Checkpoint) Marshal() ([]byte, error) { return json.Marshal(cp) }

// UnmarshalCheckpoint deserialises a Checkpoint from JSON.
func UnmarshalCheckpoint(data []byte) (Checkpoint, error) {
	var cp Checkpoint
	return cp, json.Unmarshal(data, &cp)
}

// ─── FeedIterator ─────────────────────────────────────────────────────────────

// FeedIterator walks through paginated FYP/Following/Trending feed pages.
type FeedIterator struct {
	client    *Client
	fetchFn   func(ctx context.Context, cursor int64) (*FeedPage, error)
	cursor    int64
	seen      int
	maxVideos int
	done      bool
	page      []Video
	err       error
	query     string
}

// FeedIteratorOption configures a FeedIterator.
type FeedIteratorOption func(*FeedIterator)

// WithMaxVideos caps the total videos returned across all pages.
func WithMaxVideos(n int) FeedIteratorOption {
	return func(it *FeedIterator) { it.maxVideos = n }
}

// WithFeedCheckpoint resumes from a saved checkpoint.
func WithFeedCheckpoint(cp Checkpoint) FeedIteratorOption {
	return func(it *FeedIterator) {
		it.cursor = cp.Cursor
		it.seen = cp.Seen
	}
}

// NewFYPIterator creates an iterator over the For You Page feed.
func NewFYPIterator(c *Client, pageSize int, opts ...FeedIteratorOption) *FeedIterator {
	it := &FeedIterator{client: c, query: "fyp"}
	it.fetchFn = func(ctx context.Context, cursor int64) (*FeedPage, error) {
		return c.ForYouFeed(ctx, pageSize, cursor)
	}
	for _, o := range opts {
		o(it)
	}
	return it
}

// NewFollowingIterator creates an iterator over the Following feed.
func NewFollowingIterator(c *Client, pageSize int, opts ...FeedIteratorOption) *FeedIterator {
	it := &FeedIterator{client: c, query: "following"}
	it.fetchFn = func(ctx context.Context, cursor int64) (*FeedPage, error) {
		return c.FollowingFeed(ctx, pageSize, cursor)
	}
	for _, o := range opts {
		o(it)
	}
	return it
}

// NewUserVideosIterator creates an iterator over a user's posted videos.
func NewUserVideosIterator(c *Client, secUID string, pageSize int, opts ...FeedIteratorOption) *FeedIterator {
	it := &FeedIterator{client: c, query: "user:" + secUID}
	it.fetchFn = func(ctx context.Context, cursor int64) (*FeedPage, error) {
		vp, err := c.GetUserVideos(ctx, secUID, pageSize, cursor)
		if err != nil {
			return nil, err
		}
		return &FeedPage{Videos: vp.Videos, HasMore: vp.HasMore, Cursor: vp.Cursor}, nil
	}
	for _, o := range opts {
		o(it)
	}
	return it
}

// Next fetches the next page. Returns false when exhausted.
func (it *FeedIterator) Next(ctx context.Context) bool {
	if it.done {
		return false
	}
	if ctx.Err() != nil {
		it.err = ctx.Err()
		it.done = true
		return false
	}
	if it.maxVideos > 0 && it.seen >= it.maxVideos {
		it.done = true
		return false
	}

	page, err := it.fetchFn(ctx, it.cursor)
	if err != nil {
		it.err = err
		it.done = true
		return false
	}

	if len(page.Videos) == 0 {
		it.done = true
		return false
	}

	videos := page.Videos
	if it.maxVideos > 0 {
		rem := it.maxVideos - it.seen
		if rem < len(videos) {
			videos = videos[:rem]
			it.done = true
		}
	}

	it.page = videos
	it.seen += len(videos)
	it.cursor = page.Cursor
	if !page.HasMore {
		it.done = true
	}

	return len(it.page) > 0
}

// Page returns the videos from the most recent Next() call.
func (it *FeedIterator) Page() []Video { return it.page }

// Err returns the first error encountered, if any.
func (it *FeedIterator) Err() error { return it.err }

// Seen returns the total number of videos returned so far.
func (it *FeedIterator) Seen() int { return it.seen }

// Checkpoint returns a serialisable position for resuming later.
func (it *FeedIterator) Checkpoint() Checkpoint {
	return Checkpoint{
		Cursor:    it.cursor,
		Seen:      it.seen,
		Query:     it.query,
		CreatedAt: time.Now(),
	}
}

// ─── CommentIterator ──────────────────────────────────────────────────────────

// CommentIterator walks through paginated comment pages on a video.
type CommentIterator struct {
	client  *Client
	videoID string
	cursor  int64
	seen    int
	done    bool
	page    []Comment
	err     error
}

// NewCommentIterator creates an iterator over a video's comments.
func NewCommentIterator(c *Client, videoID string) *CommentIterator {
	return &CommentIterator{client: c, videoID: videoID}
}

// Next fetches the next page of comments. Returns false when exhausted.
func (it *CommentIterator) Next(ctx context.Context) bool {
	if it.done {
		return false
	}
	page, err := it.client.GetComments(ctx, it.videoID, 20, it.cursor)
	if err != nil {
		it.err = err
		it.done = true
		return false
	}
	if len(page.Comments) == 0 {
		it.done = true
		return false
	}
	it.page = page.Comments
	it.seen += len(page.Comments)
	it.cursor = page.Cursor + int64(len(page.Comments))
	if !page.HasMore {
		it.done = true
	}
	return true
}

// Page returns comments from the most recent Next() call.
func (it *CommentIterator) Page() []Comment { return it.page }

// Err returns the first error, if any.
func (it *CommentIterator) Err() error { return it.err }

// Seen returns total comments returned so far.
func (it *CommentIterator) Seen() int { return it.seen }

// Checkpoint returns a cursor for resuming iteration.
func (it *CommentIterator) Checkpoint() Checkpoint {
	return Checkpoint{Cursor: it.cursor, Seen: it.seen, Query: "comments:" + it.videoID, CreatedAt: time.Now()}
}
