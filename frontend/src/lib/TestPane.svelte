<script lang="ts">
  import type { main } from '../../wailsjs/go/models'

  export let testMessage: string
  export let obs: main.OBSStatus | null
  export let onSend: () => Promise<void>
</script>

<p class="lead">Injects one chat line into the same alerts and TTS path as YouTube. Start is not required.</p>

<section class="card">
  <header class="card-head">
    <h2>Payload</h2>
    <span class="badge {obs && obs.listening ? 'badge-ok' : 'badge-danger'}">{obs && obs.listening ? 'Ready' : 'OBS offline'}</span>
  </header>
  <label class="field">
    Message
    <input autocomplete="off" bind:value={testMessage} placeholder="@jump or hello" spellcheck="false" type="text" />
  </label>
  <p class="hint">Token plus a YAML command name tests an alert. Any other text tests TTS.</p>
  <div class="actions">
    <button class="btn btn-primary" disabled={!obs || !obs.listening || !testMessage.trim()} type="button" on:click={onSend}>
      Send
    </button>
  </div>
  {#if obs && !obs.listening}
    <p class="err">HTTP listener is down{obs.error ? ': ' + obs.error : ''}.</p>
  {/if}
</section>
