<script lang="ts">
  import { onMount } from 'svelte'
  import {
    ConfigPath,
    GetOBSStatus,
    GetRunStatus,
    GetSettings,
    PickCommandsFile,
    PickMediaDirectory,
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
  import TestPane from './lib/TestPane.svelte'

  type Page = 'home' | 'config' | 'test'

  let page: Page = 'home'
  let streamId = ''
  let apiKey = ''
  let port = '8080'
  let ttsEnabled = true
  let ttsVoiceName = ''
  let alertsEnabled = true
  let alertsToken = '@'
  let alertsMediaPath = ''
  let alertsCommandsFilePath = ''
  let webhookEnabled = false
  let testMessage = ''
  let configPath = ''
  let status = ''
  let error = ''
  let saving = false
  let starting = false
  let obs: main.OBSStatus | null = null
  let run: main.RunStatus | null = null

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
    try {
      const [s, path] = await Promise.all([GetSettings(), ConfigPath()])
      applyForm(s)
      configPath = path
      await refreshOBS()
      await refreshRun()
    } catch (e) {
      error = String(e)
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
      status = 'Saved.'
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
      status = 'Starting chat…'
    } catch (e) {
      error = String(e)
      await refreshRun()
    } finally {
      starting = false
    }
  }

  async function stop(): Promise<void> {
    error = ''
    await Stop()
    await refreshRun()
    status = 'Stopped.'
  }

  async function copy(url: string): Promise<void> {
    await ClipboardSetText(url)
    status = 'Copied URL.'
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
      status = 'Test message sent.'
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
</script>

<main>
  <nav class="panes" aria-label="Window panes">
    <button class="pane-btn" class:active={page === 'home'} type="button" on:click={() => { page = 'home' }}>Home</button>
    <button class="pane-btn" class:active={page === 'config'} type="button" on:click={() => { page = 'config' }}>Config</button>
    <button class="pane-btn" class:active={page === 'test'} type="button" on:click={() => { page = 'test' }}>Test</button>
  </nav>

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
      {saving}
      {configPath}
      {obs}
      onSave={save}
      onPickYaml={pickYaml}
      onPickMedia={pickMedia}
      onCopyChat={copyChat}
      onCopyOverlay={copyOverlay}
    />
  {:else if page === 'test'}
    <TestPane bind:testMessage {obs} onSend={sendTest} />
  {:else}
    <HomePane {run} {obs} {starting} onStart={start} onStop={stop} onInterrupt={interruptTTS} />
  {/if}

  {#if status}
    <p class="ok">{status}</p>
  {/if}
  {#if error}
    <p class="err">{error}</p>
  {/if}
</main>
