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
  export let onSave: () => Promise<void>
  export let onLookup: () => Promise<void>
  export let lookingUp: boolean
  export let onPickYaml: () => Promise<void>
  export let onPickMedia: () => Promise<void>

  let voices: main.TTSVoice[] = []

  $: selected = voices.find((v) => v.name === ttsVoiceName)
  $: hasApiKey = apiKey.trim() !== ''
  $: lookupDisabled = lookingUp || !hasApiKey

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

<p class="lead">Stored on this machine. Chat and overlay URLs are on Home.</p>

<section class="card">
  <header class="card-head">
    <h2>Connection</h2>
  </header>
  <label class="field">
    Stream / video ID
    <span class="path-row">
      <input autocomplete="off" bind:value={streamId} spellcheck="false" type="text" />
      <button
        class="btn btn-small"
        disabled={lookupDisabled}
        title={hasApiKey ? 'Find the current live or upcoming stream' : 'A YouTube API key is required'}
        type="button"
        on:click={onLookup}
      >
        {lookingUp ? 'Finding…' : 'Find latest'}
      </button>
    </span>
  </label>
  <p class="hint">
    Paste any recent video from the channel once. Find latest loads the current live stream, or the next upcoming one if nothing is live. VODs are skipped. Save to persist.
    {#if !hasApiKey}
      A YouTube API key is required for Find latest.
    {/if}
  </p>

  <label class="field">
    YouTube API key
    <input autocomplete="off" bind:value={apiKey} spellcheck="false" type="password" />
  </label>
  <p class="hint">Optional. Empty uses the no-key client. An invalid key does not fall back.</p>

  <label class="field">
    HTTP port
    <input autocomplete="off" bind:value={port} spellcheck="false" type="text" />
  </label>
  <p class="hint">Saving a new port restarts the local OBS listener.</p>
</section>

<section class="card" class:card-compact={!alertsEnabled}>
  <header class="card-head">
    <h2>Alerts</h2>
    <label class="toggle toggle-head">
      {alertsEnabled ? 'On' : 'Off'}
      <input bind:checked={alertsEnabled} type="checkbox" />
    </label>
  </header>
  {#if alertsEnabled}
    <label class="field">
      Command token
      <input autocomplete="off" bind:value={alertsToken} maxlength="1" spellcheck="false" type="text" />
    </label>
    <p class="hint">Single character, for example <code>@</code>.</p>
    <label class="field">
      commands.yaml
      <span class="path-row">
        <input autocomplete="off" bind:value={alertsCommandsFilePath} spellcheck="false" type="text" />
        <button class="btn btn-small" type="button" on:click={onPickYaml}>Browse</button>
      </span>
    </label>
    <label class="field">
      Media folder
      <span class="path-row">
        <input autocomplete="off" bind:value={alertsMediaPath} spellcheck="false" type="text" />
        <button class="btn btn-small" type="button" on:click={onPickMedia}>Browse</button>
      </span>
    </label>
    <p class="hint">Files are served at <code>/obs/media/</code>. An empty YAML path skips matching; a bad file fails Start.</p>
  {/if}
</section>

<section class="card" class:card-compact={!ttsEnabled}>
  <header class="card-head">
    <h2>Text to speech</h2>
    <label class="toggle toggle-head">
      {ttsEnabled ? 'On' : 'Off'}
      <input bind:checked={ttsEnabled} type="checkbox" />
    </label>
  </header>
  {#if ttsEnabled}
    <label class="field">
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
          <span>{selected.languages}</span>
        {/if}
        {#if selected.gender}
          <span>{selected.gender}</span>
        {/if}
        {#if selected.details}
          <span class="voice-details">{selected.details}</span>
        {/if}
      </p>
    {:else}
      <p class="hint">Leave empty for the operating-system default voice.</p>
    {/if}
  {/if}
</section>

<section class="card">
  <header class="card-head">
    <h2>HTTP API</h2>
  </header>
  <label class="toggle">
    Enable /api/webhook and /api/interrupt
    <input bind:checked={webhookEnabled} type="checkbox" />
  </label>
  <p class="hint">For external automation. The Test pane does not require this.</p>
</section>

<div class="actions">
  <button class="btn btn-primary" disabled={saving} type="button" on:click={onSave}>
    {saving ? 'Saving…' : 'Save changes'}
  </button>
</div>

{#if configPath}
  <p class="path">Settings file <code>{configPath}</code></p>
{/if}
