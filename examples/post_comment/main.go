//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	tiktok "github.com/teslashibe/tiktok-go"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: post_comment <videoID> <text>")
		os.Exit(1)
	}
	videoID := os.Args[1]
	text := os.Args[2]

	cookies := tiktok.Cookies{
		SessionID: os.Getenv("TT_SESSION_ID"),
		CSRFToken: os.Getenv("TT_CSRF_TOKEN"),
		MsToken:   os.Getenv("TT_MS_TOKEN"),
		SIDtt:     os.Getenv("TT_SID_TT"),
	}

	client, err := tiktok.New(cookies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	comment, err := client.PostComment(context.Background(), videoID, text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "PostComment error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Comment posted!\n")
	fmt.Printf("  ID:     %s\n", comment.CID)
	fmt.Printf("  Text:   %s\n", comment.Text)
	fmt.Printf("  Status: %d\n", comment.Status)
	fmt.Printf("  Author: @%s\n", comment.User.UniqueID)
}
