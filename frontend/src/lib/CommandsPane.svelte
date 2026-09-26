<script lang="ts">
  import { onDestroy, onMount, tick } from 'svelte'
  import {
    GetAlertCommands,
    PickAlertMediaFile,
    PreviewAlert,
    SaveAlertCommands,
  } from '../../wailsjs/go/commands/Commands.js'
  import { commands as cmdModels } from '../../wailsjs/go/models'
  import type { NoticeKind } from './Notifications.svelte'

  type CommandRow = { name: string; file: string; volume: string; scale: string }

  export let mediaPath: string
  export let commandsFilePath: string
  export let canTest: boolean
  export let alertsEnabled: boolean
  export let notify: (text: string, kind?: NoticeKind) => void

  let path = ''
  let rows: CommandRow[] = []
  let saving = false
  let loading = false
  let alive = true

  $: hasMedia = mediaPath.trim() !== ''

  let menu = -1
  let confirmDelete = -1
  let namesTick = 0
  let paneScroll: HTMLElement

  function nameKey(name: string): string {
    return name.trim().toLowerCase()
  }

  function onNameInput(row: { name: string }, ev: Event): void {
    row.name = (ev.currentTarget as HTMLInputElement).value
    namesTick += 1
  }

  function collectDuplicateNames(list: CommandRow[], tick: number): Set<string> {
    void tick
    const counts = new Map<string, number>()
    for (const row of list) {
      const key = nameKey(row.name)
      if (key === '') {
        continue
      }
      counts.set(key, (counts.get(key) ?? 0) + 1)
    }
    const dup = new Set<string>()
    for (const [key, n] of counts) {
      if (n > 1) {
        dup.add(key)
      }
    }
    return dup
  }

  $: duplicateNames = collectDuplicateNames(rows, namesTick)
  $: hasDuplicateNames = duplicateNames.size > 0

  async function save(): Promise<void> {
    if (hasDuplicateNames) {
      return
    }
    saving = true
    try {
      const commands = rows.map((row, i) => {
        const name = row.name.trim()
        const file = row.file.trim()
        if (!name || !file) {
          throw new Error(`Command ${i + 1} needs a name and file`)
        }
        const item: { name: string; file: string; volume?: number; scale?: number } = { name, file }
        const volume = optionalFloat(row.volume, `Command ${name} volume`)
        const scale = optionalFloat(row.scale, `Command ${name} scale`)
        if (volume !== undefined) {
          item.volume = volume
        }
        if (scale !== undefined) {
          item.scale = scale
        }
        return item
      })
      await SaveAlertCommands(cmdModels.AlertCommandsFile.createFrom({ path, commands }))
      await load()
      notify('Commands saved')
    } catch (e) {
      notify(String(e), 'err')
    } finally {
      if (alive) {
        saving = false
      }
    }
  }

  async function addAndFocus(): Promise<void> {
    rows = [...rows, emptyRow()]
    await tick()
    const row = paneScroll?.querySelector<HTMLElement>('.command-row:last-child')
    const name = row?.querySelector<HTMLInputElement>('.command-main input')
    row?.scrollIntoView({ block: 'end' })
    name?.focus({ preventScroll: true })
  }

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
    rows = rows.filter((_, index) => index !== i)
  }

  onMount(() => {
    const onDoc = (): void => closeMenu()
    document.addEventListener('click', onDoc)
    void load()
    return () => document.removeEventListener('click', onDoc)
  })

  onDestroy(() => {
    alive = false
    closeMenu()
  })

  function cmp(a: string, b: string): number {
    return a.trim().localeCompare(b.trim(), undefined, { sensitivity: 'base', numeric: true })
  }

  function sortBy(field: 'name' | 'file'): void {
    rows = [...rows].sort((a, b) => cmp(a[field], b[field]))
  }

  function playableIndexes(list: CommandRow[]): number[] {
    const out: number[] = []
    list.forEach((row, i) => {
      if (row.file.trim()) {
        out.push(i)
      }
    })
    return out
  }

  $: playable = playableIndexes(rows)
  $: canPreview = canTest && alertsEnabled
  $: canPlayRandom = canPreview && playable.length > 0
  $: previewTitle = !alertsEnabled
    ? 'Alerts are off in Configuration'
    : !canTest
      ? 'OBS overlay is offline'
      : ''

  function playRandom(): void {
    if (!canPlayRandom) {
      return
    }
    const index = playable[Math.floor(Math.random() * playable.length)]
    void testCommand(index)
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

  function optionalFloat(raw: string, label: string): number | undefined {
    const t = raw.trim().replace(/,/g, '.').replace(/\.$/, '')
    if (!t) {
      return undefined
    }
    if (!/^(?:\d+(?:\.\d{1,2})?|\.\d{1,2})$/.test(t)) {
      throw new Error(`${label} must be a number with at most two digits after the decimal point`)
    }
    const n = Number(t)
    if (!Number.isFinite(n)) {
      throw new Error(`${label} must be a number`)
    }
    return n
  }

  function formatDecimal(n: number): string {
    return String(Number(n.toFixed(2)))
  }

  function emptyRow(): CommandRow {
    return { name: '', file: '', volume: '', scale: '' }
  }

  async function load(): Promise<void> {
    if (!commandsFilePath.trim()) {
      path = ''
      rows = []
      return
    }
    loading = true
    try {
      const got = await GetAlertCommands()
      if (!alive) {
        return
      }
      path = got.path ?? ''
      rows = (got.commands ?? []).map((c) => ({
        name: c.name ?? '',
        file: c.file ?? '',
        volume: c.volume == null ? '' : formatDecimal(c.volume),
        scale: c.scale == null ? '' : formatDecimal(c.scale),
      }))
    } catch (e) {
      if (!alive) {
        return
      }
      path = ''
      rows = []
      notify(String(e), 'err')
    } finally {
      if (alive) {
        loading = false
      }
    }
  }

  async function testCommand(index: number): Promise<void> {
    const row = rows[index]
    if (!row || !row.file.trim()) {
      notify('Command needs a file', 'err')
      return
    }
    try {
      const volume = row.volume.trim() ? Number(row.volume) : 1
      const scale = row.scale.trim() ? Number(row.scale) : 1
      await PreviewAlert(row.file.trim(), volume, scale)
      const label = row.name.trim()
      notify(label ? `Alert sent to overlay: ${label}` : 'Alert sent to overlay')
    } catch (e) {
      notify(String(e), 'err')
    }
  }

  async function pickFile(index: number): Promise<void> {
    try {
      const p = await PickAlertMediaFile(mediaPath)
      if (!p || !alive) {
        return
      }
      rows = rows.map((row, i) => (i === index ? { ...row, file: p } : row))
    } catch (e) {
      notify(String(e), 'err')
    }
  }
</script>

<div class="pane-dock">
<div class="pane-scroll" bind:this={paneScroll}>
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
                <input
                  autocomplete="off"
                  class:name-dup={duplicateNames.has(nameKey(row.name))}
                  aria-invalid={duplicateNames.has(nameKey(row.name))}
                  value={row.name}
                  placeholder="jump"
                  spellcheck="false"
                  type="text"
                  on:input={(e) => onNameInput(row, e)}
                />
                {#if duplicateNames.has(nameKey(row.name))}
                  <span class="name-error">This command already exists</span>
                {/if}
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
                    on:click={() => pickFile(i)}
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
                        disabled={!canPreview || !row.file.trim()}
                        type="button"
                        title={previewTitle || 'Play this command on the OBS overlay'}
                        on:click={() => { closeMenu(); void testCommand(i) }}
                      >Play</button>
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
  </section>
{/if}
</div>
{#if path}
  <div class="pane-footer">
    <div class="actions">
      <button class="btn" disabled={loading || !path} type="button" on:click={addAndFocus}>Add command</button>
      <button
        class="btn"
        disabled={loading || !canPlayRandom}
        title={previewTitle || (playable.length ? 'Play a random command on the OBS overlay' : 'Add a command with a file')}
        type="button"
        on:click={playRandom}
      >Play random</button>
      <button class="btn" disabled={loading || saving} type="button" on:click={() => void load()}>Reload</button>
      <button
        class="btn btn-primary"
        disabled={loading || saving || hasDuplicateNames}
        title={hasDuplicateNames ? 'Fix duplicate command names before saving' : undefined}
        type="button"
        on:click={save}
      >
        {saving ? 'Saving…' : 'Save YAML'}
      </button>
    </div>
  </div>
{/if}
</div>
