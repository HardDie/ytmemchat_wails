<script lang="ts">
  import type { IconName } from './types'

  export let name: IconName
  export let size = 16
</script>

<svg viewBox="0 0 16 16" width={size} height={size} aria-hidden="true">
  {#if name === 'plus'}
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" d="M8 3.25v9.5M3.25 8h9.5" />
  {:else if name === 'reload'}
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M3.1 6.15A4.6 4.6 0 0 1 12.4 5.2" />
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M12.9 9.85A4.6 4.6 0 0 1 3.6 10.8" />
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M11.15 2.7v2.85h2.7" />
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M4.85 13.3v-2.85H2.15" />
  {:else if name === 'save'}
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" d="M3 2.75h7.15L13.25 5.85V13.25H3V2.75z" />
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" d="M5.25 2.75V6h4.6" />
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" d="M5.25 13.25v-3.2h5.5v3.2" />
  {:else if name === 'export'}
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M3 6.4V12.75h10V6.4" />
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M8 8.6V2.6" />
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M5.4 4.9 8 2.6l2.6 2.3" />
  {:else if name === 'import'}
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M3 6.4V12.75h10V6.4" />
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M8 2.6v6" />
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M5.4 6.3 8 8.6l2.6-2.3" />
  {:else if name === 'dice'}
    <rect x="2.75" y="2.75" width="10.5" height="10.5" rx="2" fill="none" stroke="currentColor" stroke-width="1.5" />
    <circle cx="5.75" cy="5.75" r="0.9" fill="currentColor" />
    <circle cx="8" cy="8" r="0.9" fill="currentColor" />
    <circle cx="10.25" cy="10.25" r="0.9" fill="currentColor" />
  {:else if name === 'copy'}
    <path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" d="M5.6 4.15h-1.1A1.25 1.25 0 0 0 3.25 5.4v7.1c0 .69.56 1.25 1.25 1.25h7c.69 0 1.25-.56 1.25-1.25V5.4c0-.69-.56-1.25-1.25-1.25H10.4" />
    <rect x="5.85" y="2.2" width="4.3" height="2.7" rx="0.55" fill="none" stroke="currentColor" stroke-width="1.5" />
  {:else if name === 'close'}
    <path fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" d="M4.5 4.5l7 7M11.5 4.5l-7 7" />
  {:else if name === 'folder'}
    <path fill="currentColor" d="M1.5 3.5A1.5 1.5 0 0 1 3 2h3.2c.3 0 .6.1.8.4L8 3.5h5A1.5 1.5 0 0 1 14.5 5v7A1.5 1.5 0 0 1 13 13.5H3A1.5 1.5 0 0 1 1.5 12Z" />
  {:else if name === 'more'}
    <circle cx="3" cy="8" r="1.5" fill="currentColor" />
    <circle cx="8" cy="8" r="1.5" fill="currentColor" />
    <circle cx="13" cy="8" r="1.5" fill="currentColor" />
  {/if}
</svg>
