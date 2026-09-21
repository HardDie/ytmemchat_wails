package home

import (
	"context"
	"strings"
	"testing"

	"github.com/HardDie/ytmemchat_wails/internal/youtube"
)

func TestLookupLatestStream_usesSavedKeyAndFormID(t *testing.T) {
	st := &homeStub{streamID: "oldvid", apiKey: "secret"}
	st.lookup = func(_ context.Context, key, vid string) (youtube.LatestBroadcast, error) {
		if key != "secret" || vid != "oldvid" {
			t.Fatalf("key=%q vid=%q", key, vid)
		}
		return youtube.LatestBroadcast{VideoID: "newlive", ChannelID: "UCabc", Kind: youtube.BroadcastLive}, nil
	}
	got, err := New(st).LookupLatestStream("", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.StreamID != "newlive" || got.Kind != "live" || got.ChannelID != "UCabc" {
		t.Fatalf("%+v", got)
	}
	if st.streamID != "oldvid" {
		t.Fatal("lookup must not save")
	}
	if !st.notified {
		t.Fatal("expected run status refresh")
	}
}

func TestLookupLatestStream_prefersCallArgs(t *testing.T) {
	st := &homeStub{streamID: "old", apiKey: "saved"}
	st.lookup = func(_ context.Context, key, vid string) (youtube.LatestBroadcast, error) {
		if key != "formkey" || vid != "formvid" {
			t.Fatalf("key=%q vid=%q", key, vid)
		}
		return youtube.LatestBroadcast{VideoID: "soon", Kind: youtube.BroadcastUpcoming}, nil
	}
	got, err := New(st).LookupLatestStream(" formvid ", " formkey ")
	if err != nil {
		t.Fatal(err)
	}
	if got.StreamID != "soon" || got.Kind != "upcoming" {
		t.Fatalf("%+v", got)
	}
}

func TestLookupLatestStream_requiresIDAndKey(t *testing.T) {
	h := New(&homeStub{})
	if _, err := h.LookupLatestStream("", "k"); err == nil || !strings.Contains(err.Error(), "stream ID") {
		t.Fatalf("id err = %v", err)
	}
	if _, err := h.LookupLatestStream("vid", ""); err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("key err = %v", err)
	}
}
