<script lang="ts">
  import type { update } from '../../wailsjs/go/models'

  export let current: string
  export let status: update.Status | null
  export let busy: boolean
  export let downloaded: boolean
  export let onCheck: () => Promise<void>
  export let onDownload: () => Promise<void>
  export let onApply: () => Promise<void>
  export let onOpen: () => void

  $: latest = status ? status.latest : ''
  $: canDownload = !!(status && status.asset) && !busy
  $: canApply = downloaded && !!(status && status.canInstall) && !busy
</script>

<p class="lead">Checks GitHub Releases for a tagged build. Download verifies SHA-256. Install quits this window and replaces the app on disk.</p>

<section class="card">
  <header class="card-head">
    <h2>Version</h2>
    <span class="badge">{current || 'unknown'}</span>
  </header>
  {#if status}
    <p class="hint">Latest release: <strong>{latest}</strong></p>
    {#if status.same}
      <p class="ok">This binary matches the latest GitHub tag.</p>
    {:else if status.newer}
      <p class="hint">A newer tag is available.</p>
    {:else}
      <p class="hint">This build is not a release tag (dev or a commit). You can still download the latest archive.</p>
    {/if}
    {#if status.asset}
      <p class="hint">Archive: {status.asset}</p>
    {:else}
      <p class="err">No archive for this OS in that release.</p>
    {/if}
    {#if status.notes}
      <pre class="notes">{status.notes}</pre>
    {/if}
  {:else}
    <p class="hint">Click Check to query GitHub. Nothing is downloaded until you ask.</p>
  {/if}
  {#if downloaded}
    <p class="ok">Archive verified. Quit and install when you are ready (OBS chat will stop with the app).</p>
  {/if}
  {#if status && status.asset && !status.canInstall}
    <p class="err">This folder is not writable. Download the zip from GitHub and replace the app yourself.</p>
  {/if}
  <div class="actions">
    <button class="btn btn-primary" disabled={busy} type="button" on:click={() => void onCheck()}>
      {busy ? 'Working…' : 'Check'}
    </button>
    <button class="btn" disabled={!canDownload} type="button" on:click={() => void onDownload()}>
      Download
    </button>
    <button class="btn" disabled={!canApply} type="button" on:click={() => void onApply()}>
      Quit and install
    </button>
    <button class="btn" type="button" on:click={onOpen}>
      Open GitHub
    </button>
  </div>
</section>
