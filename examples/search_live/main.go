//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	tiktok "github.com/teslashibe/tiktok-go"
)

func main() {
	keyword := "music"
	if len(os.Args) > 1 {
		keyword = os.Args[1]
	}

	cookies := tiktok.Cookies{
		SessionID: os.Getenv("TT_SESSION_ID"),
		CSRFToken: os.Getenv("TT_CSRF_TOKEN"),
		MsToken:   os.Getenv("TT_MS_TOKEN"),
	}

	client, err := tiktok.New(cookies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	page, err := client.SearchLive(context.Background(), keyword, 10, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "SearchLive error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Live rooms for %q (%d results, hasMore=%v):\n\n", keyword, len(page.Rooms), page.HasMore)
	for _, r := range page.Rooms {
		fmt.Printf("  [%d] %s\n", r.ID, r.Title)
		fmt.Printf("    Host:    @%s (%s)\n", r.Owner.DisplayID, r.Owner.Nickname)
		fmt.Printf("    Viewers: %d  Likes: %d\n", r.UserCount, r.LikeCount)
		if r.Hashtag.Title != "" {
			fmt.Printf("    Tag:     #%s\n", r.Hashtag.Title)
		}
		fmt.Println()
	}
}
