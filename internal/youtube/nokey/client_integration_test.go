//go:build integration

package nokey

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestLiveHTMLChat(t *testing.T) {
	vid := os.Getenv("YOUTUBE_STREAM_ID")
	if vid == "" {
		t.Skip("set YOUTUBE_STREAM_ID to run live no-key chat test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	it, err := New().GetMessageIterator(ctx, vid)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	_, _ = it.Next()
}
