#!/usr/bin/env node
// Captures each window pane and writes docs/screenshots/window.gif for the README.
// Uses ?screenshot=1 so Wails bindings are stubbed (no Go/Wails window required).

import { spawn } from 'node:child_process'
import { readFileSync, writeFileSync } from 'node:fs'
import { createServer } from 'node:net'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import pkg from 'gifenc'
import { PNG } from 'pngjs'
import { chromium } from 'playwright'

const { GIFEncoder, applyPalette, quantize } = pkg

const here = dirname(fileURLToPath(import.meta.url))
const repo = join(here, '..', '..')
const frontend = join(repo, 'frontend')
const outDir = join(repo, 'docs', 'screenshots')
const gifPath = join(outDir, 'window.gif')
const width = 760
const height = 680
const frameDelayMs = 2500

const panes = [
  { file: 'home.png', heading: 'Home', click: null },
  { file: 'config.png', heading: 'Configuration', click: 'Configuration' },
  { file: 'commands.png', heading: 'Commands', click: 'Commands' },
  { file: 'test.png', heading: 'Test message', click: 'Test' },
]

function unusedPort() {
  return new Promise((resolve, reject) => {
    const s = createServer()
    s.listen(0, '127.0.0.1', () => {
      const { port } = s.address()
      s.close((err) => (err ? reject(err) : resolve(port)))
    })
    s.on('error', reject)
  })
}

function startVite(port) {
  const child = spawn('npx', ['vite', '--host', '127.0.0.1', '--port', String(port), '--strictPort'], {
    cwd: frontend,
    stdio: ['ignore', 'pipe', 'pipe'],
    env: { ...process.env, BROWSER: 'none' },
  })
  let output = ''
  child.stdout.on('data', (b) => {
    output += b.toString()
  })
  child.stderr.on('data', (b) => {
    output += b.toString()
  })
  const ready = new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error(`vite did not start:\n${output}`))
    }, 30000)
    const onChunk = () => {
      if (output.includes('Local:')) {
        clearTimeout(timer)
        resolve()
      }
    }
    child.stdout.on('data', onChunk)
    child.stderr.on('data', onChunk)
    child.on('error', reject)
    child.on('exit', (code) => {
      if (code) {
        clearTimeout(timer)
        reject(new Error(`vite exited ${code}:\n${output}`))
      }
    })
  })
  return { child, ready }
}

function writeGif(pngPaths, dest) {
  const frames = pngPaths.map((p) => PNG.sync.read(readFileSync(p)))
  const w = frames[0].width
  const h = frames[0].height
  for (const f of frames) {
    if (f.width !== w || f.height !== h) {
      throw new Error('screenshot frames must be the same size')
    }
  }
  const gif = GIFEncoder()
  for (const frame of frames) {
    const palette = quantize(frame.data, 256)
    const index = applyPalette(frame.data, palette)
    gif.writeFrame(index, w, h, { palette, delay: frameDelayMs, repeat: 0 })
  }
  gif.finish()
  writeFileSync(dest, Buffer.from(gif.bytes()))
}

async function main() {
  const port = await unusedPort()
  const vite = startVite(port)
  const browser = await chromium.launch()
  const pngPaths = []
  try {
    await vite.ready
    const page = await browser.newPage({
      viewport: { width, height },
      deviceScaleFactor: 1,
    })
    await page.goto(`http://127.0.0.1:${port}/?screenshot=1`, { waitUntil: 'networkidle' })
    await page.locator('.shell').waitFor()

    const nav = page.getByRole('navigation', { name: 'Window panes' })
    for (const pane of panes) {
      if (pane.click) {
        await nav.getByRole('button', { name: pane.click, exact: true }).click()
      }
      await page.getByRole('heading', { level: 1, name: pane.heading }).waitFor()
      await page.waitForTimeout(150)
      const dest = join(outDir, pane.file)
      await page.locator('.shell').screenshot({ path: dest, type: 'png' })
      pngPaths.push(dest)
      console.log('wrote', dest)
    }
  } finally {
    await browser.close()
    vite.child.kill('SIGTERM')
  }
  writeGif(pngPaths, gifPath)
  console.log('wrote', gifPath)
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
