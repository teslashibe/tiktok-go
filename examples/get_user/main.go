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
		fmt.Fprintln(os.Stderr, "usage: get_user <username>")
		os.Exit(1)
	}
	username := os.Args[1]

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

	info, err := client.GetUser(context.Background(), username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetUser error: %v\n", err)
		os.Exit(1)
	}

	u, s := info.User, info.Stats
	fmt.Printf("@%s (%s)\n", u.UniqueID, u.Nickname)
	fmt.Printf("  ID:        %s\n", u.ID)
	fmt.Printf("  SecUID:    %s\n", u.SecUID)
	fmt.Printf("  Verified:  %v\n", u.Verified)
	fmt.Printf("  Private:   %v\n", u.PrivateAccount)
	fmt.Printf("  Bio:       %s\n", u.Signature)
	fmt.Printf("  Followers: %d\n", s.FollowerCount)
	fmt.Printf("  Following: %d\n", s.FollowingCount)
	fmt.Printf("  Likes:     %d\n", s.Heart)
	fmt.Printf("  Videos:    %d\n", s.VideoCount)
	if u.RoomID != "" {
		fmt.Printf("  🔴 LIVE    roomId=%s\n", u.RoomID)
	}
}
