<script lang="ts">
  import { onMount } from 'svelte'
  import {
    AppVersion,
    ConfigPath,
    GetOBSStatus,
    GetRunStatus,
    GetSettings,
    GetAlertCommands,
    SaveAlertCommands,
    LookupLatestStream,
    PickCommandsFile,
    PickMediaDirectory,
    PickAlertMediaFile,
    PreviewAlert,
    SaveSettings,
    SendTestMessage,
    InterruptTTS,
    Start,
    Stop,
  } from '../wailsjs/go/main/App.js'
  import { main } from '../wailsjs/go/models'
  import { ClipboardSetText, EventsOn } from '../wailsjs/runtime/runtime'
  import HomePane from './lib/HomePane.svelte'
  import ConfigPane from './lib/ConfigPane.svelte'
  import CommandsPane from './lib/CommandsPane.svelte'
  import TestPane from './lib/TestPane.svelte'

  type Page = 'home' | 'config' | 'commands' | 'test'

  const titles: Record<Page, string> = {
    home: 'Home',
    config: 'Configuration',
    commands: 'Commands',
    test: 'Test message',
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
  let webhookEnabled = false
  let interruptHotkeyEnabled = true
  let interruptHotkeyChord = 'Ctrl+Shift+I'
  let interruptHotkeyError = ''
  let testMessage = ''
  let configPath = ''
  let status = ''
  let error = ''
  let saving = false
  let starting = false
  let lookingUp = false
  let commandsSaving = false
  let commandsLoading = false
  let commandsPath = ''
  let commandRows: Array<{ name: string; file: string; volume: string; scale: string }> = []
  let obs: main.OBSStatus | null = null
  let run: main.RunStatus | null = null
  let appVersion = ''

  function applyForm(s: main.SettingsForm): void {
    streamId = s.streamId ?? ''
    apiKey = s.apiKey ?? ''
    port = s.port || '8080'
    ttsEnabled = !!s.ttsEnabled
    ttsVoiceName = s.ttsVoiceName ?? ''
    alertsEnabled = !!s.alertsEnabled
    alertsToken = s.alertsToken || '@'
    alertsMediaPath = s.alertsMediaPath ?? ''
    alertsCommandsFilePath = s.alertsCommandsFilePath ?? ''
    webhookEnabled = !!s.webhookEnabled
    interruptHotkeyEnabled = s.interruptHotkeyEnabled !== false
    interruptHotkeyChord = s.interruptHotkeyChord || 'Ctrl+Shift+I'
    interruptHotkeyError = s.interruptHotkeyError ?? ''
  }

  function formPayload(): main.SettingsForm {
    return main.SettingsForm.createFrom({
      streamId,
      apiKey,
      port,
      ttsEnabled,
      ttsVoiceName,
      alertsEnabled,
      alertsToken,
      alertsMediaPath,
      alertsCommandsFilePath,
      webhookEnabled,
      interruptHotkeyEnabled,
      interruptHotkeyChord,
    })
  }

  async function refreshOBS(): Promise<void> {
    obs = await GetOBSStatus()
  }

  async function refreshRun(): Promise<void> {
    run = await GetRunStatus()
  }

  onMount(async () => {
    EventsOn('pipeline', (s: main.RunStatus) => {
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
      error = String(e)
    }
    return () => {
      window.clearInterval(quotaTick)
    }
  })

  async function save(): Promise<void> {
    saving = true
    error = ''
    status = ''
    try {
      await SaveSettings(formPayload())
      applyForm(await GetSettings())
      await refreshOBS()
      await refreshRun()
      status = 'Settings saved'
    } catch (e) {
      error = String(e)
      try {
        await refreshOBS()
      } catch {
        /* keep save error */
      }
    } finally {
      saving = false
    }
  }

  async function start(): Promise<void> {
    starting = true
    error = ''
    status = ''
    try {
      await save()
      if (error) {
        return
      }
      await Start()
      await refreshRun()
      status = ''
    } catch (e) {
      error = String(e)
      await refreshRun()
    } finally {
      starting = false
    }
  }

  async function stop(): Promise<void> {
    error = ''
    status = ''
    await Stop()
    await refreshRun()
  }

  async function copy(url: string): Promise<void> {
    await ClipboardSetText(url)
    status = 'URL copied'
    error = ''
  }

  function copyChat(): void {
    if (obs) copy(obs.chatUrl)
  }

  function copyOverlay(): void {
    if (obs) copy(obs.overlayUrl)
  }

  async function sendTest(): Promise<void> {
    error = ''
    status = ''
    try {
      await SendTestMessage(testMessage)
      status = 'Test message sent'
    } catch (e) {
      error = String(e)
    }
  }

  async function interruptTTS(): Promise<void> {
    error = ''
    status = ''
    try {
      await InterruptTTS()
    } catch (e) {
      error = String(e)
    }
  }

  async function lookupLatest(): Promise<void> {
    lookingUp = true
    error = ''
    status = ''
    try {
      const got = await LookupLatestStream(streamId, apiKey)
      if (got && got.streamId) {
        streamId = got.streamId
        const kind = got.kind === 'upcoming' ? 'upcoming' : 'live'
        status = `Latest ${kind} stream: ${got.streamId}`
      }
      await refreshRun()
    } catch (e) {
      error = String(e)
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
      await save()
    } catch (e) {
      error = String(e)
    }
  }

  async function testCommand(index: number): Promise<void> {
    error = ''
    status = ''
    const row = commandRows[index]
    if (!row || !row.file.trim()) {
      error = 'Command needs a file'
      return
    }
    try {
      const volume = row.volume.trim() ? Number(row.volume) : 1
      const scale = row.scale.trim() ? Number(row.scale) : 1
      await PreviewAlert(row.file.trim(), volume, scale)
      status = 'Alert sent to overlay'
    } catch (e) {
      error = String(e)
    }
  }

  async function pickCommandFile(index: number): Promise<void> {
    error = ''
    status = ''
    try {
      const p = await PickAlertMediaFile(alertsMediaPath)
      if (!p) {
        return
      }
      commandRows = commandRows.map((row, i) => (i === index ? { ...row, file: p } : row))
    } catch (e) {
      error = String(e)
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
      error = String(e)
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
    error = ''
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
      error = String(e)
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
    error = ''
    status = ''
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
      await SaveAlertCommands(main.AlertCommandsFile.createFrom({ path: commandsPath, commands }))
      await loadCommands()
      status = 'Commands saved'
    } catch (e) {
      error = String(e)
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
          bind:webhookEnabled
          bind:interruptHotkeyEnabled
          bind:interruptHotkeyChord
          interruptHotkeyError={interruptHotkeyError}
          {saving}
          {configPath}
          onSave={save}
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
        />
      {:else if page === 'test'}
        <TestPane bind:testMessage {obs} onSend={sendTest} />
      {:else}
        <HomePane {run} {obs} {streamId} {apiKey} {starting} {lookingUp} interruptHotkeyEnabled={interruptHotkeyEnabled} interruptHotkeyChord={interruptHotkeyChord} interruptHotkeyError={interruptHotkeyError} onStart={start} onStop={stop} onInterrupt={interruptTTS} onLookup={lookupLatest} onCopyChat={copyChat} onCopyOverlay={copyOverlay} />
      {/if}
    </div>

    {#if status || error}
      <div class="flash">
        {#if status}
          <p class="ok">{status}</p>
        {/if}
        {#if error}
          <p class="err">{error}</p>
        {/if}
      </div>
    {/if}
  </div>
</div>
