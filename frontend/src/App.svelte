<script lang="ts">
  import { onMount } from 'svelte'
  import { ConfigPath, GetOBSStatus, GetSettings, SaveSettings } from '../wailsjs/go/main/App.js'
  import { main } from '../wailsjs/go/models'
  import { ClipboardSetText } from '../wailsjs/runtime/runtime'

  let streamId = ''
  let apiKey = ''
  let port = '8080'
  let configPath = ''
  let status = ''
  let error = ''
  let saving = false
  let obs: main.OBSStatus | null = null

  async function refreshOBS(): Promise<void> {
    obs = await GetOBSStatus()
  }

  onMount(async () => {
    try {
      const [s, path] = await Promise.all([GetSettings(), ConfigPath()])
      streamId = s.streamId ?? ''
      apiKey = s.apiKey ?? ''
      port = s.port || '8080'
      configPath = path
      await refreshOBS()
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
  <p class="lead">Settings for this machine. OBS pages are served while this window is open.</p>

  <label>
    Stream / video ID
    <input autocomplete="off" bind:value={streamId} spellcheck="false" type="text" />
  </label>
  <p class="hint">The <code>v=</code> value from the YouTube watch URL. Required later to Start; you can save without it.</p>

  <label>
    YouTube API key (optional)
    <input autocomplete="off" bind:value={apiKey} spellcheck="false" type="password" />
  </label>
  <p class="hint">Leave empty to use the no-key live chat client. A wrong key does not fall back.</p>

  <label>
    HTTP port
    <input autocomplete="off" bind:value={port} spellcheck="false" type="text" />
  </label>
  <p class="hint">Changing the port restarts the OBS listener after a successful save. Restart the app if you saved while it was not listening.</p>

  <div class="actions">
    <button class="btn" disabled={saving} type="button" on:click={save}>
      {saving ? 'Saving…' : 'Save'}
    </button>
  </div>

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
