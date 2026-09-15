<script lang="ts">
  import { onMount } from 'svelte'
  import { GetTTSVoices } from '../../wailsjs/go/main/App.js'
  import type { main } from '../../wailsjs/go/models'

  export let voiceName: string
  export let saving: boolean

  export let onBack: () => void
  export let onSave: () => Promise<void>

  let voices: main.TTSVoice[] = []

  $: selected = voices.find((v) => v.name === voiceName)

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

<div class="page-head">
  <button class="btn btn-small" type="button" on:click={onBack}>Back</button>
  <h1>Text to speech</h1>
</div>
<p class="lead">Voice used for chat lines that did not trigger an alert. Turn TTS on or off from the home page.</p>

<label>
  Voice
  {#if voices.length > 0}
    <select bind:value={voiceName}>
      <option value="">System default</option>
      {#if voiceName && !voices.some((v) => v.name === voiceName)}
        <option value={voiceName}>{voiceName}</option>
      {/if}
      {#each voices as v}
        <option value={v.name}>{voiceLabel(v)}</option>
      {/each}
    </select>
  {:else}
    <input autocomplete="off" bind:value={voiceName} placeholder="System default" spellcheck="false" type="text" />
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
  <p class="hint">Leave empty for the OS default. macOS uses <code>say</code>, Windows SAPI, Linux <code>espeak</code>.</p>
{/if}

<div class="actions">
  <button class="btn" disabled={saving} type="button" on:click={onSave}>
    {saving ? 'Saving…' : 'Save'}
  </button>
</div>
