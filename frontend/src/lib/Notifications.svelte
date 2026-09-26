<script context="module" lang="ts">
  export type NoticeKind = 'ok' | 'err'

  export type Notice = {
    id: number
    text: string
    kind: NoticeKind
  }
</script>

<script lang="ts">
  import { onDestroy } from 'svelte'
  import { fly } from 'svelte/transition'
  import { flip } from 'svelte/animate'

  /** How long a toast stays visible. Change this to retune the timeout. */
  const NOTIFICATION_MS = 4000

  export let notices: Notice[] = []

  const timers = new Map<number, number>()
  const paused = new Set<number>()

  function arm(id: number): void {
    const timer = window.setTimeout(() => dismiss(id), NOTIFICATION_MS)
    timers.set(id, timer)
  }

  function dismiss(id: number): void {
    const timer = timers.get(id)
    if (timer != null) {
      window.clearTimeout(timer)
      timers.delete(id)
    }
    paused.delete(id)
    notices = notices.filter((n) => n.id !== id)
  }

  function pause(id: number): void {
    paused.add(id)
    const timer = timers.get(id)
    if (timer == null) {
      return
    }
    window.clearTimeout(timer)
    timers.delete(id)
  }

  function resume(id: number): void {
    paused.delete(id)
    if (!notices.some((n) => n.id === id) || timers.has(id)) {
      return
    }
    arm(id)
  }

  function watch(items: Notice[]): void {
    const live = new Set(items.map((n) => n.id))
    for (const id of timers.keys()) {
      if (live.has(id)) {
        continue
      }
      const timer = timers.get(id)
      if (timer != null) {
        window.clearTimeout(timer)
      }
      timers.delete(id)
    }
    for (const id of paused) {
      if (!live.has(id)) {
        paused.delete(id)
      }
    }
    for (const item of items) {
      if (!timers.has(item.id) && !paused.has(item.id)) {
        arm(item.id)
      }
    }
  }

  $: watch(notices)

  onDestroy(() => {
    for (const timer of timers.values()) {
      window.clearTimeout(timer)
    }
    timers.clear()
    paused.clear()
  })
</script>

<div class="notices" aria-live="polite">
  {#each notices as notice (notice.id)}
    <div
      class="notice"
      class:notice-err={notice.kind === 'err'}
      role={notice.kind === 'err' ? 'alert' : 'status'}
      in:fly={{ x: 28, duration: 280 }}
      out:fly={{ x: 24, duration: 180 }}
      animate:flip={{ duration: 180 }}
      on:mouseenter={() => pause(notice.id)}
      on:mouseleave={() => resume(notice.id)}
    >
      <span class="notice-mark" aria-hidden="true"></span>
      <p>{notice.text}</p>
    </div>
  {/each}
</div>

<style>
  .notices {
    position: fixed;
    top: 16px;
    right: 22px;
    z-index: 40;
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: min(320px, calc(100% - 44px));
    pointer-events: none;
  }

  .notice {
    pointer-events: auto;
    display: flex;
    align-items: flex-start;
    gap: 8px;
    width: 100%;
    padding: 10px 12px;
    border-radius: var(--radius);
    background:
      linear-gradient(0deg, rgba(52, 211, 153, 0.12), rgba(52, 211, 153, 0.12)),
      var(--bg-elevated);
    border: 1px solid var(--border-strong);
    box-shadow: var(--shadow);
    color: var(--text);
    font-size: 13px;
    line-height: 1.45;
  }

  .notice p {
    margin: 0;
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .notice-mark {
    flex: 0 0 auto;
    width: 6px;
    height: 6px;
    margin-top: 6px;
    border-radius: 50%;
    background: var(--ok);
  }

  .notice.notice-err {
    background:
      linear-gradient(0deg, rgba(248, 113, 113, 0.12), rgba(248, 113, 113, 0.12)),
      var(--bg-elevated);
  }

  .notice.notice-err .notice-mark {
    background: var(--danger);
  }
</style>
