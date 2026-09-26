package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/HardDie/ytmemchat_wails/internal/alerts"
	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/tts"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
)

const (
	msgFromChat    = "chat"
	msgFromTest    = "test"
	msgFromWebhook = "webhook"
)

// debugLog is Info-level and only called when settings debug is on.
// Tests replace it to capture lines.
var debugLog = func(msg string, args ...any) {
	slog.Info(msg, args...)
}

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
	debug bool
}

func (p *alertPublisher) Alert(msg string) bool {
	if p.debug {
		debugLog("debug: started parsing command", "text", msg)
	}
	if !p.inner.Alert(msg) {
		if p.debug {
			debugLog("debug: command not found", "text", msg)
		}
		return false
	}
	c := <-p.out
	if p.debug {
		debugLog("debug: found command", "file", c.Filename)
		path := commandMediaPath(p.inner.MediaPath(), c.Filename)
		if err := regularMediaFile(path); err != nil {
			debugLog("debug: error path to the file", "path", path, "err", err)
		} else {
			debugLog("debug: success", "path", path)
		}
	}
	p.srv.PublishOverlay(obs.AlertOverlay(c.Filename, c.Volume, c.Scale))
	return true
}

type ttsPublisher struct {
	eng   *tts.TTS
	out   <-chan tts.Speech
	srv   overlaySink
	debug bool
}

func (p *ttsPublisher) SynthesizeAudio(text string) error {
	if p.debug {
		debugLog("debug: started text to message", "text", text)
	}
	if err := p.eng.SynthesizeAudio(text); err != nil {
		if p.debug {
			debugLog(ttsDebugMessage(err), "err", err)
		}
		return err
	}
	sp := <-p.out
	p.srv.PublishOverlay(obs.TTSOverlay(sp.WAV, sp.Volume))
	return nil
}

func ttsDebugMessage(err error) string {
	if errors.Is(err, exec.ErrNotFound) {
		return "debug: app for tts not found"
	}
	return "debug: error text to message"
}

func commandMediaPath(mediaDir, filename string) string {
	filename = strings.TrimSpace(filename)
	mediaDir = strings.TrimSpace(mediaDir)
	if filename == "" {
		return mediaDir
	}
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(mediaDir, filepath.FromSlash(filename))
}

func regularMediaFile(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() {
		return fmt.Errorf("not a file")
	}
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
	match         alertMatcher
	speak         synthesizer
	sink          overlaySink
	debug         bool
	alertsEnabled bool
	commandsPath  string
	commandsErr   string
}

func (a *App) installOverlayLocked(strict bool) error {
	if a.httpSrv == nil {
		a.overlay.Store(nil)
		return nil
	}
	match, err := a.matcherFactory()(a.settings, a.httpSrv)
	commandsPath := ""
	commandsErr := ""
	if a.settings.Alerts.Enabled {
		commandsPath = a.settings.Alerts.EffectiveCommandsPath()
	}
	if err != nil {
		commandsErr = err.Error()
		if a.settings.Debug {
			debugLog("debug: error path to the file", "path", commandsPath, "err", commandsErr)
		}
		if strict {
			return err
		}
		slog.Error("alerts matcher", "err", err)
		match = nil
	} else if a.settings.Alerts.Enabled && match == nil {
		if commandsPath == "" {
			commandsErr = "commands file path is empty"
		} else if _, statErr := os.Stat(commandsPath); statErr != nil {
			commandsErr = statErr.Error()
		}
	}
	if a.settings.Debug && commandsErr != "" && err == nil {
		debugLog("debug: error path to the file", "path", commandsPath, "err", commandsErr)
	}
	if a.settings.Debug && match != nil {
		debugLog("debug: commands file loaded", "path", commandsPath)
	}
	speak, err := a.synthFactory()(a.settings, a.httpSrv)
	if err != nil {
		if a.settings.Debug {
			debugLog(ttsDebugMessage(err), "err", err)
		}
		if strict {
			return err
		}
		slog.Error("tts engine", "err", err)
		speak = nil
	}
	a.httpSrv.SetDebug(a.settings.Debug)
	a.overlay.Store(&overlayState{
		match:         match,
		speak:         speak,
		sink:          a.httpSrv,
		debug:         a.settings.Debug,
		alertsEnabled: a.settings.Alerts.Enabled,
		commandsPath:  commandsPath,
		commandsErr:   commandsErr,
	})
	return nil
}

func (a *App) dispatchLine(fallback overlaySink, source string, msg *youtube.ChatMessage) {
	st := a.overlay.Load()
	if msg != nil && overlayDebug(a, st) {
		debugLog("debug: received message", "source", source, "author", msg.Author, "text", msg.Message)
		if st != nil && st.alertsEnabled && st.match == nil {
			debugLog("debug: started parsing command", "text", msg.Message)
			debugLog("debug: error path to the file", "path", st.commandsPath, "err", st.commandsErr)
		}
	}
	if st != nil && st.sink != nil {
		dispatchChat(st.sink, st.match, st.speak, msg)
		return
	}
	if fallback != nil {
		dispatchChat(fallback, nil, nil, msg)
	}
}

func overlayDebug(a *App, st *overlayState) bool {
	if st != nil {
		return st.debug
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.Debug
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
			a.dispatchLine(srv, msgFromWebhook, &youtube.ChatMessage{
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
	path := s.Alerts.EffectiveCommandsPath()
	if path == "" {
		slog.Info("alerts enabled but commandsFilePath is empty; skipping matcher")
		return nil, nil
	}
	if strings.TrimSpace(s.Alerts.CommandsFilePath) == "" {
		_, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				slog.Info("alerts enabled but default commands.yaml is missing; skipping matcher")
				return nil, nil
			}
			return nil, fmt.Errorf("alerts commands file: %w", err)
		}
	}
	out := make(chan alerts.Clip, 1)
	inner, err := alerts.New(alerts.Config{
		Token:            s.Alerts.Token,
		MediaPath:        s.Alerts.MediaPath,
		CommandsFilePath: path,
		Out:              out,
	})
	if err != nil {
		return nil, fmt.Errorf("alerts commands file: %w", err)
	}
	return &alertPublisher{inner: inner, out: out, srv: srv, debug: s.Debug}, nil
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
	return &ttsPublisher{eng: eng, out: out, srv: srv, debug: s.Debug}, nil
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
