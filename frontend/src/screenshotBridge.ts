// Loaded only when the page URL has ?screenshot=1 (README capture).
// Installs fake Wails bindings so Vite can render panes without the Go runtime.

type Fn = (...args: unknown[]) => Promise<unknown>

function ok<T>(value: T): Fn {
  return () => Promise.resolve(value)
}

const settings = {
  streamId: 'xxxxxxxxxxx',
  apiKey: 'demo-key-not-real',
  port: '8080',
  ttsEnabled: true,
  ttsVoiceName: '',
  alertsEnabled: true,
  alertsToken: '@',
  alertsMediaPath: '/Users/demo/ytmemchat/media',
  alertsCommandsFilePath: '/Users/demo/ytmemchat/commands.yaml',
  webhookEnabled: false,
  interruptHotkeyEnabled: true,
  interruptHotkeyChord: 'Ctrl+Shift+I',
  interruptHotkeyError: '',
}

const obs = {
  listening: true,
  error: '',
  chatUrl: 'http://127.0.0.1:8080/obs/chat',
  overlayUrl: 'http://127.0.0.1:8080/obs/overlay',
  indexUrl: 'http://127.0.0.1:8080/',
}

const run = {
  running: false,
  connecting: false,
  usingApiKey: false,
  quotaUnits: 720,
  quotaUnitsLimit: 10000,
  quotaSearch: 1,
  quotaSearchLimit: 100,
  error: '',
}

const home: Record<string, Fn> = {
  GetOBSStatus: ok(obs),
  GetRunStatus: ok(run),
  InterruptTTS: ok(undefined),
  LookupLatestStream: ok({ streamId: settings.streamId, channelId: 'UCxxxxxxxx', kind: 'live' }),
  Start: ok(undefined),
  Stop: ok(undefined),
}

const sidebar: Record<string, Fn> = {
  AppVersion: ok('demo'),
}

const configuration: Record<string, Fn> = {
  ConfigPath: ok('/Users/demo/Library/Application Support/ytmemchat/config.json'),
  GetSettings: ok(settings),
  GetTTSVoices: ok([
    { name: 'Samantha', languages: 'en_US', gender: 'Female', details: '' },
    { name: 'Alex', languages: 'en_US', gender: 'Male', details: '' },
  ]),
  PickCommandsFile: ok(''),
  PickMediaDirectory: ok(''),
  SaveSettings: ok(undefined),
}

const commands: Record<string, Fn> = {
  GetAlertCommands: ok({
    path: settings.alertsCommandsFilePath,
    commands: [
      { name: 'jump', file: 'jump.webm', volume: 1, scale: 1 },
      { name: 'clap', file: 'clap.mp4', volume: 0.8, scale: 1.2 },
    ],
  }),
  PickAlertMediaFile: ok(''),
  PreviewAlert: ok(undefined),
  SaveAlertCommands: ok(undefined),
}

const testPane: Record<string, Fn> = {
  SendTestMessage: ok(undefined),
  FlushChat: ok(undefined),
}

const w = window as unknown as {
  go: {
    home: { Home: Record<string, Fn> }
    sidebar: { Sidebar: Record<string, Fn> }
    configuration: { Configuration: Record<string, Fn> }
    commands: { Commands: Record<string, Fn> }
    test: { Test: Record<string, Fn> }
  }
  runtime: {
    EventsOnMultiple: () => () => void
    ClipboardSetText: () => Promise<void>
  }
}

w.go = {
  home: { Home: home },
  sidebar: { Sidebar: sidebar },
  configuration: { Configuration: configuration },
  commands: { Commands: commands },
  test: { Test: testPane },
}
w.runtime = {
  EventsOnMultiple: () => () => {},
  ClipboardSetText: () => Promise.resolve(),
}

export {}
