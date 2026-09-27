<script lang="ts">
  import { clampDecimal } from './decimal'

  export let label: string
  export let value = ''
  export let placeholder = 'omit'

  function setDecimal(el: HTMLInputElement, next: string): void {
    const v = clampDecimal(next)
    value = v
    if (el.value !== v) {
      el.value = v
    }
  }

  function onKey(ev: KeyboardEvent): void {
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

  function onBefore(ev: InputEvent): void {
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
        setDecimal(el, next)
      }
    }
  }

  function onInput(ev: Event): void {
    const el = ev.currentTarget as HTMLInputElement
    setDecimal(el, el.value)
  }
</script>

<label class="field">
  {label}
  <input
    autocomplete="off"
    inputmode="decimal"
    {placeholder}
    spellcheck="false"
    type="text"
    {value}
    on:keydown={onKey}
    on:beforeinput={onBefore}
    on:input={onInput}
  />
</label>
