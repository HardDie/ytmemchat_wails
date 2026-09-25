<script lang="ts">
  import { onDestroy, onMount } from 'svelte'
  import { GetTTSVoices } from '../../wailsjs/go/configuration/Configuration.js'
  import type { configuration } from '../../wailsjs/go/models'

  export let streamId: string
  export let apiKey: string
  export let port: string
  export let ttsEnabled: boolean
  export let ttsVoiceName: string
  export let alertsEnabled: boolean
  export let alertsToken: string
  export let alertsMediaPath: string
  export let alertsCommandsFilePath: string
  export let alertsCommandsFileCustom: boolean
  export let webhookEnabled: boolean
  export let interruptHotkeyEnabled: boolean
  export let interruptHotkeyChord: string
  export let interruptHotkeyError: string
  export let apiKeyInKeychain: boolean
  export let saving: boolean
  export let configPath: string
  export let onSave: () => Promise<void>
  export let onLookup: () => Promise<void>
  export let lookingUp: boolean
  export let onPickYaml: () => Promise<void>
  export let onPickMedia: () => Promise<void>

  let voices: configuration.TTSVoice[] = []
  let recording = false

  $: selected = voices.find((v) => v.name === ttsVoiceName)
  $: hasApiKey = apiKey.trim() !== ''
  $: lookupDisabled = lookingUp || !hasApiKey

  function voiceLabel(v: configuration.TTSVoice): string {
    const bits = [v.name]
    if (v.languages) {
      bits.push(v.languages)
    }
    if (v.gender) {
      bits.push(v.gender)
    }
    return bits.join(' — ')
  }

  function chordFromEvent(e: KeyboardEvent): string | null {
    if (['Control', 'Shift', 'Alt', 'Meta'].includes(e.key)) {
      return null
    }
    const key = keyName(e)
    if (!key) {
      return null
    }
    const parts: string[] = []
    if (e.ctrlKey) {
      parts.push('Ctrl')
    }
    if (e.metaKey) {
      parts.push('Cmd')
    }
    if (e.altKey) {
      parts.push('Alt')
    }
    if (e.shiftKey) {
      parts.push('Shift')
    }
    if (parts.length === 0) {
      return null
    }
    parts.push(key)
    return parts.join('+')
  }

  function keyName(e: KeyboardEvent): string {
    const c = e.code || ''
    if (c.startsWith('Key') && c.length === 4) {
      return c.slice(3)
    }
    if (c.startsWith('Digit') && c.length === 6) {
      return c.slice(5)
    }
    if (/^F([1-9]|1[012])$/.test(c)) {
      return c
    }
    const named: Record<string, string> = {
      Space: 'Space',
      Escape: 'Escape',
      Tab: 'Tab',
      Enter: 'Enter',
      Backspace: 'Delete',
      Delete: 'Delete',
      ArrowLeft: 'Left',
      ArrowRight: 'Right',
      ArrowUp: 'Up',
      ArrowDown: 'Down',
    }
    return named[c] || named[e.key] || ''
  }

  function stopRecord(): void {
    if (!recording) {
      return
    }
    recording = false
    if (typeof window !== 'undefined') {
      window.removeEventListener('keydown', onRecordKey, true)
    }
  }

  function startRecord(): void {
    if (recording) {
      return
    }
    recording = true
    window.addEventListener('keydown', onRecordKey, true)
  }

  function onRecordKey(e: KeyboardEvent): void {
    if (!recording) {
      return
    }
    e.preventDefault()
    e.stopPropagation()
    if (e.key === 'Escape' && !e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey) {
      stopRecord()
      return
    }
    const chord = chordFromEvent(e)
    if (!chord) {
      return
    }
    interruptHotkeyChord = chord
    stopRecord()
  }

  onDestroy(stopRecord)

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
  <p class="hint">Optional. Empty uses the no-key client. An invalid key does not fall back.
    {#if apiKeyInKeychain}
      API key stored in the OS keychain.
    {:else}
      API key stored in the settings file (keychain unavailable).
    {/if}
  </p>

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
      Media folder
      <span class="path-row">
        <input autocomplete="off" bind:value={alertsMediaPath} spellcheck="false" type="text" />
        <button class="btn btn-small" type="button" on:click={onPickMedia}>Browse</button>
      </span>
    </label>
    <p class="hint">Files are served at <code>/obs/media/</code>. <code>commands.yaml</code> in this folder is created when you save the first command.</p>
    <label class="toggle">
      Use custom commands.yaml path
      <input bind:checked={alertsCommandsFileCustom} type="checkbox" />
    </label>
    {#if alertsCommandsFileCustom}
      <label class="field">
        commands.yaml
        <span class="path-row">
          <input autocomplete="off" bind:value={alertsCommandsFilePath} spellcheck="false" type="text" />
          <button class="btn btn-small" type="button" on:click={onPickYaml}>Browse</button>
        </span>
      </label>
      <p class="hint">Edit names, files, volume, and scale in the Commands pane after this path is saved. A missing custom file fails Start.</p>
    {/if}
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

<section class="card" class:card-compact={!interruptHotkeyEnabled}>
  <header class="card-head">
    <h2>Interrupt shortcut</h2>
    <label class="toggle toggle-head">
      {interruptHotkeyEnabled ? 'On' : 'Off'}
      <input bind:checked={interruptHotkeyEnabled} type="checkbox" />
    </label>
  </header>
  {#if interruptHotkeyEnabled}
    <label class="field">
      Key combination
      <span class="path-row">
        <input
          readonly
          class:recording
          autocomplete="off"
          spellcheck="false"
          type="text"
          value={recording ? 'Press a shortcut…' : interruptHotkeyChord}
        />
        <button class="btn btn-small" type="button" on:click={startRecord}>Set</button>
        <button class="btn btn-small" type="button" on:click={() => { stopRecord(); interruptHotkeyChord = 'Ctrl+Shift+I' }}>Default</button>
      </span>
    </label>
    <p class="hint">
      Works while OBS is fullscreen and this window is in the background. Needs at least one modifier (Ctrl, Cmd, Alt, or Shift). Escape cancels recording. Save to apply. Linux needs X11 (not pure Wayland).
    </p>
    {#if interruptHotkeyError}
      <p class="err">Could not register shortcut: {interruptHotkeyError}</p>
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
