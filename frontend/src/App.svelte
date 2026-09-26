<script lang="ts">
  import { onMount } from 'svelte'
  import { AppVersion } from '../wailsjs/go/sidebar/Sidebar.js'
  import {
    GetOBSStatus,
    GetRunStatus,
    InterruptTTS,
    LookupLatestStream,
    Start,
    Stop,
  } from '../wailsjs/go/home/Home.js'
  import {
    ConfigPath,
    GetSettings,
    PickCommandsFile,
    PickMediaDirectory,
    SaveSettings,
  } from '../wailsjs/go/configuration/Configuration.js'
  import {
    GetAlertCommands,
    PickAlertMediaFile,
    PreviewAlert,
    SaveAlertCommands,
  } from '../wailsjs/go/commands/Commands.js'
  import { FlushChat, SendTestMessage } from '../wailsjs/go/test/Test.js'
  import { ApplyAndQuit, Check as CheckUpdate, Download as DownloadUpdate } from '../wailsjs/go/update/Update.js'
  import { commands as cmdModels, configuration, home, update as updateModels } from '../wailsjs/go/models'
  import { BrowserOpenURL, ClipboardSetText, EventsOn } from '../wailsjs/runtime/runtime'
  import HomePane from './lib/HomePane.svelte'
  import ConfigPane from './lib/ConfigPane.svelte'
  import CommandsPane from './lib/CommandsPane.svelte'
  import TestPane from './lib/TestPane.svelte'
  import UpdatePane from './lib/UpdatePane.svelte'
  import Notifications from './lib/Notifications.svelte'
  import type { Notice, NoticeKind } from './lib/Notifications.svelte'

  type Page = 'home' | 'config' | 'commands' | 'test' | 'update'

  const titles: Record<Page, string> = {
    home: 'Home',
    config: 'Configuration',
    commands: 'Commands',
    test: 'Test message',
    update: 'Update',
  }

  let page: Page = 'home'
  let streamId = ''
  let apiKey = ''
  let port = '8080'
  let ttsEnabled = false
  let ttsVoiceName = ''
  let alertsEnabled = false
  let alertsToken = '@'
  let alertsMediaPath = ''
  let alertsCommandsFilePath = ''
  let alertsCommandsFileCustom = false
  let webhookEnabled = false
  let interruptHotkeyEnabled = true
  let interruptHotkeyChord = 'Ctrl+Shift+I'
  let interruptHotkeyError = ''
  let apiKeyInKeychain = false
  let debug = false
  let testMessage = ''
  let configPath = ''
  let saving = false
  let starting = false
  let lookingUp = false
  let commandsSaving = false
  let commandsLoading = false
  let commandsPath = ''
  let commandRows: Array<{ name: string; file: string; volume: string; scale: string }> = []
  let obs: home.OBSStatus | null = null
  let run: home.RunStatus | null = null
  let appVersion = ''
  let updateStatus: updateModels.Status | null = null
  let updateBusy = false
  let updateDownloaded = false

  let notices: Notice[] = []
  let noticeSeq = 0

  function notify(text: string, kind: NoticeKind = 'ok'): void {
    const id = ++noticeSeq
    notices = [{ id, text, kind }, ...notices]
  }

  function applyForm(s: configuration.SettingsForm): void {
    streamId = s.streamId ?? ''
    apiKey = s.apiKey ?? ''
    port = s.port || '8080'
    ttsEnabled = !!s.ttsEnabled
    ttsVoiceName = s.ttsVoiceName ?? ''
    alertsEnabled = !!s.alertsEnabled
    alertsToken = s.alertsToken || '@'
    alertsMediaPath = s.alertsMediaPath ?? ''
    alertsCommandsFilePath = s.alertsCommandsFilePath ?? ''
    alertsCommandsFileCustom = !!s.alertsCommandsFileCustom
    webhookEnabled = !!s.webhookEnabled
    interruptHotkeyEnabled = s.interruptHotkeyEnabled !== false
    interruptHotkeyChord = s.interruptHotkeyChord || 'Ctrl+Shift+I'
    interruptHotkeyError = s.interruptHotkeyError ?? ''
    apiKeyInKeychain = !!s.apiKeyInKeychain
    debug = !!s.debug
  }

  function formPayload(): configuration.SettingsForm {
    return configuration.SettingsForm.createFrom({
      streamId,
      apiKey,
      port,
      ttsEnabled,
      ttsVoiceName,
      alertsEnabled,
      alertsToken,
      alertsMediaPath,
      alertsCommandsFilePath,
      alertsCommandsFileCustom,
      webhookEnabled,
      interruptHotkeyEnabled,
      interruptHotkeyChord,
      debug,
    })
  }

  async function refreshOBS(): Promise<void> {
    obs = await GetOBSStatus()
  }

  async function refreshRun(): Promise<void> {
    run = await GetRunStatus()
  }

  onMount(async () => {
    EventsOn('pipeline', (s: home.RunStatus) => {
      run = s
    })
    const quotaTick = window.setInterval(() => {
      GetRunStatus()
        .then((s) => {
          run = s
        })
        .catch(() => {
          /* keep last status */
        })
    }, 2000)
    try {
      const [s, path, ver] = await Promise.all([GetSettings(), ConfigPath(), AppVersion()])
      applyForm(s)
      configPath = path
      appVersion = ver || 'dev'
      await refreshOBS()
      await refreshRun()
    } catch (e) {
      notify(String(e), 'err')
    }
    return () => {
      window.clearInterval(quotaTick)
    }
  })

  async function save(silent = false): Promise<boolean> {
    saving = true
    try {
      await SaveSettings(formPayload())
      applyForm(await GetSettings())
      await refreshOBS()
      await refreshRun()
      if (!silent) {
        notify('Settings saved')
      }
      return true
    } catch (e) {
      notify(String(e), 'err')
      try {
        await refreshOBS()
      } catch {
        /* keep save error */
      }
      return false
    } finally {
      saving = false
    }
  }

  async function start(): Promise<void> {
    starting = true
    try {
      if (!(await save(true))) {
        return
      }
      await Start()
      await refreshRun()
    } catch (e) {
      notify(String(e), 'err')
      await refreshRun()
    } finally {
      starting = false
    }
  }

  async function stop(): Promise<void> {
    await Stop()
    await refreshRun()
  }

  async function copy(url: string): Promise<void> {
    await ClipboardSetText(url)
    notify('URL copied')
  }

  function copyChat(): void {
    if (obs) copy(obs.chatUrl)
  }

  function copyOverlay(): void {
    if (obs) copy(obs.overlayUrl)
  }

  async function sendTest(): Promise<void> {
    try {
      await SendTestMessage(testMessage)
      notify('Test message sent')
    } catch (e) {
      notify(String(e), 'err')
    }
  }

  async function flushChat(): Promise<void> {
    try {
      await FlushChat()
      notify('Chat flushed')
    } catch (e) {
      notify(String(e), 'err')
    }
  }

  async function checkUpdate(): Promise<void> {
    updateBusy = true
    updateDownloaded = false
    try {
      updateStatus = await CheckUpdate()
      notify('Checked GitHub')
    } catch (e) {
      notify(String(e), 'err')
    } finally {
      updateBusy = false
    }
  }

  async function downloadUpdate(): Promise<void> {
    updateBusy = true
    try {
      await DownloadUpdate()
      updateDownloaded = true
      notify('Archive verified')
    } catch (e) {
      notify(String(e), 'err')
    } finally {
      updateBusy = false
    }
  }

  async function applyUpdate(): Promise<void> {
    updateBusy = true
    try {
      await ApplyAndQuit()
      notify('Installing…')
    } catch (e) {
      notify(String(e), 'err')
      updateBusy = false
    }
  }

  function openReleases(): void {
    const url = (updateStatus && updateStatus.url) || 'https://github.com/HardDie/ytmemchat_wails/releases'
    BrowserOpenURL(url)
  }

  async function interruptTTS(): Promise<void> {
    try {
      await InterruptTTS()
    } catch (e) {
      notify(String(e), 'err')
    }
  }

  async function lookupLatest(): Promise<void> {
    lookingUp = true
    try {
      const got = await LookupLatestStream(streamId, apiKey)
      if (got && got.streamId) {
        streamId = got.streamId
        const kind = got.kind === 'upcoming' ? 'upcoming' : 'live'
        notify(`Latest ${kind} stream: ${got.streamId}`)
      }
      await refreshRun()
    } catch (e) {
      notify(String(e), 'err')
    } finally {
      lookingUp = false
    }
  }

  async function pickYaml(): Promise<void> {
    try {
      const p = await PickCommandsFile()
      if (!p) {
        return
      }
      alertsCommandsFilePath = p
      alertsCommandsFileCustom = true
      await save()
    } catch (e) {
      notify(String(e), 'err')
    }
  }

  async function testCommand(index: number): Promise<void> {
    const row = commandRows[index]
    if (!row || !row.file.trim()) {
      notify('Command needs a file', 'err')
      return
    }
    try {
      const volume = row.volume.trim() ? Number(row.volume) : 1
      const scale = row.scale.trim() ? Number(row.scale) : 1
      await PreviewAlert(row.file.trim(), volume, scale)
      const label = row.name.trim()
      notify(label ? `Alert sent to overlay: ${label}` : 'Alert sent to overlay')
    } catch (e) {
      notify(String(e), 'err')
    }
  }

  async function pickCommandFile(index: number): Promise<void> {
    try {
      const p = await PickAlertMediaFile(alertsMediaPath)
      if (!p) {
        return
      }
      commandRows = commandRows.map((row, i) => (i === index ? { ...row, file: p } : row))
    } catch (e) {
      notify(String(e), 'err')
    }
  }

  async function pickMedia(): Promise<void> {
    try {
      const p = await PickMediaDirectory()
      if (!p) {
        return
      }
      alertsMediaPath = p
      await save()
    } catch (e) {
      notify(String(e), 'err')
    }
  }

  function optionalFloat(raw: string, label: string): number | undefined {
    const t = raw.trim().replace(/,/g, '.').replace(/\.$/, '')
    if (!t) {
      return undefined
    }
    if (!/^(?:\d+(?:\.\d{1,2})?|\.\d{1,2})$/.test(t)) {
      throw new Error(`${label} must be a number with at most two digits after the decimal point`)
    }
    const n = Number(t)
    if (!Number.isFinite(n)) {
      throw new Error(`${label} must be a number`)
    }
    return n
  }

  function formatDecimal(n: number): string {
    return String(Number(n.toFixed(2)))
  }

  function emptyRow(): { name: string; file: string; volume: string; scale: string } {
    return { name: '', file: '', volume: '', scale: '' }
  }

  async function loadCommands(): Promise<void> {
    if (!alertsCommandsFilePath.trim()) {
      commandsPath = ''
      commandRows = []
      return
    }
    commandsLoading = true
    try {
      const got = await GetAlertCommands()
      commandsPath = got.path ?? ''
      commandRows = (got.commands ?? []).map((c) => ({
        name: c.name ?? '',
        file: c.file ?? '',
        volume: c.volume == null ? '' : formatDecimal(c.volume),
        scale: c.scale == null ? '' : formatDecimal(c.scale),
      }))
    } catch (e) {
      commandsPath = ''
      commandRows = []
      notify(String(e), 'err')
    } finally {
      commandsLoading = false
    }
  }

  async function openCommands(): Promise<void> {
    page = 'commands'
    await loadCommands()
  }

  function reorderCommands(next: Array<{ name: string; file: string; volume: string; scale: string }>): void {
    commandRows = next
  }

  function addCommand(): void {
    commandRows = [...commandRows, emptyRow()]
  }

  function removeCommand(index: number): void {
    commandRows = commandRows.filter((_, i) => i !== index)
  }

  async function saveCommands(): Promise<void> {
    commandsSaving = true
    try {
      const commands = commandRows.map((row, i) => {
        const name = row.name.trim()
        const file = row.file.trim()
        if (!name || !file) {
          throw new Error(`Command ${i + 1} needs a name and file`)
        }
        const item: { name: string; file: string; volume?: number; scale?: number } = { name, file }
        const volume = optionalFloat(row.volume, `Command ${name} volume`)
        const scale = optionalFloat(row.scale, `Command ${name} scale`)
        if (volume !== undefined) {
          item.volume = volume
        }
        if (scale !== undefined) {
          item.scale = scale
        }
        return item
      })
      await SaveAlertCommands(cmdModels.AlertCommandsFile.createFrom({ path: commandsPath, commands }))
      await loadCommands()
      notify('Commands saved')
    } catch (e) {
      notify(String(e), 'err')
    } finally {
      commandsSaving = false
    }
  }
</script>

<div class="shell">
  <aside class="sidebar">
    <div class="brand">
      <div class="brand-name">ytmemchat</div>
      <div class="brand-sub">Operator console</div>
    </div>
    <nav class="panes" aria-label="Window panes">
      <button class="pane-btn" class:active={page === 'home'} type="button" on:click={() => { page = 'home' }}>Home</button>
      <button class="pane-btn" class:active={page === 'config'} type="button" on:click={() => { page = 'config' }}>Configuration</button>
      <button class="pane-btn" class:active={page === 'commands'} type="button" on:click={openCommands}>Commands</button>
      <button class="pane-btn" class:active={page === 'test'} type="button" on:click={() => { page = 'test' }}>Test</button>
      <button class="pane-btn" class:active={page === 'update'} type="button" on:click={() => { page = 'update' }}>Update</button>
    </nav>
    {#if appVersion}
      <div class="sidebar-version" title={appVersion}>{appVersion}</div>
    {/if}
  </aside>

  <div class="workspace">
    <header class="toolbar">
      <h1>{titles[page]}</h1>
      <span class="toolbar-note">Not shown on stream</span>
    </header>

    <div class="content">
      {#if page === 'config'}
        <ConfigPane
          bind:streamId
          bind:apiKey
          bind:port
          bind:ttsEnabled
          bind:ttsVoiceName
          bind:alertsEnabled
          bind:alertsToken
          bind:alertsMediaPath
          bind:alertsCommandsFilePath
          bind:alertsCommandsFileCustom
          bind:webhookEnabled
          bind:interruptHotkeyEnabled
          bind:interruptHotkeyChord
          interruptHotkeyError={interruptHotkeyError}
          {apiKeyInKeychain}
          bind:debug
          {saving}
          {configPath}
          onSave={async () => {
            await save()
          }}
          onLookup={lookupLatest}
          lookingUp={lookingUp}
          onPickYaml={pickYaml}
          onPickMedia={pickMedia}
        />
      {:else if page === 'commands'}
        <CommandsPane
          path={commandsPath}
          rows={commandRows}
          saving={commandsSaving}
          loading={commandsLoading}
          onAdd={addCommand}
          onRemove={removeCommand}
          onReorder={reorderCommands}
          onSave={saveCommands}
          onReload={loadCommands}
          mediaPath={alertsMediaPath}
          onPickFile={pickCommandFile}
          onTest={testCommand}
          canTest={!!(obs && obs.listening)}
          {alertsEnabled}
        />
      {:else if page === 'test'}
        <TestPane bind:testMessage {obs} onSend={sendTest} onFlush={flushChat} />
      {:else if page === 'update'}
        <UpdatePane current={appVersion} status={updateStatus} busy={updateBusy} downloaded={updateDownloaded} onCheck={checkUpdate} onDownload={downloadUpdate} onApply={applyUpdate} onOpen={openReleases} />
      {:else}
        <HomePane {run} {obs} {streamId} {apiKey} {starting} {lookingUp} interruptHotkeyEnabled={interruptHotkeyEnabled} interruptHotkeyChord={interruptHotkeyChord} interruptHotkeyError={interruptHotkeyError} onStart={start} onStop={stop} onInterrupt={interruptTTS} onLookup={lookupLatest} onCopyChat={copyChat} onCopyOverlay={copyOverlay} />
      {/if}
    </div>

  </div>

  <Notifications bind:notices />
</div>
