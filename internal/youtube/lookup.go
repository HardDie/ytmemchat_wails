package youtube

import (
	"context"
	"fmt"
	"strings"

	yt "google.golang.org/api/youtube/v3"
)

// BroadcastKind is whether the resolved video is currently live or only scheduled.
type BroadcastKind string

const (
	// BroadcastLive is an active live stream.
	BroadcastLive BroadcastKind = "live"
	// BroadcastUpcoming is a scheduled stream that has not started.
	BroadcastUpcoming BroadcastKind = "upcoming"
)

// LatestBroadcast is a live or upcoming video on a channel (not a completed VOD).
type LatestBroadcast struct {
	// VideoID is the watch URL v= value.
	VideoID string
	// ChannelID is the UC… channel that owns the video.
	ChannelID string
	// Kind is [BroadcastLive] or [BroadcastUpcoming].
	Kind BroadcastKind
}

// LookupLatestBroadcast loads the channel from knownVideoID, then returns that
// video if it is still live, else the channel's current live stream, else the
// newest upcoming stream. Completed recordings are skipped. Requires a Data API key.
func LookupLatestBroadcast(ctx context.Context, apiKey, knownVideoID string) (LatestBroadcast, error) {
	c, err := newAPIClient(ctx, apiKey, nil, "")
	if err != nil {
		return LatestBroadcast{}, err
	}
	return c.(*apiClient).lookupLatestBroadcast(ctx, knownVideoID)
}

func (c *apiClient) lookupLatestBroadcast(ctx context.Context, videoID string) (LatestBroadcast, error) {
	videoID = strings.TrimSpace(videoID)
	if videoID == "" {
		return LatestBroadcast{}, fmt.Errorf("%w: empty id", ErrUnknownVideo)
	}
	resp, err := c.service.Videos.List([]string{"snippet", "liveStreamingDetails"}).Id(videoID).Context(ctx).Do()
	if err != nil {
		return LatestBroadcast{}, fmt.Errorf("youtube: videos.list: %w", mapAPIError(err))
	}
	if len(resp.Items) == 0 {
		return LatestBroadcast{}, fmt.Errorf("%w: %s", ErrUnknownVideo, videoID)
	}
	item := resp.Items[0]
	channelID := ""
	if item.Snippet != nil {
		channelID = strings.TrimSpace(item.Snippet.ChannelId)
	}
	if channelID == "" {
		return LatestBroadcast{}, fmt.Errorf("youtube: video %s has no channel id", videoID)
	}
	if isActiveLive(item) {
		return LatestBroadcast{VideoID: videoID, ChannelID: channelID, Kind: BroadcastLive}, nil
	}
	liveID, err := c.searchOneVideo(ctx, channelID, "live")
	if err != nil {
		return LatestBroadcast{}, err
	}
	if liveID != "" {
		return LatestBroadcast{VideoID: liveID, ChannelID: channelID, Kind: BroadcastLive}, nil
	}
	upID, err := c.searchOneVideo(ctx, channelID, "upcoming")
	if err != nil {
		return LatestBroadcast{}, err
	}
	if upID != "" {
		return LatestBroadcast{VideoID: upID, ChannelID: channelID, Kind: BroadcastUpcoming}, nil
	}
	return LatestBroadcast{ChannelID: channelID}, fmt.Errorf("%w", ErrNoBroadcast)
}

func isActiveLive(item *yt.Video) bool {
	if item == nil || item.LiveStreamingDetails == nil {
		return false
	}
	d := item.LiveStreamingDetails
	if d.ActiveLiveChatId != "" {
		return true
	}
	return d.ActualStartTime != "" && d.ActualEndTime == ""
}

func (c *apiClient) searchOneVideo(ctx context.Context, channelID, eventType string) (string, error) {
	resp, err := c.service.Search.List([]string{"id"}).
		ChannelId(channelID).
		Type("video").
		EventType(eventType).
		Order("date").
		MaxResults(1).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("youtube: search.list: %w", mapAPIError(err))
	}
	if len(resp.Items) == 0 || resp.Items[0].Id == nil {
		return "", nil
	}
	return strings.TrimSpace(resp.Items[0].Id.VideoId), nil
}
