<script lang="ts">
  import type { home } from '../../wailsjs/go/models'

  export let testMessage: string
  export let obs: home.OBSStatus | null
  export let onSend: () => Promise<void>

  $: canSend = !!(obs && obs.listening && testMessage.trim())

  function submit(): void {
    if (!canSend) {
      return
    }
    void onSend()
  }
</script>

<p class="lead">Injects one chat line into the same alerts and TTS path as YouTube. Start is not required.</p>

<section class="card">
  <header class="card-head">
    <h2>Payload</h2>
    <span class="badge {obs && obs.listening ? 'badge-ok' : 'badge-danger'}">{obs && obs.listening ? 'Ready' : 'OBS offline'}</span>
  </header>
  <form on:submit|preventDefault={submit}>
    <label class="field">
      Message
      <input autocomplete="off" bind:value={testMessage} placeholder="@jump or hello" spellcheck="false" type="text" />
    </label>
    <p class="hint">Token plus a YAML command name tests an alert. Any other text tests TTS. Press Enter to send.</p>
    <div class="actions">
      <button class="btn btn-primary" disabled={!canSend} type="submit">
        Send
      </button>
    </div>
  </form>
  {#if obs && !obs.listening}
    <p class="err">HTTP listener is down{obs.error ? ': ' + obs.error : ''}.</p>
  {/if}
</section>
