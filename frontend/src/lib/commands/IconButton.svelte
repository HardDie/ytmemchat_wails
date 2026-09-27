<script lang="ts">
  import Icon from './Icon.svelte'
  import type { IconName } from './types'

  export let name: IconName
  export let label: string
  export let title = ''
  export let disabled = false
  export let primary = false
</script>

<button
  class="btn btn-icon"
  class:btn-primary={primary}
  {disabled}
  aria-label={label}
  title={title || label}
  type="button"
  on:click
>
  <Icon {name} />
</button>

<style>
  .btn-icon {
    min-width: 32px;
    width: 32px;
    padding: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
</style>
