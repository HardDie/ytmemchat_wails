<script lang="ts">
  import DecimalField from './DecimalField.svelte'
  import Icon from './Icon.svelte'
  import type { CommandFields } from './types'

  export let row: CommandFields
  export let duplicate = false
  export let hasMedia = false
  export let confirmDelete = false
  export let canPreview = false
  export let previewTitle = ''
  export let onName: () => void = () => {}
  export let onPick: () => void = () => {}
  export let onPlay: () => void = () => {}
  export let onRequestDelete: () => void = () => {}
  export let onCancelDelete: () => void = () => {}
  export let onConfirmDelete: () => void = () => {}

  function onNameInput(ev: Event): void {
    row.name = (ev.currentTarget as HTMLInputElement).value
    onName()
  }

  $: fileTitle = hasMedia ? 'Choose a file with the folder button' : 'Set a media folder in Configuration'
  $: pickTitle = hasMedia ? 'Choose a file inside the media folder' : 'Set a media folder in Configuration'
  $: deleteLabel = row.name.trim() ? `“${row.name.trim()}”` : 'this command'
  $: canPlay = canPreview && row.file.trim() !== ''
  $: playTitle = previewTitle || (row.file.trim() ? 'Play this command on the OBS overlay' : 'Choose a file first')
</script>

<div class="command-row">
  <div class="command-main">
    <label class="field">
      Name
      <input
        autocomplete="off"
        class:name-dup={duplicate}
        aria-invalid={duplicate}
        value={row.name}
        placeholder="jump"
        spellcheck="false"
        type="text"
        on:input={onNameInput}
      />
      {#if duplicate}
        <span class="name-error">This command already exists</span>
      {/if}
    </label>
    <label class="field">
      File
      <span class="file-in">
        <input
          readonly
          autocomplete="off"
          bind:value={row.file}
          placeholder="Choose a file"
          spellcheck="false"
          title={fileTitle}
          type="text"
        />
        <button
          class="file-in-btn"
          disabled={!hasMedia}
          title={pickTitle}
          type="button"
          aria-label="Choose file"
          on:click={onPick}
        >
          <Icon name="folder" size={14} />
        </button>
      </span>
    </label>
  </div>
  <div class="command-meta">
    <DecimalField label="Volume" bind:value={row.volume} />
    <DecimalField label="Scale" bind:value={row.scale} />
    <div class="command-actions">
      <button
        class="row-icon"
        disabled={!canPlay}
        type="button"
        aria-label="Play"
        title={playTitle}
        on:click={onPlay}
      >
        <Icon name="play" size={13} />
      </button>
      <button
        class="row-icon trash"
        class:open={confirmDelete}
        type="button"
        aria-label={confirmDelete ? 'Cancel delete' : 'Delete'}
        aria-expanded={confirmDelete}
        title={confirmDelete ? 'Cancel delete' : 'Delete'}
        on:click={onRequestDelete}
      >
        <Icon name="trash" size={14} />
      </button>
      {#if confirmDelete}
        <div class="command-confirm" role="dialog" aria-label="Delete command">
          <p>Delete {deleteLabel}?</p>
          <button type="button" on:click={onCancelDelete}>Cancel</button>
          <button class="confirm-danger" type="button" on:click={onConfirmDelete}>Delete</button>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .command-row {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin: 0;
    padding: 16px;
    min-width: 0;
    max-width: 100%;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: rgba(0, 0, 0, 0.18);
  }

  .command-main,
  .command-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    align-items: flex-end;
    min-width: 0;
    width: 100%;
  }

  .command-row :global(.field) {
    margin-bottom: 0;
    min-width: 0;
    flex: 1 1 12rem;
  }

  .command-main .file-in {
    position: relative;
    display: block;
    width: 100%;
    min-width: 0;
  }

  .command-main .file-in input {
    padding-right: 32px;
  }

  .file-in-btn {
    position: absolute;
    top: 0;
    right: 0;
    width: 32px;
    height: 32px;
    margin: 0;
    padding: 0;
    border: none;
    border-radius: 0 6px 6px 0;
    background: transparent;
    color: var(--muted);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .file-in-btn:hover:not(:disabled) {
    color: var(--text);
    background: rgba(255, 255, 255, 0.06);
  }

  .file-in-btn:disabled {
    opacity: 0.35;
    cursor: default;
  }

  .command-meta :global(.field) {
    flex: 0 1 8rem;
    max-width: 10rem;
  }

  .command-row :global(label) {
    min-width: 0;
    max-width: 100%;
  }

  .command-row :global(input) {
    display: block;
    width: 100%;
    min-width: 0;
    max-width: 100%;
  }

  .command-row input.name-dup,
  .command-row input.name-dup:focus {
    border-color: var(--danger);
    box-shadow: 0 0 0 3px rgba(248, 113, 113, 0.28);
  }

  .command-row .name-error {
    color: var(--danger);
    font-size: 11px;
    font-weight: 600;
    line-height: 1.3;
  }

  .command-actions {
    position: relative;
    flex: 0 0 auto;
    align-self: flex-end;
    display: flex;
    gap: 4px;
  }

  .row-icon {
    width: 32px;
    height: 32px;
    margin: 0;
    padding: 0;
    border: 1px solid var(--border-strong);
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.06);
    color: var(--muted);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .row-icon:hover:not(:disabled) {
    color: var(--text);
    background: rgba(255, 255, 255, 0.1);
  }

  .row-icon.trash:hover,
  .row-icon.trash.open {
    color: var(--danger);
    border-color: rgba(248, 113, 113, 0.45);
    background: rgba(248, 113, 113, 0.1);
  }

  .row-icon:disabled {
    opacity: 0.4;
    cursor: default;
  }

  .command-confirm {
    position: absolute;
    right: 0;
    bottom: calc(100% + 4px);
    z-index: 5;
    min-width: 11rem;
    padding: 4px;
    border: 1px solid var(--border-strong);
    border-radius: 6px;
    background: var(--bg-elevated);
    box-shadow: var(--shadow);
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .command-confirm p {
    margin: 0;
    padding: 6px 10px 4px;
    font-size: 12px;
    font-weight: 600;
    color: var(--muted);
  }

  .command-confirm button {
    margin: 0;
    border: none;
    border-radius: 4px;
    background: transparent;
    color: var(--text);
    font: inherit;
    font-size: 12px;
    font-weight: 600;
    text-align: left;
    padding: 8px 10px;
    cursor: pointer;
  }

  .command-confirm button:hover {
    background: rgba(255, 255, 255, 0.06);
  }

  .command-confirm .confirm-danger {
    color: var(--danger);
  }

  @container commands (min-width: 720px) {
    .command-row {
      flex-direction: row;
      flex-wrap: nowrap;
      align-items: flex-end;
    }

    .command-main {
      flex: 1 1 auto;
      flex-wrap: nowrap;
      width: auto;
      min-width: 0;
    }

    .command-meta {
      flex: 0 0 auto;
      flex-wrap: nowrap;
      width: auto;
    }

    .command-meta :global(.field) {
      flex: 0 0 6rem;
      max-width: 6rem;
    }
  }

  @media (min-width: 1100px) {
    .command-row {
      flex-direction: row;
      flex-wrap: nowrap;
      align-items: flex-end;
    }

    .command-main {
      flex: 1 1 auto;
      flex-wrap: nowrap;
      width: auto;
      min-width: 0;
    }

    .command-meta {
      flex: 0 0 auto;
      flex-wrap: nowrap;
      width: auto;
    }

    .command-meta :global(.field) {
      flex: 0 0 6rem;
      max-width: 6rem;
    }
  }
</style>
