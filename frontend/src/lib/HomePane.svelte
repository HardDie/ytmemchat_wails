<script lang="ts">
  import type { main } from '../../wailsjs/go/models'

  export let run: main.RunStatus | null
  export let obs: main.OBSStatus | null
  export let streamId: string
  export let apiKey: string
  export let starting: boolean
  export let lookingUp: boolean
  export let onStart: () => Promise<void>
  export let onStop: () => Promise<void>
  export let onInterrupt: () => Promise<void>
  export let onLookup: () => Promise<void>
  export let onCopyChat: () => void
  export let onCopyOverlay: () => void
  export let interruptHotkeyEnabled: boolean
  export let interruptHotkeyChord: string
  export let interruptHotkeyError: string

  $: hasApiKey = apiKey.trim() !== ''
  $: lookupDisabled = lookingUp || !hasApiKey
  $: lookupHint = hasApiKey
    ? 'Finds the current live stream, or the next upcoming one. Skips VODs.'
    : 'Find latest requires a YouTube API key (set it in Configuration).'

  $: youtubeLabel = !run
    ? 'Unknown'
    : run.connecting
      ? 'Connecting'
      : run.running
        ? 'Connected'
        : 'Stopped'

  $: youtubeBadge = !run
    ? ''
    : run.connecting
      ? 'badge-warn'
      : run.running
        ? 'badge-ok'
        : ''

  $: youtubeDetail = run && run.running
    ? (run.usingApiKey ? 'YouTube Data API v3' : 'No API key')
    : 'Start uses the stream ID saved in Configuration.'

  $: quotaUnits = run && run.quotaUnits ? run.quotaUnits : 0
  $: quotaUnitsLimit = run && run.quotaUnitsLimit ? run.quotaUnitsLimit : 10000
  $: quotaSearch = run && run.quotaSearch ? run.quotaSearch : 0
  $: quotaSearchLimit = run && run.quotaSearchLimit ? run.quotaSearchLimit : 100
  $: quotaRemaining = Math.max(0, quotaUnitsLimit - quotaUnits)
  $: quotaClass = quotaRemaining === 0 ? 'quota-danger' : quotaRemaining < 3600 ? 'quota-warn' : ''

  function fmtQuota(n: number): string {
    return n.toLocaleString('en-US')
  }
</script>

<section class="card">
  <header class="card-head">
    <h2>YouTube live chat</h2>
    <div class="status-cluster">
      <code class="stream-id" title={streamId || 'Set a stream ID in Configuration'}>{streamId || 'No stream ID'}</code>
      <span class="badge {youtubeBadge}">{youtubeLabel}</span>
    </div>
  </header>
  <p class="hint">{youtubeDetail}</p>
  {#if hasApiKey}
    <p class="hint quota {quotaClass}">
      Spent today: {fmtQuota(quotaUnits)} / {fmtQuota(quotaUnitsLimit)} units{#if quotaSearch > 0} · search {fmtQuota(quotaSearch)} / {fmtQuota(quotaSearchLimit)}{/if}
    </p>
    <p class="hint">This app’s Data API calls today. liveChatMessages.list costs 5 units. Resets at midnight Pacific Time. Other tools on the same Google Cloud project are not included.</p>
  {/if}
  {#if run && run.error}
    <p class="err">{run.error}</p>
  {/if}
  <div class="actions">
    <button class="btn btn-primary" disabled={starting || (run && (run.running || run.connecting))} type="button" on:click={onStart}>
      {run && run.connecting ? 'Connecting…' : 'Start'}
    </button>
    <button
      class="btn"
      disabled={lookupDisabled}
      title={hasApiKey ? 'Find the current live or upcoming stream' : 'A YouTube API key is required'}
      type="button"
      on:click={onLookup}
    >
      {lookingUp ? 'Finding…' : 'Find latest'}
    </button>
    <button class="btn" disabled={!run || (!run.running && !run.connecting)} type="button" on:click={onStop}>
      Stop
    </button>
  </div>
  <p class="hint">{lookupHint}</p>
</section>

<section class="card">
  <header class="card-head">
    <h2>Overlay speech</h2>
    <span class="badge {obs && obs.listening ? 'badge-ok' : 'badge-danger'}">{obs && obs.listening ? 'OBS ready' : 'OBS offline'}</span>
  </header>
  <p class="hint">Stops current and queued TTS in the OBS overlay. Does not stop meme alerts.</p>
  <div class="actions">
    <button class="btn" disabled={!obs || !obs.listening} type="button" on:click={onInterrupt}>
      Interrupt speech
    </button>
  </div>
  {#if interruptHotkeyEnabled && interruptHotkeyChord}
    <p class="hint">Global shortcut <kbd>{interruptHotkeyChord}</kbd> works even when OBS is fullscreen.</p>
  {/if}
  {#if interruptHotkeyError}
    <p class="err">Shortcut not active: {interruptHotkeyError}</p>
  {/if}
  {#if obs && !obs.listening}
    <p class="err">HTTP listener is down{obs.error ? ': ' + obs.error : ''}.</p>
  {/if}
</section>

{#if obs}
  <section class="card">
    <header class="card-head">
      <h2>OBS Browser Sources</h2>
      <span class="badge {obs.listening ? 'badge-ok' : 'badge-danger'}">{obs.listening ? 'Listening' : 'Offline'}</span>
    </header>
    {#if !obs.listening && obs.error}
      <p class="err">{obs.error}</p>
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
    <p class="hint">Do not use the index URL as an OBS source: <code>{obs.indexUrl}</code></p>
  </section>
{/if}
