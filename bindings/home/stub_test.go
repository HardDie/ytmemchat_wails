package home

import (
	"context"

	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
)

type homeStub struct {
	obs         OBSStatus
	run         RunStatus
	overlay     *obs.Server
	streamID    string
	apiKey      string
	ctx         context.Context
	lookup      func(context.Context, string, string) (youtube.LatestBroadcast, error)
	notified    bool
	startErr    error
	startCalled bool
	stopCalled  bool
}

func (s *homeStub) Start() error {
	s.startCalled = true
	return s.startErr
}

func (s *homeStub) Stop() { s.stopCalled = true }

func (s *homeStub) GetRunStatus() RunStatus { return s.run }

func (s *homeStub) GetOBSStatus() OBSStatus { return s.obs }

func (s *homeStub) OverlayServer() *obs.Server { return s.overlay }

func (s *homeStub) SavedYouTube() (streamID, apiKey string) {
	return s.streamID, s.apiKey
}

func (s *homeStub) DialogContext() context.Context { return s.ctx }

func (s *homeStub) LookupBroadcast(ctx context.Context, apiKey, videoID string) (youtube.LatestBroadcast, error) {
	if s.lookup == nil {
		return youtube.LatestBroadcast{}, nil
	}
	return s.lookup(ctx, apiKey, videoID)
}

func (s *homeStub) NotifyRun() { s.notified = true }
