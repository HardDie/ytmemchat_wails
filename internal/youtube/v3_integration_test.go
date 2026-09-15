//go:build integration

package youtube

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestLiveAPI_v3(t *testing.T) {
	key := os.Getenv("YOUTUBE_API_KEY")
	vid := os.Getenv("YOUTUBE_STREAM_ID")
	if key == "" || vid == "" {
		t.Skip("set YOUTUBE_API_KEY and YOUTUBE_STREAM_ID to run live Data API test")
	}
	c, err := New(key)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	it, err := c.GetMessageIterator(ctx, vid)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	_, _ = it.Next()
}
