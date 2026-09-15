<script lang="ts">
  import type { main } from '../../wailsjs/go/models'

  export let testMessage: string
  export let obs: main.OBSStatus | null
  export let onSend: () => Promise<void>
</script>

<h1>Test message</h1>
<p class="lead">Sends a fake chat line through alerts and TTS, same as live chat. YouTube Start is not required.</p>

<label>
  Message
  <input autocomplete="off" bind:value={testMessage} placeholder="@jump or hello" spellcheck="false" type="text" />
</label>
<p class="hint">Use the command token plus a YAML name to test an alert, or plain text to test TTS.</p>

<div class="actions">
  <button class="btn" disabled={!obs || !obs.listening || !testMessage.trim()} type="button" on:click={onSend}>
    Send
  </button>
</div>
{#if obs && !obs.listening}
  <p class="err">OBS HTTP is not listening{obs.error ? ': ' + obs.error : ''}.</p>
{/if}
