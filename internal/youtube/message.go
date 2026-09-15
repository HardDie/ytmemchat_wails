package youtube

import (
	"context"
	"time"
)

// ChatMessage is a normalized live chat event (text, Super Chat, and similar).
type ChatMessage struct {
	// ID is the YouTube message id when the API provides one.
	ID string
	// Author is the sender display name.
	Author string
	// ImgURL is the sender avatar URL.
	ImgURL string
	// Message is the text, or a formatted Super Chat line.
	Message string
	// Type is the YouTube event type (for example "textMessageEvent").
	Type string
	// Timestamp is when the message was published.
	Timestamp time.Time
}

// Client opens a live chat iterator for a video ID.
type Client interface {
	// GetMessageIterator connects to chat, skips history, and returns an iterator.
	GetMessageIterator(ctx context.Context, liveVideoID string) (MessageIterator, error)
}

// MessageIterator is a stream of [ChatMessage] values until the context ends or chat stops.
type MessageIterator interface {
	// Next blocks until a message is available. ok is false when the stream is done.
	Next() (*ChatMessage, bool)
	// GetChan is the underlying channel for use in select.
	GetChan() <-chan *ChatMessage
}
