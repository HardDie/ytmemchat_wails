<script lang="ts">
  export let token: string
  export let commandsPath: string
  export let mediaPath: string
  export let saving: boolean

  export let onBack: () => void
  export let onSave: () => Promise<void>
  export let onPickYaml: () => Promise<void>
  export let onPickMedia: () => Promise<void>
</script>

<div class="page-head">
  <button class="btn btn-small" type="button" on:click={onBack}>Back</button>
  <h1>Alerts</h1>
</div>
<p class="lead">Command token and files used when alerts are enabled on the home page.</p>

<label>
  Command token
  <input autocomplete="off" bind:value={token} maxlength="1" spellcheck="false" type="text" />
</label>
<p class="hint">One character, for example <code>@</code>. Chat like <code>@jump</code> plays the matching command.</p>

<label>
  commands.yaml
  <span class="path-row">
    <input autocomplete="off" bind:value={commandsPath} spellcheck="false" type="text" />
    <button class="btn btn-small" type="button" on:click={onPickYaml}>Browse</button>
  </span>
</label>
<p class="hint">If set, Start loads this file. A missing or invalid file fails Start. Empty skips matching.</p>

<label>
  Media folder
  <span class="path-row">
    <input autocomplete="off" bind:value={mediaPath} spellcheck="false" type="text" />
    <button class="btn btn-small" type="button" on:click={onPickMedia}>Browse</button>
  </span>
</label>
<p class="hint">Files in this folder are served at <code>/obs/media/</code>.</p>

<div class="actions">
  <button class="btn" disabled={saving} type="button" on:click={onSave}>
    {saving ? 'Saving…' : 'Save'}
  </button>
</div>
