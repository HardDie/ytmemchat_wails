<script lang="ts">
  import type { main } from '../../wailsjs/go/models'

  export let run: main.RunStatus | null
  export let obs: main.OBSStatus | null
  export let starting: boolean
  export let onStart: () => Promise<void>
  export let onStop: () => Promise<void>
  export let onInterrupt: () => Promise<void>

  $: youtubeLabel = !run
    ? 'Unknown'
    : run.connecting
      ? 'Connecting…'
      : run.running
        ? (run.usingApiKey ? 'Running (YouTube API key)' : 'Running (no API key)')
        : 'Stopped'
</script>

<h1>ytmemchat</h1>
<p class="lead">This window is setup only. Chat and alerts play in OBS, not here.</p>

<h2>YouTube</h2>
<p class={run && run.running ? 'ok' : 'hint'}>{youtubeLabel}</p>
{#if run && run.error}
  <p class="err">{run.error}</p>
{/if}
<div class="actions">
  <button class="btn" disabled={starting || (run && (run.running || run.connecting))} type="button" on:click={onStart}>
    {run && run.connecting ? 'Connecting…' : 'Start'}
  </button>
  <button class="btn" disabled={!run || (!run.running && !run.connecting)} type="button" on:click={onStop}>
    Stop
  </button>
</div>

<h2>Audio</h2>
<p class="hint">Stops overlay speech in OBS. Silent if nothing is playing. Does not stop meme alerts.</p>
<div class="actions">
  <button class="btn" disabled={!obs || !obs.listening} type="button" on:click={onInterrupt}>
    Interrupt
  </button>
</div>
{#if obs && !obs.listening}
  <p class="err">OBS HTTP is not listening{obs.error ? ': ' + obs.error : ''}.</p>
{/if}
