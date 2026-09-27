<script lang="ts">
  import Icon from './Icon.svelte'

  export let value = ''
  export let onChange: (value: string) => void = () => {}
  export let onClear: () => void = () => {}

  let input: HTMLInputElement

  function onInput(ev: Event): void {
    const next = (ev.currentTarget as HTMLInputElement).value
    value = next
    onChange(next)
  }

  function clear(): void {
    onClear()
    input?.focus()
  }
</script>

<span class="command-search">
  <input
    bind:this={input}
    autocomplete="off"
    aria-label="Search commands"
    placeholder="Search"
    spellcheck="false"
    type="text"
    {value}
    on:input={onInput}
  />
  {#if value}
    <button class="command-search-clear" type="button" aria-label="Clear search" title="Clear" on:click={clear}>
      <Icon name="close" size={12} />
    </button>
  {/if}
</span>

<style>
  .command-search {
    position: relative;
    flex: 1 1 11rem;
    min-width: 9rem;
    max-width: 16rem;
  }

  .command-search input {
    height: 28px;
    padding-right: 26px;
    font-size: 12px;
    font-weight: 500;
  }

  .command-search-clear {
    position: absolute;
    top: 50%;
    right: 4px;
    width: 18px;
    height: 18px;
    margin: 0;
    padding: 0;
    border: none;
    border-radius: 4px;
    background: transparent;
    color: var(--muted);
    display: flex;
    align-items: center;
    justify-content: center;
    transform: translateY(-50%);
    cursor: pointer;
  }

  .command-search-clear:hover {
    color: var(--text);
    background: rgba(255, 255, 255, 0.08);
  }
</style>
