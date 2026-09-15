<script lang="ts">
  import { onMount } from 'svelte'
  import { GetTTSVoices } from '../../wailsjs/go/main/App.js'
  import type { main } from '../../wailsjs/go/models'

  export let streamId: string
  export let apiKey: string
  export let port: string
  export let ttsEnabled: boolean
  export let ttsVoiceName: string
  export let alertsEnabled: boolean
  export let alertsToken: string
  export let alertsMediaPath: string
  export let alertsCommandsFilePath: string
  export let webhookEnabled: boolean
  export let saving: boolean
  export let configPath: string
  export let obs: main.OBSStatus | null

  export let onSave: () => Promise<void>
  export let onPickYaml: () => Promise<void>
  export let onPickMedia: () => Promise<void>
  export let onCopyChat: () => void
  export let onCopyOverlay: () => void

  let voices: main.TTSVoice[] = []

  $: selected = voices.find((v) => v.name === ttsVoiceName)

  function voiceLabel(v: main.TTSVoice): string {
    const bits = [v.name]
    if (v.languages) {
      bits.push(v.languages)
    }
    if (v.gender) {
      bits.push(v.gender)
    }
    return bits.join(' — ')
  }

  onMount(async () => {
    try {
      voices = (await GetTTSVoices()) ?? []
    } catch {
      voices = []
    }
  })
</script>

<h1>Config</h1>
<p class="lead">Saved on this machine. OBS Browser Sources use the URLs below.</p>

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

<h2>Alerts</h2>
<label class="toggle">
  <input bind:checked={alertsEnabled} type="checkbox" />
  Enable alerts
</label>
<label>
  Command token
  <input autocomplete="off" bind:value={alertsToken} maxlength="1" spellcheck="false" type="text" />
</label>
<p class="hint">One character, for example <code>@</code>.</p>
<label>
  commands.yaml
  <span class="path-row">
    <input autocomplete="off" bind:value={alertsCommandsFilePath} spellcheck="false" type="text" />
    <button class="btn btn-small" type="button" on:click={onPickYaml}>Browse</button>
  </span>
</label>
<label>
  Media folder
  <span class="path-row">
    <input autocomplete="off" bind:value={alertsMediaPath} spellcheck="false" type="text" />
    <button class="btn btn-small" type="button" on:click={onPickMedia}>Browse</button>
  </span>
</label>
<p class="hint">Served at <code>/obs/media/</code>. Empty commands path skips matching; a bad file fails Start.</p>

<h2>Text to speech</h2>
<label class="toggle">
  <input bind:checked={ttsEnabled} type="checkbox" />
  Speak non-alert messages
</label>
<label>
  Voice
  {#if voices.length > 0}
    <select bind:value={ttsVoiceName}>
      <option value="">System default</option>
      {#if ttsVoiceName && !voices.some((v) => v.name === ttsVoiceName)}
        <option value={ttsVoiceName}>{ttsVoiceName}</option>
      {/if}
      {#each voices as v}
        <option value={v.name}>{voiceLabel(v)}</option>
      {/each}
    </select>
  {:else}
    <input autocomplete="off" bind:value={ttsVoiceName} placeholder="System default" spellcheck="false" type="text" />
  {/if}
</label>
{#if selected}
  <p class="hint voice-meta">
    {#if selected.languages}
      <span>Languages: {selected.languages}</span>
    {/if}
    {#if selected.gender}
      <span>Gender: {selected.gender}</span>
    {/if}
    {#if selected.details}
      <span class="voice-details">{selected.details}</span>
    {/if}
  </p>
{:else}
  <p class="hint">Empty voice uses the OS default.</p>
{/if}

<h2>HTTP API</h2>
<label class="toggle">
  <input bind:checked={webhookEnabled} type="checkbox" />
  Enable POST /api/webhook and /api/interrupt
</label>
<p class="hint">For external automation. The Test pane does not need this.</p>

<div class="actions">
  <button class="btn" disabled={saving} type="button" on:click={onSave}>
    {saving ? 'Saving…' : 'Save'}
  </button>
</div>

{#if obs}
  <h2>OBS Browser Sources</h2>
  {#if obs.listening}
    <p class="ok">HTTP listening</p>
  {:else}
    <p class="err">HTTP not listening{obs.error ? ': ' + obs.error : ''}</p>
  {/if}
  <p class="url-row">
    <span>Chat</span>
    <code>{obs.chatUrl}</code>
    <button class="btn btn-small" type="button" on:click={onCopyChat}>Copy</button>
  </p>
  <p class="url-row">
    <span>Overlay</span>
    <code>{obs.overlayUrl}</code>
    <button class="btn btn-small" type="button" on:click={onCopyOverlay}>Copy</button>
  </p>
  <p class="hint">Index (not for OBS): <code>{obs.indexUrl}</code></p>
{/if}
{#if configPath}
  <p class="path">File: <code>{configPath}</code></p>
{/if}
