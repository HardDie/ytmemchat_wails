package youtube

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"

	"github.com/HardDie/ytmemchat_wails/internal/youtube/quota"
)

const defaultPollDelay = 5 * time.Second

type apiClient struct {
	service *yt.Service
	log     *slog.Logger
}

// New builds a Data API v3 client. apiKey must be non-empty after trim.
func New(apiKey string) (Client, error) {
	return newAPIClient(context.Background(), apiKey, nil, "")
}

func newAPIClient(ctx context.Context, apiKey string, hc *http.Client, endpoint string) (Client, error) {
	return newAPIClientTracked(ctx, apiKey, hc, endpoint, quota.Default)
}

func newAPIClientTracked(ctx context.Context, apiKey string, hc *http.Client, endpoint string, tr *quota.Tracker) (Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, ErrEmptyAPIKey
	}
	if tr == nil {
		tr = quota.Default
	}
	// Local spend estimate only; does not call Google for remaining quota.
	// WithHTTPClient skips option.WithAPIKey, so the key is set on the transport.
	hc = withAPIKey(quota.WrapClient(hc, tr), apiKey)
	opts := []option.ClientOption{option.WithHTTPClient(hc)}
	if endpoint != "" {
		opts = append(opts, option.WithEndpoint(endpoint))
	}
	svc, err := yt.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("youtube: create service: %w", mapAPIError(err))
	}
	return &apiClient{
		service: svc,
		log:     slog.Default(),
	}, nil
}

// GetMessageIterator implements [Client].
func (c *apiClient) GetMessageIterator(ctx context.Context, liveVideoID string) (MessageIterator, error) {
	liveChatID, err := c.getLiveChatID(ctx, liveVideoID)
	if err != nil {
		return nil, err
	}
	it := &apiIterator{
		ctx:          ctx,
		service:      c.service,
		liveChatID:   liveChatID,
		pollingDelay: defaultPollDelay,
		messageChan:  make(chan *ChatMessage, 10),
		log:          c.log,
	}
	if err = it.initializeToken(); err != nil {
		return nil, err
	}
	go it.startPolling()
	return it, nil
}

type apiIterator struct {
	ctx          context.Context
	service      *yt.Service
	liveChatID   string
	pageToken    string
	pollingDelay time.Duration
	messageChan  chan *ChatMessage
	log          *slog.Logger
}

func (it *apiIterator) initializeToken() error {
	call := it.service.LiveChatMessages.List(it.liveChatID, []string{"snippet"}).Context(it.ctx)
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("youtube: initial chat list: %w", mapAPIError(err))
	}
	it.pageToken = resp.NextPageToken
	if resp.PollingIntervalMillis > 0 {
		it.pollingDelay = time.Duration(resp.PollingIntervalMillis) * time.Millisecond
	}
	return nil
}

// Next implements [MessageIterator].
func (it *apiIterator) Next() (*ChatMessage, bool) {
	select {
	case msg, ok := <-it.messageChan:
		return msg, ok
	case <-it.ctx.Done():
		return nil, false
	}
}

// GetChan implements [MessageIterator].
func (it *apiIterator) GetChan() <-chan *ChatMessage {
	return it.messageChan
}

func (it *apiIterator) startPolling() {
	defer close(it.messageChan)
	for {
		select {
		case <-it.ctx.Done():
			return
		case <-time.After(it.pollingDelay):
		}
		resp, err := it.service.LiveChatMessages.List(it.liveChatID, []string{"snippet", "authorDetails"}).
			PageToken(it.pageToken).
			Context(it.ctx).
			Do()
		if err != nil {
			if it.ctx.Err() != nil {
				return
			}
			if strings.Contains(err.Error(), "liveChatEnded") {
				it.log.Info("youtube live chat ended")
				return
			}
			it.log.Error("youtube poll failed", slog.String("err", err.Error()))
			continue
		}
		for _, msg := range resp.Items {
			chatMsg := convertToChatMessage(msg)
			select {
			case it.messageChan <- chatMsg:
			case <-it.ctx.Done():
				return
			}
		}
		it.pageToken = resp.NextPageToken
		if resp.PollingIntervalMillis > 0 {
			it.pollingDelay = time.Duration(resp.PollingIntervalMillis) * time.Millisecond
		}
	}
}

func (c *apiClient) getLiveChatID(ctx context.Context, videoID string) (string, error) {
	call := c.service.Videos.List([]string{"liveStreamingDetails"}).Id(videoID).Context(ctx)
	response, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("youtube: videos.list: %w", mapAPIError(err))
	}
	if len(response.Items) == 0 || response.Items[0].LiveStreamingDetails == nil || response.Items[0].LiveStreamingDetails.ActiveLiveChatId == "" {
		return "", fmt.Errorf("%w: %s", ErrNotLive, videoID)
	}
	return response.Items[0].LiveStreamingDetails.ActiveLiveChatId, nil
}

func convertToChatMessage(ym *yt.LiveChatMessage) *ChatMessage {
	if ym == nil {
		return &ChatMessage{}
	}
	out := &ChatMessage{ID: ym.Id}
	if ym.Snippet != nil {
		out.Type = ym.Snippet.Type
		out.Message = ym.Snippet.DisplayMessage
		if t, err := time.Parse(time.RFC3339, ym.Snippet.PublishedAt); err == nil {
			out.Timestamp = t
		}
		if ym.Snippet.Type == "superChatEvent" && ym.Snippet.SuperChatDetails != nil {
			d := ym.Snippet.SuperChatDetails
			out.Message = fmt.Sprintf("[%s %s] %s", d.Currency, formatMoney(d.AmountMicros), out.Message)
		}
	}
	if ym.AuthorDetails != nil {
		out.Author = ym.AuthorDetails.DisplayName
		out.ImgURL = ym.AuthorDetails.ProfileImageUrl
	}
	return out
}

func formatMoney(micros uint64) string {
	return strconv.FormatFloat(float64(micros)/1_000_000.0, 'f', 2, 64)
}
