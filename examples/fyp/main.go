//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	tiktok "github.com/teslashibe/tiktok-go"
)

func main() {
	cookies := tiktok.Cookies{
		SessionID: os.Getenv("TT_SESSION_ID"),
		CSRFToken: os.Getenv("TT_CSRF_TOKEN"),
		MsToken:   os.Getenv("TT_MS_TOKEN"),
		TTWid:     os.Getenv("TT_TTWID"),
		OdinTT:    os.Getenv("TT_ODIN_TT"),
		SIDUcpV1:  os.Getenv("TT_SID_UCP_V1"),
		UIDtt:     os.Getenv("TT_UID_TT"),
	}

	client, err := tiktok.New(cookies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	it := tiktok.NewFYPIterator(client, 10, tiktok.WithMaxVideos(50))

	fmt.Println("Streaming For You Page...")
	for it.Next(ctx) {
		for _, v := range it.Page() {
			fmt.Printf("[%s] @%s — %s\n  plays=%-10d likes=%-10d comments=%d\n",
				v.ID, v.Author.UniqueID, truncate(v.Desc, 60),
				v.Stats.PlayCount, v.Stats.DiggCount, v.Stats.CommentCount)
		}
	}
	if err := it.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "iterator error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nTotal: %d videos\n", it.Seen())
}

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}
