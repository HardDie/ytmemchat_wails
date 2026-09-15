<script lang="ts">
  import { onMount } from 'svelte'
  import { ConfigPath, GetSettings, SaveSettings } from '../wailsjs/go/main/App.js'
  import { main } from '../wailsjs/go/models'

  let streamId = ''
  let apiKey = ''
  let port = '8080'
  let configPath = ''
  let status = ''
  let error = ''
  let saving = false

  onMount(async () => {
    try {
      const [s, path] = await Promise.all([GetSettings(), ConfigPath()])
      streamId = s.streamId ?? ''
      apiKey = s.apiKey ?? ''
      port = s.port || '8080'
      configPath = path
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
      status = 'Saved.'
    } catch (e) {
      error = String(e)
    } finally {
      saving = false
    }
  }
</script>

<main>
  <h1>ytmemchat</h1>
  <p class="lead">Settings for this machine. OBS overlays are not started from this screen yet.</p>

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
  <p class="hint">OBS will use <code>http://127.0.0.1:&lt;port&gt;/obs/chat</code> (default 8080).</p>

  <div class="actions">
    <button class="btn" disabled={saving} type="button" on:click={save}>
      {saving ? 'Saving…' : 'Save'}
    </button>
  </div>

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
