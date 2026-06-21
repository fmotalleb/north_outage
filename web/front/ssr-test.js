// SSR test using esbuild to bundle JSX
import { build } from 'esbuild'
import path from 'node:path'
import fs from 'node:fs'

const result = await build({
  entryPoints: [path.resolve('ssr-entry.jsx')],
  bundle: true,
  write: false,
  platform: 'node',
  format: 'esm',
  external: ['react', 'react-dom', 'react-dom/server'],
  jsx: 'automatic',
  loader: { '.jsx': 'jsx', '.js': 'jsx' },
  resolveExtensions: ['.jsx', '.js'],
})

const code = result.outputFiles[0].text
const tmpFile = path.resolve('.ssr-entry.bundle.mjs')
fs.writeFileSync(tmpFile, code)

const mod = await import(tmpFile)
const App = mod.App
const Header = mod.Header
const FilterBar = mod.FilterBar
const OutageCard = mod.OutageCard
const WeatherProviderSelector = mod.WeatherProviderSelector

const React = (await import('react')).default
const { renderToString } = await import('react-dom/server')

const outages = JSON.parse(fs.readFileSync('sample-data.json', 'utf8'))

let pass = 0, fail = 0
function assert(label, cond) {
  if (cond) { console.log(`  ✓ ${label}`); pass++ }
  else { console.log(`  ✗ ${label}`); fail++ }
}

console.log('\n=== SSR component tests ===\n')

const headerHtml = renderToString(
  React.createElement(Header, {
    total: 7, active: 4, upcoming: 3, past: 0,
    lastUpdated: new Date('2026-06-20T08:00:00Z'),
  })
)
assert('Header renders h1', headerHtml.includes('قطعی‌های برنامه‌ریزی‌شده'))
assert('Header shows total count', headerHtml.includes('7'))
assert('Header shows active count', headerHtml.includes('4'))
assert('Header shows upcoming count', headerHtml.includes('3'))
assert('Header includes stat gradients', headerHtml.includes('bg-gradient-to-l'))

const filterHtml = renderToString(
  React.createElement(FilterBar, {
    cities: ['قایمشهر', 'آمل', 'بابل', 'سوادکوه'],
    filters: { city: 'all', status: 'all', q: '', date: 'all' },
    setFilters: () => {},
    sort: 'start_asc',
    setSort: () => {},
    resultCount: 7,
    onRefresh: () => {},
    loading: false,
  })
)
assert('FilterBar has search input', filterHtml.includes('placeholder='))
assert('FilterBar has city select with آمل', filterHtml.includes('آمل'))
assert('FilterBar has status select', filterHtml.includes('در جریان'))
assert('FilterBar has date select', filterHtml.includes('فردا'))
assert('FilterBar has sort dropdown', filterHtml.includes('نزدیک‌ترین'))

const cardHtml = renderToString(
  React.createElement(OutageCard, {
    outage: outages[0],
    weatherProviderId: 'open-meteo',
    expanded: false,
    onToggle: () => {},
  })
)
assert('Card shows city name', cardHtml.includes('قایمشهر'))
assert('Card shows address', cardHtml.includes('چمازکتی') || cardHtml.includes('جویبار'))
assert('Card shows status badge (upcoming)', cardHtml.includes('پیش‌رو'))
assert('Card shows start time block', cardHtml.includes('شروع قطعی'))
assert('Card shows end time block', cardHtml.includes('پایان قطعی'))
assert('Card has expand button', cardHtml.includes('جزئیات'))

const selHtml = renderToString(
  React.createElement(WeatherProviderSelector, {
    value: 'open-meteo',
    onChange: () => {},
  })
)
assert('Selector lists Open-Meteo', selHtml.includes('Open-Meteo'))
assert('Selector lists Met.no', selHtml.includes('Met.no'))
assert('Selector lists 7Timer!', selHtml.includes('7Timer!'))
assert('Selector mentions no API key', selHtml.includes('کلید'))

const appHtml = renderToString(React.createElement(App))
assert('App renders without errors', appHtml.length > 1000)

console.log(`\n=== ${pass} passed, ${fail} failed ===\n`)
fs.unlinkSync(tmpFile)
process.exit(fail === 0 ? 0 : 1)
