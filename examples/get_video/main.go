//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	tiktok "github.com/teslashibe/tiktok-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: get_video <videoID>")
		os.Exit(1)
	}
	videoID := os.Args[1]

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

	v, err := client.GetVideo(context.Background(), "", videoID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetVideo error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Video %s\n", v.ID)
	fmt.Printf("  Author:   @%s\n", v.Author.UniqueID)
	fmt.Printf("  Desc:     %s\n", v.Desc)
	fmt.Printf("  Duration: %ds  %dx%d\n", v.Video.Duration, v.Video.Width, v.Video.Height)
	fmt.Printf("  Sound:    %s — %s\n", v.Music.Title, v.Music.AuthorName)
	fmt.Printf("  Plays:    %d\n", v.Stats.PlayCount)
	fmt.Printf("  Likes:    %d\n", v.Stats.DiggCount)
	fmt.Printf("  Comments: %d\n", v.Stats.CommentCount)
	fmt.Printf("  Shares:   %d\n", v.Stats.ShareCount)
	fmt.Printf("  Saves:    %d\n", v.Stats.CollectCount)
	fmt.Printf("  PlayURL:  %s\n", v.Video.PlayAddr)
	if v.POI != nil {
		fmt.Printf("  Location: %s, %s\n", v.POI.Name, v.POI.Address)
	}
	if len(v.Challenges) > 0 {
		for _, ch := range v.Challenges {
			fmt.Printf("  #%s\n", ch.Title)
		}
	}
}
