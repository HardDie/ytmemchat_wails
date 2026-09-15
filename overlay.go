package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/HardDie/ytmemchat_wails/internal/alerts"
	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/tts"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
)

// alertMatcher is alerts.Alerts or a test fake. Nil means alerts are off.
type alertMatcher interface {
	Alert(msg string) bool
}

// synthesizer is tts.TTS or a test fake. Nil means TTS is off.
type synthesizer interface {
	SynthesizeAudio(text string) error
}

type matcherFactory func(config.Settings, *obs.Server) (alertMatcher, error)
type synthFactory func(config.Settings, *obs.Server) (synthesizer, error)

type overlaySink interface {
	PublishChat(obs.ChatEvent)
	PublishOverlay(obs.OverlayEvent)
}

type alertPublisher struct {
	inner *alerts.Alerts
	out   <-chan alerts.Clip
	srv   overlaySink
}

func (p *alertPublisher) Alert(msg string) bool {
	if !p.inner.Alert(msg) {
		return false
	}
	c := <-p.out
	p.srv.PublishOverlay(obs.AlertOverlay(c.Filename, c.Volume, c.Scale))
	return true
}

type ttsPublisher struct {
	eng *tts.TTS
	out <-chan tts.Speech
	srv overlaySink
}

func (p *ttsPublisher) SynthesizeAudio(text string) error {
	if err := p.eng.SynthesizeAudio(text); err != nil {
		return err
	}
	sp := <-p.out
	p.srv.PublishOverlay(obs.TTSOverlay(sp.WAV, sp.Volume))
	return nil
}

func (a *App) matcherFactory() matcherFactory {
	if a.newMatcher != nil {
		return a.newMatcher
	}
	return defaultMatcher
}

func (a *App) synthFactory() synthFactory {
	if a.newSynth != nil {
		return a.newSynth
	}
	return defaultSynth
}

type overlayState struct {
	match alertMatcher
	speak synthesizer
	sink  overlaySink
}

func (a *App) installOverlayLocked(strict bool) error {
	if a.httpSrv == nil {
		a.overlay.Store(nil)
		return nil
	}
	match, err := a.matcherFactory()(a.settings, a.httpSrv)
	if err != nil {
		if strict {
			return err
		}
		slog.Error("alerts matcher", "err", err)
		match = nil
	}
	speak, err := a.synthFactory()(a.settings, a.httpSrv)
	if err != nil {
		if strict {
			return err
		}
		slog.Error("tts engine", "err", err)
		speak = nil
	}
	a.overlay.Store(&overlayState{match: match, speak: speak, sink: a.httpSrv})
	return nil
}

func (a *App) dispatchLine(fallback overlaySink, msg *youtube.ChatMessage) {
	st := a.overlay.Load()
	if st != nil && st.sink != nil {
		dispatchChat(st.sink, st.match, st.speak, msg)
		return
	}
	if fallback != nil {
		dispatchChat(fallback, nil, nil, msg)
	}
}

func (a *App) drainInjected(ctx context.Context, wg *sync.WaitGroup, srv *obs.Server) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-srv.Injected():
			if !ok {
				return
			}
			a.dispatchLine(srv, &youtube.ChatMessage{
				Author:    msg.Author,
				Message:   msg.Text,
				Timestamp: msg.SentAt,
			})
		}
	}
}

func defaultMatcher(s config.Settings, srv *obs.Server) (alertMatcher, error) {
	if !s.Alerts.Enabled {
		return nil, nil
	}
	if strings.TrimSpace(s.Alerts.CommandsFilePath) == "" {
		slog.Info("alerts enabled but commandsFilePath is empty; skipping matcher")
		return nil, nil
	}
	out := make(chan alerts.Clip, 1)
	inner, err := alerts.New(alerts.Config{
		Token:            s.Alerts.Token,
		MediaPath:        s.Alerts.MediaPath,
		CommandsFilePath: s.Alerts.CommandsFilePath,
		Out:              out,
	})
	if err != nil {
		return nil, fmt.Errorf("alerts commands file: %w", err)
	}
	return &alertPublisher{inner: inner, out: out, srv: srv}, nil
}

func defaultSynth(s config.Settings, srv *obs.Server) (synthesizer, error) {
	if !s.TTS.Enabled {
		return nil, nil
	}
	out := make(chan tts.Speech, 1)
	eng := tts.New(tts.Config{
		VoiceName: s.TTS.VoiceName,
		Out:       out,
	})
	return &ttsPublisher{eng: eng, out: out, srv: srv}, nil
}

func dispatchChat(sink overlaySink, match alertMatcher, speak synthesizer, msg *youtube.ChatMessage) {
	if msg == nil {
		return
	}
	sink.PublishChat(obs.NewChatEvent(msg.Author, msg.ImgURL, msg.Message, msg.Timestamp))
	if match != nil && match.Alert(msg.Message) {
		return
	}
	if speak == nil {
		return
	}
	if err := speak.SynthesizeAudio(msg.Message); err != nil {
		slog.Error("tts", "err", err)
	}
}
