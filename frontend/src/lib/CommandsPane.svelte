<script lang="ts">
  import { onDestroy, onMount, tick } from 'svelte'
  import {
    ExportAlertCommands,
    GetAlertCommands,
    ImportAlertCommands,
    PickAlertMediaFile,
    PreviewAlert,
    SaveAlertCommands,
  } from '../../wailsjs/go/commands/Commands.js'
  import { commands as cmdModels } from '../../wailsjs/go/models'
  import { ClipboardSetText } from '../../wailsjs/runtime/runtime'
  import type { NoticeKind } from './Notifications.svelte'
  import CommandRow from './commands/CommandRow.svelte'
  import CommandSearch from './commands/CommandSearch.svelte'
  import IconButton from './commands/IconButton.svelte'
  import { formatDecimal, optionalFloat } from './commands/decimal'
  import type { CommandFields } from './commands/types'

  export let mediaPath: string
  export let commandsFilePath: string
  export let token: string
  export let canTest: boolean
  export let alertsEnabled: boolean
  export let notify: (text: string, kind?: NoticeKind) => void
  export let onImported: () => Promise<void> = async () => {}

  let path = ''
  let rows: CommandFields[] = []
  let saving = false
  let exporting = false
  let importing = false
  let loading = false
  let alive = true

  $: hasMedia = mediaPath.trim() !== ''

  let menu = -1
  let confirmDelete = -1
  let namesTick = 0
  let paneScroll: HTMLElement
  let query = ''

  function nameKey(name: string): string {
    return name.trim().toLowerCase()
  }

  function collectDuplicateNames(list: CommandFields[], tick: number): Set<string> {
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
    query = ''
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

  function rowMatches(row: CommandFields, raw: string): boolean {
    const q = raw.trim().toLowerCase()
    if (q === '') {
      return true
    }
    return row.name.toLowerCase().includes(q) || row.file.toLowerCase().includes(q)
  }

  function onSearch(next: string): void {
    if (menu >= 0 && rows[menu] && !rowMatches(rows[menu], next)) {
      closeMenu()
    }
  }

  function clearSearch(): void {
    query = ''
    closeMenu()
  }

  $: visibleRows = rows
    .map((row, i) => ({ row, i }))
    .filter(({ row }) => rowMatches(row, query))

  function playableIndexes(list: CommandFields[]): number[] {
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

  function prefixedLines(list: CommandFields[], prefix: string, tick: number): string[] {
    void tick
    const p = prefix.trim() || '@'
    const lines: string[] = []
    for (const row of list) {
      const name = row.name.trim()
      if (name === '') {
        continue
      }
      lines.push(p + name)
    }
    return lines
  }

  $: copyLines = prefixedLines(rows, token, namesTick)

  async function exportCommands(): Promise<void> {
    exporting = true
    try {
      const saved = await ExportAlertCommands()
      if (!alive || !saved) {
        return
      }
      notify('Commands exported')
    } catch (e) {
      notify(String(e), 'err')
    } finally {
      if (alive) {
        exporting = false
      }
    }
  }

  async function importCommands(): Promise<void> {
    importing = true
    try {
      const dest = await ImportAlertCommands()
      if (!alive || !dest) {
        return
      }
      await onImported()
      await tick()
      await load()
      notify('Commands imported')
    } catch (e) {
      notify(String(e), 'err')
    } finally {
      if (alive) {
        importing = false
      }
    }
  }

  async function copyCommands(): Promise<void> {
    if (copyLines.length === 0) {
      return
    }
    try {
      await ClipboardSetText(copyLines.join('\n'))
      notify('Commands copied')
    } catch (e) {
      notify(String(e), 'err')
    }
  }

  function playRandom(): void {
    if (!canPlayRandom) {
      return
    }
    const index = playable[Math.floor(Math.random() * playable.length)]
    void testCommand(index)
  }

  function emptyRow(): CommandFields {
    return { name: '', file: '', volume: '', scale: '' }
  }

  function nameFromMediaPath(file: string): string {
    const base = (file.split('/').pop() ?? '').trim()
    const dot = base.lastIndexOf('.')
    if (dot <= 0) {
      return base
    }
    return base.slice(0, dot).trim()
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
      rows = rows.map((row, i) => {
        if (i !== index) {
          return row
        }
        const next = { ...row, file: p }
        if (row.name.trim() === '') {
          next.name = nameFromMediaPath(p)
        }
        return next
      })
      namesTick += 1
    } catch (e) {
      notify(String(e), 'err')
    }
  }

  $: busy = loading || saving || exporting || importing
  $: saveTitle = hasDuplicateNames
    ? 'Fix duplicate command names before saving'
    : saving
      ? 'Saving…'
      : 'Save YAML'
</script>

<div class="pane-dock">
  <div class="pane-scroll" bind:this={paneScroll}>
    <p class="lead">Each command plays a clip on the overlay. Volume and scale can stay blank.</p>

    {#if !path}
      <section class="card">
        <header class="card-head">
          <h2>commands.yaml</h2>
          <span class="badge badge-warn">Not set</span>
        </header>
        <p class="hint">Set a media folder in Configuration, or turn on a custom commands.yaml path, then return here.</p>
        <p class="hint">Import unpacks a shared zip into a folder you choose.</p>
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
            <CommandSearch bind:value={query} onChange={onSearch} onClear={clearSearch} />
          </div>
          {#if visibleRows.length === 0}
            <p class="hint">No commands match.</p>
          {:else}
            <div class="command-list">
              {#each visibleRows as { row, i } (i)}
                <CommandRow
                  {row}
                  duplicate={duplicateNames.has(nameKey(row.name))}
                  {hasMedia}
                  menuOpen={menu === i}
                  confirmDelete={confirmDelete === i}
                  {canPreview}
                  {previewTitle}
                  onName={() => namesTick += 1}
                  onPick={() => pickFile(i)}
                  onToggleMenu={() => toggleMenu(i)}
                  onPlay={() => { closeMenu(); void testCommand(i) }}
                  onRequestDelete={() => requestDelete(i)}
                  onCancelDelete={() => { confirmDelete = -1 }}
                  onConfirmDelete={() => confirmRemove(i)}
                />
              {/each}
            </div>
          {/if}
        {/if}
      </section>
    {/if}
  </div>
  <div class="pane-footer">
    <div class="actions command-footer">
      <IconButton name="plus" label="Add command" disabled={loading || !path} on:click={addAndFocus} />
      <IconButton name="reload" label="Reload" disabled={busy || !path} on:click={() => void load()} />
      <IconButton
        name="save"
        primary
        label={saving ? 'Saving…' : 'Save YAML'}
        title={saveTitle}
        disabled={busy || hasDuplicateNames || !path}
        on:click={save}
      />
      <IconButton
        name="export"
        label={exporting ? 'Exporting…' : 'Export'}
        title={exporting ? 'Exporting…' : 'Export the saved commands and the media files they use'}
        disabled={busy || !path}
        on:click={() => void exportCommands()}
      />
      <IconButton
        name="import"
        label={importing ? 'Importing…' : 'Import'}
        title={importing ? 'Importing…' : 'Import a zip of commands and the media files they use'}
        disabled={busy}
        on:click={() => void importCommands()}
      />
      <div class="command-footer-end">
        <IconButton
          name="dice"
          label="Play random"
          title={previewTitle || (playable.length ? 'Play a random command on the OBS overlay' : 'Add a command with a file')}
          disabled={loading || !canPlayRandom}
          on:click={playRandom}
        />
        <IconButton
          name="copy"
          label="Copy commands"
          title={copyLines.length ? 'Copy each command with the token, one per line' : 'Add a command name first'}
          disabled={loading || copyLines.length === 0}
          on:click={() => void copyCommands()}
        />
      </div>
    </div>
  </div>
</div>

<style>
  .sort-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
    margin: 16px 0 14px;
    font-size: 12px;
    font-weight: 600;
    color: var(--muted);
  }

  .command-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin: 0 0 16px;
    min-width: 0;
    width: 100%;
    container-type: inline-size;
    container-name: commands;
  }

  .command-footer {
    flex-wrap: nowrap;
    align-items: center;
  }

  .command-footer-end {
    margin-left: auto;
    display: flex;
    gap: 8px;
  }
</style>
