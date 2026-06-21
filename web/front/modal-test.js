// Verify WeatherModal fetches weather data correctly
import { build } from 'esbuild'
import fs from 'node:fs'
import path from 'node:path'
import { JSDOM } from 'jsdom'

const result = await build({
  entryPoints: ['./src/components/WeatherModal.jsx'],
  bundle: true,
  write: false,
  platform: 'node',
  format: 'esm',
  jsx: 'automatic',
  loader: { '.jsx': 'jsx', '.js': 'jsx' },
  external: ['react', 'react-dom', 'react-dom/client'],
  resolveExtensions: ['.jsx', '.js'],
})

const code = result.outputFiles[0].text
const tmp = path.resolve('.modal-bundle.mjs')
fs.writeFileSync(tmp, code)

// Set up jsdom
const dom = new JSDOM('<!DOCTYPE html><html><body><div id="root"></div></body></html>', {
  url: 'http://localhost/',
  pretendToBeVisual: true,
})

global.window = dom.window
global.document = dom.window.document
global.navigator = dom.window.navigator
global.HTMLElement = dom.window.HTMLElement
global.Element = dom.window.Element
global.Intl = dom.window.Intl || Intl
global.fetch = async (url) => {
  if (String(url).includes('/api/events')) {
    return { ok: true, json: async () => ([]) }
  }
  // Mock weather response
  return {
    ok: true,
    json: async () => ({
      hourly: {
        time: ['2026-06-20T08:00', '2026-06-20T09:00', '2026-06-20T10:00'],
        temperature_2m: [25.5, 26.0, 27.0],
        relative_humidity_2m: [60, 65, 62],
        cloud_cover: [40, 50, 30],
      },
    }),
  }
}

// Now dynamically import the bundled module
const mod = await import(tmp)
const WeatherModal = mod.default

const React = (await import('react')).default
const ReactDOM = (await import('react-dom/client')).default

const outage = {
  unique_hash: 'h1',
  city: 'آمل',
  address: 'تست',
  start: '2026-06-20T08:00:00+03:30',
  end: '2026-06-20T23:55:00+03:30',
}

const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(
  React.createElement(WeatherModal, {
    open: true,
    onClose: () => {},
    outage,
    providerId: 'open-meteo',
  })
)

await new Promise((r) => setTimeout(r, 2500))

const html = document.body.innerHTML
console.log('=== WeatherModal test ===')
console.log('Modal title (هواشناسی بازه قطعی):', html.includes('هواشناسی بازه قطعی'))
console.log('City (آمل):', html.includes('آمل'))
console.log('Temperature (25.5):', html.includes('25.5'))
console.log('Humidity (60):', html.includes('60'))
console.log('Cloud (40):', html.includes('40'))
console.log('Loading shimmer:', html.includes('shimmer'))
console.log('Sun gradient (sunGrad):', html.includes('sunGrad'))

fs.unlinkSync(tmp)
console.log('\n=== Done ===')
