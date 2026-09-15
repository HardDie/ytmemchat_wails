import './style.css'

async function boot(): Promise<void> {
  if (new URLSearchParams(location.search).has('screenshot')) {
    await import('./screenshotBridge')
  }
  const { default: App } = await import('./App.svelte')
  const target = document.getElementById('app')
  if (!target) {
    throw new Error('#app missing')
  }
  new App({ target })
}

void boot()
