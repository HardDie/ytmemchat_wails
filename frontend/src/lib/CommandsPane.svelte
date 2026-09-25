<script lang="ts">
  import { onDestroy, onMount } from 'svelte'

  export let path: string
  export let rows: Array<{ name: string; file: string; volume: string; scale: string }>
  export let saving: boolean
  export let loading: boolean
  export let onAdd: () => void
  export let onRemove: (index: number) => void
  export let onReorder: (next: Array<{ name: string; file: string; volume: string; scale: string }>) => void
  export let onSave: () => Promise<void>
  export let onReload: () => Promise<void>
  export let mediaPath: string
  export let onPickFile: (index: number) => Promise<void>
  export let onTest: (index: number) => Promise<void>
  export let canTest: boolean

  $: hasMedia = mediaPath.trim() !== ''

  let menu = -1
  let confirmDelete = -1

  function closeMenu(): void {
    menu = -1
    confirmDelete = -1
  }

  function toggleMenu(i: number): void {
    if (menu === i) {
      closeMenu()
      return
    }
    menu = i
    confirmDelete = -1
  }

  function requestDelete(i: number): void {
    confirmDelete = i
  }

  function confirmRemove(i: number): void {
    closeMenu()
    onRemove(i)
  }

  onMount(() => {
    const onDoc = (): void => closeMenu()
    document.addEventListener('click', onDoc)
    return () => document.removeEventListener('click', onDoc)
  })

  onDestroy(closeMenu)

  function cmp(a: string, b: string): number {
    return a.trim().localeCompare(b.trim(), undefined, { sensitivity: 'base', numeric: true })
  }

  function sortBy(field: 'name' | 'file'): void {
    onReorder([...rows].sort((a, b) => cmp(a[field], b[field])))
  }

  function clampDecimal(raw: string): string {
    let s = raw.replace(/,/g, '.').replace(/[^\d.]/g, '')
    const dot = s.indexOf('.')
    if (dot === -1) {
      return s
    }
    return s.slice(0, dot + 1) + s.slice(dot + 1).replace(/\./g, '').slice(0, 2)
  }

  function setDecimal(row: { volume: string; scale: string }, field: 'volume' | 'scale', el: HTMLInputElement, next: string): void {
    const v = clampDecimal(next)
    row[field] = v
    if (el.value !== v) {
      el.value = v
    }
  }

  function onDecimalKey(ev: KeyboardEvent): void {
    if (ev.ctrlKey || ev.metaKey || ev.altKey) {
      return
    }
    if (ev.key.length !== 1) {
      return
    }
    const el = ev.currentTarget as HTMLInputElement
    const start = el.selectionStart ?? el.value.length
    const end = el.selectionEnd ?? el.value.length
    const next = el.value.slice(0, start) + ev.key.replace(/,/g, '.') + el.value.slice(end)
    if (clampDecimal(next) !== next) {
      ev.preventDefault()
    }
  }

  function onDecimalBefore(row: { volume: string; scale: string }, field: 'volume' | 'scale', ev: InputEvent): void {
    const el = ev.currentTarget as HTMLInputElement
    if (ev.inputType.startsWith('delete') || ev.inputType.startsWith('history')) {
      return
    }
    const data = ev.data
    if (data == null) {
      return
    }
    const start = el.selectionStart ?? el.value.length
    const end = el.selectionEnd ?? el.value.length
    const insert = data.replace(/,/g, '.')
    const next = el.value.slice(0, start) + insert + el.value.slice(end)
    const clamped = clampDecimal(next)
    if (clamped !== next) {
      ev.preventDefault()
      if (ev.inputType === 'insertFromPaste' || ev.inputType === 'insertFromDrop' || insert.length > 1) {
        setDecimal(row, field, el, next)
      }
    }
  }

  function onDecimalInput(row: { volume: string; scale: string }, field: 'volume' | 'scale', ev: Event): void {
    const el = ev.currentTarget as HTMLInputElement
    setDecimal(row, field, el, el.value)
  }
</script>

<p class="lead">Edits the YAML file used for overlay alerts. Leave volume and scale blank to omit them (playback uses 1). When set, they must be non-negative numbers with at most two digits after the decimal point. The folder icon on File picks a file inside the Config media folder (including subfolders) and stores the path without that folder prefix.</p>

{#if !path}
  <section class="card">
    <header class="card-head">
      <h2>commands.yaml</h2>
      <span class="badge badge-warn">Not set</span>
    </header>
    <p class="hint">Set a media folder in Configuration, or turn on a custom commands.yaml path, then return here.</p>
  </section>
{:else}
  <section class="card">
    <header class="card-head">
      <h2>Commands</h2>
      <div class="status-cluster">
        <span class="badge">{rows.length}</span>
      </div>
    </header>
    <p class="path">File <code>{path}</code></p>
    {#if loading}
      <p class="hint">Loading…</p>
    {:else if rows.length === 0}
      <p class="hint">No commands yet. Add one, then save.</p>
    {:else}
      <div class="sort-row">
        <span>Sort by</span>
        <button class="btn btn-small" disabled={loading || saving} type="button" on:click={() => sortBy('name')}>Command</button>
        <button class="btn btn-small" disabled={loading || saving} type="button" on:click={() => sortBy('file')}>Filename</button>
      </div>
      <div class="command-list">
        {#each rows as row, i}
          <div class="command-row">
            <div class="command-main">
              <label class="field">
                Name
                <input autocomplete="off" bind:value={row.name} placeholder="jump" spellcheck="false" type="text" />
              </label>
              <label class="field">
                File
                <span class="file-in">
                  <input autocomplete="off" bind:value={row.file} placeholder="jump.mp3" spellcheck="false" type="text" />
                  <button
                    class="file-in-btn"
                    disabled={!hasMedia}
                    title={hasMedia ? 'Choose a file inside the media folder' : 'Set a media folder in Configuration'}
                    type="button"
                    aria-label="Choose file"
                    on:click={() => onPickFile(i)}
                  >
                    <svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
                      <path fill="currentColor" d="M1.5 3.5A1.5 1.5 0 0 1 3 2h3.2c.3 0 .6.1.8.4L8 3.5h5A1.5 1.5 0 0 1 14.5 5v7A1.5 1.5 0 0 1 13 13.5H3A1.5 1.5 0 0 1 1.5 12Z" />
                    </svg>
                  </button>
                </span>
              </label>
            </div>
            <div class="command-meta">
              <label class="field">
                Volume
                <input
                  autocomplete="off"
                  inputmode="decimal"
                  value={row.volume}
                  placeholder="omit"
                  spellcheck="false"
                  type="text"
                  on:keydown={onDecimalKey}
                  on:beforeinput={(e) => onDecimalBefore(row, 'volume', e)}
                  on:input={(e) => onDecimalInput(row, 'volume', e)}
                />
              </label>
              <label class="field">
                Scale
                <input
                  autocomplete="off"
                  inputmode="decimal"
                  value={row.scale}
                  placeholder="omit"
                  spellcheck="false"
                  type="text"
                  on:keydown={onDecimalKey}
                  on:beforeinput={(e) => onDecimalBefore(row, 'scale', e)}
                  on:input={(e) => onDecimalInput(row, 'scale', e)}
                />
              </label>
              <div class="command-more" on:click|stopPropagation>
                <button
                  class="file-in-btn command-more-btn"
                  type="button"
                  aria-label="More actions"
                  title="More"
                  on:click|stopPropagation={() => toggleMenu(i)}
                >
                  <svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
                    <circle cx="3" cy="8" r="1.5" fill="currentColor" />
                    <circle cx="8" cy="8" r="1.5" fill="currentColor" />
                    <circle cx="13" cy="8" r="1.5" fill="currentColor" />
                  </svg>
                </button>
                {#if menu === i}
                  <div class="command-menu">
                    {#if confirmDelete === i}
                      <p class="menu-note">Delete {row.name.trim() ? `“${row.name.trim()}”` : 'this command'}?</p>
                      <button type="button" on:click={() => { confirmDelete = -1 }}>Cancel</button>
                      <button class="menu-danger" type="button" on:click={() => confirmRemove(i)}>Delete</button>
                    {:else}
                      <button
                        disabled={!canTest || !row.file.trim()}
                        type="button"
                        title={canTest ? 'Play this command on the OBS overlay' : 'OBS overlay is offline'}
                        on:click={() => { closeMenu(); onTest(i) }}
                      >Test on overlay</button>
                      <button class="menu-danger" type="button" on:click={() => requestDelete(i)}>Delete</button>
                    {/if}
                  </div>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
    <div class="actions">
      <button class="btn" disabled={loading || !path} type="button" on:click={onAdd}>Add command</button>
      <button class="btn" disabled={loading || saving} type="button" on:click={onReload}>Reload</button>
      <button class="btn btn-primary" disabled={loading || saving} type="button" on:click={onSave}>
        {saving ? 'Saving…' : 'Save YAML'}
      </button>
    </div>
  </section>
{/if}
