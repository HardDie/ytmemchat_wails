package nokey

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/HardDie/ytmemchat_wails/internal/youtube"
	"github.com/PuerkitoBio/goquery"
)

const (
	defaultUA          = "Firefox/99"
	defaultPollDelay   = 5 * time.Second
	liveChatURLFmt     = "https://www.youtube.com/live_chat?v=%s"
	getLiveChatURL     = "https://www.youtube.com/youtubei/v1/live_chat/get_live_chat"
	innertubeClientVer = "2.9999099"
)

type apiClient struct {
	log        *slog.Logger
	http       *http.Client
	pollDelay  time.Duration
	liveChat   string
	getChatURL string
}

// New returns a no-key live chat client using [http.DefaultClient].
func New() youtube.Client {
	return newClient(http.DefaultClient, liveChatURLFmt, getLiveChatURL, defaultPollDelay)
}

func newClient(hc *http.Client, liveFmt, getURL string, delay time.Duration) youtube.Client {
	if hc == nil {
		hc = http.DefaultClient
	}
	if delay <= 0 {
		delay = defaultPollDelay
	}
	return &apiClient{
		log:        slog.Default(),
		http:       hc,
		pollDelay:  delay,
		liveChat:   liveFmt,
		getChatURL: getURL,
	}
}

// GetMessageIterator implements [youtube.Client].
func (c *apiClient) GetMessageIterator(ctx context.Context, liveVideoID string) (youtube.MessageIterator, error) {
	it := &htmlIterator{
		ctx:          ctx,
		liveVideoID:  liveVideoID,
		pollingDelay: c.pollDelay,
		messageChan:  make(chan *youtube.ChatMessage, 10),
		log:          c.log,
		http:         c.http,
		liveChat:     c.liveChat,
		getChatURL:   c.getChatURL,
	}
	if err := it.initializeToken(); err != nil {
		return nil, err
	}
	go it.startPolling()
	return it, nil
}

type htmlIterator struct {
	ctx          context.Context
	liveVideoID  string
	pageToken    string
	pollingDelay time.Duration
	messageChan  chan *youtube.ChatMessage
	log          *slog.Logger
	http         *http.Client
	liveChat     string
	getChatURL   string
}

func (it *htmlIterator) initializeToken() error {
	u := fmt.Sprintf(it.liveChat, url.QueryEscape(it.liveVideoID))
	req, err := http.NewRequestWithContext(it.ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", defaultUA)
	resp, err := it.http.Do(req)
	if err != nil {
		return fmt.Errorf("nokey: fetch live_chat: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("nokey: read live_chat: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("nokey: live_chat HTTP %d", resp.StatusCode)
	}
	token, err := parseContinuationFromHTML(string(body))
	if err != nil {
		return err
	}
	it.pageToken = token
	return nil
}

// Next implements [youtube.MessageIterator].
func (it *htmlIterator) Next() (*youtube.ChatMessage, bool) {
	select {
	case msg, ok := <-it.messageChan:
		return msg, ok
	case <-it.ctx.Done():
		return nil, false
	}
}

// GetChan implements [youtube.MessageIterator].
func (it *htmlIterator) GetChan() <-chan *youtube.ChatMessage {
	return it.messageChan
}

func (it *htmlIterator) startPolling() {
	defer close(it.messageChan)
	for {
		select {
		case <-it.ctx.Done():
			return
		case <-time.After(it.pollingDelay):
		}
		result, err := it.postGetLiveChat()
		if err != nil {
			if it.ctx.Err() != nil {
				return
			}
			it.log.Error("nokey poll failed", slog.String("err", err.Error()))
			continue
		}
		for _, action := range result.ContinuationContents.LiveChatContinuation.Actions {
			chatMsg := convertChatItem(action.AddChatItemAction.Item)
			if chatMsg.Message == "" {
				continue
			}
			select {
			case it.messageChan <- chatMsg:
			case <-it.ctx.Done():
				return
			}
		}
		applyContinuation(it, result)
	}
}

func (it *htmlIterator) postGetLiveChat() (ytInternalResponse, error) {
	payload := map[string]any{
		"context": map[string]any{
			"client": map[string]any{"clientName": "WEB", "clientVersion": innertubeClientVer},
		},
		"continuation": it.pageToken,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return ytInternalResponse{}, err
	}
	u, err := url.Parse(it.getChatURL)
	if err != nil {
		return ytInternalResponse{}, err
	}
	q := u.Query()
	q.Set("prettyPrint", "false")
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(it.ctx, http.MethodPost, u.String(), bytes.NewReader(raw))
	if err != nil {
		return ytInternalResponse{}, err
	}
	req.Header.Set("User-Agent", defaultUA)
	req.Header.Set("Content-Type", "application/json")
	resp, err := it.http.Do(req)
	if err != nil {
		return ytInternalResponse{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ytInternalResponse{}, err
	}
	if resp.StatusCode >= 400 {
		return ytInternalResponse{}, fmt.Errorf("nokey: get_live_chat HTTP %d", resp.StatusCode)
	}
	var result ytInternalResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return ytInternalResponse{}, fmt.Errorf("nokey: decode get_live_chat: %w", err)
	}
	return result, nil
}

func applyContinuation(it *htmlIterator, result ytInternalResponse) {
	conns := result.ContinuationContents.LiveChatContinuation.Continuations
	if len(conns) == 0 {
		return
	}
	switch {
	case conns[0].TimedContinuationData != nil:
		if conns[0].TimedContinuationData.Continuation != "" {
			it.pageToken = conns[0].TimedContinuationData.Continuation
		}
		if conns[0].TimedContinuationData.TimeoutMs > 0 {
			it.pollingDelay = time.Duration(conns[0].TimedContinuationData.TimeoutMs) * time.Millisecond
		}
	case conns[0].InvalidationContinuationData != nil:
		if conns[0].InvalidationContinuationData.Continuation != "" {
			it.pageToken = conns[0].InvalidationContinuationData.Continuation
		}
		if conns[0].InvalidationContinuationData.TimeoutMs > 0 {
			it.pollingDelay = time.Duration(conns[0].InvalidationContinuationData.TimeoutMs) * time.Millisecond
		}
	default:
		it.pollingDelay = defaultPollDelay
	}
}

func parseContinuationFromHTML(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("nokey: parse HTML: %w", err)
	}
	var token string
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		scriptContent := s.Text()
		if !strings.Contains(scriptContent, "liveChatRenderer") {
			return
		}
		parts := strings.Split(scriptContent, `"continuation":"`)
		if len(parts) > 1 {
			token = strings.Split(parts[1], `"`)[0]
		}
	})
	if token == "" {
		return "", fmt.Errorf("nokey: failed to extract chat continuation")
	}
	return token, nil
}

func convertChatItem(item chatItem) *youtube.ChatMessage {
	switch {
	case item.LiveChatPaidMessageRenderer != nil:
		return convertRenderer(item.LiveChatPaidMessageRenderer, "superChatEvent", true)
	case item.LiveChatPaidStickerRenderer != nil:
		return convertRenderer(item.LiveChatPaidStickerRenderer, "superStickerEvent", true)
	case item.LiveChatTextMessageRenderer != nil:
		return convertRenderer(item.LiveChatTextMessageRenderer, "textMessageEvent", false)
	default:
		return &youtube.ChatMessage{}
	}
}

func convertRenderer(r *liveChatMessageRenderer, typ string, paid bool) *youtube.ChatMessage {
	if r == nil {
		return &youtube.ChatMessage{}
	}
	msg := messageFromRuns(r.Message.Runs)
	if paid {
		amount := strings.TrimSpace(r.PurchaseAmountText.SimpleText)
		if amount != "" {
			if msg == "" {
				msg = "[" + amount + "]"
			} else {
				msg = fmt.Sprintf("[%s] %s", amount, msg)
			}
		}
	}
	return &youtube.ChatMessage{
		ID:        r.ID,
		Author:    r.AuthorName.SimpleText,
		ImgURL:    avatarURL(r.AuthorPhoto.Thumbnails),
		Message:   msg,
		Type:      typ,
		Timestamp: parseTimestampUsec(r.TimestampUsec),
	}
}

func messageFromRuns(runs []messageRun) string {
	var b strings.Builder
	for _, run := range runs {
		if run.Text != "" {
			b.WriteString(run.Text)
			continue
		}
		if run.Emoji != nil && len(run.Emoji.Shortcuts) > 0 {
			b.WriteString(run.Emoji.Shortcuts[0])
		}
	}
	return b.String()
}

func avatarURL(thumbs []thumbnail) string {
	best := ""
	bestW := -1
	for _, t := range thumbs {
		if t.URL == "" {
			continue
		}
		if t.Width >= bestW {
			bestW = t.Width
			best = t.URL
		}
	}
	if strings.HasPrefix(best, "//") {
		return "https:" + best
	}
	return best
}

func parseTimestampUsec(s string) time.Time {
	usec, err := strconv.ParseInt(s, 10, 64)
	if err != nil || usec <= 0 {
		return time.Time{}
	}
	return time.UnixMicro(usec).UTC()
}

type ytInternalResponse struct {
	ContinuationContents struct {
		LiveChatContinuation struct {
			Continuations []struct {
				TimedContinuationData *struct {
					Continuation string `json:"continuation"`
					TimeoutMs    int64  `json:"timeoutMs"`
				} `json:"timedContinuationData"`
				InvalidationContinuationData *struct {
					Continuation string `json:"continuation"`
					TimeoutMs    int64  `json:"timeoutMs"`
				} `json:"invalidationContinuationData"`
			} `json:"continuations"`
			Actions []struct {
				AddChatItemAction struct {
					Item chatItem `json:"item"`
				} `json:"addChatItemAction"`
			} `json:"actions"`
		} `json:"liveChatContinuation"`
	} `json:"continuationContents"`
}

type chatItem struct {
	LiveChatTextMessageRenderer *liveChatMessageRenderer `json:"liveChatTextMessageRenderer"`
	LiveChatPaidMessageRenderer *liveChatMessageRenderer `json:"liveChatPaidMessageRenderer"`
	LiveChatPaidStickerRenderer *liveChatMessageRenderer `json:"liveChatPaidStickerRenderer"`
}

type liveChatMessageRenderer struct {
	ID            string `json:"id"`
	TimestampUsec string `json:"timestampUsec"`
	AuthorName    struct {
		SimpleText string `json:"simpleText"`
	} `json:"authorName"`
	AuthorPhoto struct {
		Thumbnails []thumbnail `json:"thumbnails"`
	} `json:"authorPhoto"`
	Message struct {
		Runs []messageRun `json:"runs"`
	} `json:"message"`
	PurchaseAmountText struct {
		SimpleText string `json:"simpleText"`
	} `json:"purchaseAmountText"`
}

type thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type messageRun struct {
	Text  string `json:"text"`
	Emoji *struct {
		Shortcuts []string `json:"shortcuts"`
	} `json:"emoji"`
}
