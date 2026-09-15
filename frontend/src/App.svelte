<script lang="ts">
  import { onMount } from 'svelte'
  import { ConfigPath, GetOBSStatus, GetRunStatus, GetSettings, SaveSettings, Start, Stop } from '../wailsjs/go/main/App.js'
  import { main } from '../wailsjs/go/models'
  import { ClipboardSetText, EventsOn } from '../wailsjs/runtime/runtime'

  let streamId = ''
  let apiKey = ''
  let port = '8080'
  let configPath = ''
  let status = ''
  let error = ''
  let saving = false
  let starting = false
  let obs: main.OBSStatus | null = null
  let run: main.RunStatus | null = null

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
      streamId = s.streamId ?? ''
      apiKey = s.apiKey ?? ''
      port = s.port || '8080'
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
      await SaveSettings(main.SettingsForm.createFrom({ streamId, apiKey, port }))
      const s = await GetSettings()
      streamId = s.streamId ?? ''
      apiKey = s.apiKey ?? ''
      port = s.port || '8080'
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
</script>

<main>
  <h1>ytmemchat</h1>
  <p class="lead">Settings for this machine. OBS pages are served while this window is open. Start pulls live chat into the chat overlay and fans matching commands or TTS onto the overlay.</p>

  <label>
    Stream / video ID
    <input autocomplete="off" bind:value={streamId} spellcheck="false" type="text" />
  </label>
  <p class="hint">The <code>v=</code> value from the YouTube watch URL. Required to Start.</p>

  <label>
    YouTube API key (optional)
    <input autocomplete="off" bind:value={apiKey} spellcheck="false" type="password" />
  </label>
  <p class="hint">Leave empty to use the no-key live chat client. A wrong key does not fall back.</p>

  <label>
    HTTP port
    <input autocomplete="off" bind:value={port} spellcheck="false" type="text" />
  </label>
  <p class="hint">Changing the port restarts the OBS listener after a successful save.</p>

  <div class="actions">
    <button class="btn" disabled={saving} type="button" on:click={save}>
      {saving ? 'Saving…' : 'Save'}
    </button>
    <button class="btn" disabled={starting || (run && (run.running || run.connecting))} type="button" on:click={start}>
      {run && run.connecting ? 'Connecting…' : 'Start'}
    </button>
    <button class="btn" disabled={!run || (!run.running && !run.connecting)} type="button" on:click={stop}>
      Stop
    </button>
  </div>

  {#if run}
    {#if run.running}
      <p class="ok">Chat running{run.usingApiKey ? ' (YouTube API key)' : ' (no API key)'}.</p>
    {:else if run.connecting}
      <p class="ok">Connecting to YouTube chat…</p>
    {/if}
    {#if run.error}
      <p class="err">{run.error}</p>
    {/if}
  {/if}

  {#if obs}
    <section class="obs">
      <h2>OBS Browser Sources</h2>
      {#if obs.listening}
        <p class="ok">HTTP listening</p>
      {:else}
        <p class="err">HTTP not listening{obs.error ? ': ' + obs.error : ''}</p>
      {/if}
      <p class="url-row">
        <span>Chat</span>
        <code>{obs.chatUrl}</code>
        <button class="btn btn-small" type="button" on:click={copyChat}>Copy</button>
      </p>
      <p class="url-row">
        <span>Overlay</span>
        <code>{obs.overlayUrl}</code>
        <button class="btn btn-small" type="button" on:click={copyOverlay}>Copy</button>
      </p>
      <p class="hint">Index (not for OBS): <code>{obs.indexUrl}</code></p>
    </section>
  {/if}

  {#if status}
    <p class="ok">{status}</p>
  {/if}
  {#if error}
    <p class="err">{error}</p>
  {/if}
  {#if configPath}
    <p class="path">File: <code>{configPath}</code></p>
  {/if}
</main>
