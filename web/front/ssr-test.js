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
const Dropdown = mod.Dropdown

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

// FilterBar with new Dropdown
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
assert('FilterBar uses Dropdown trigger buttons', filterHtml.includes('aria-haspopup="listbox"'))
assert('FilterBar has aria-expanded', filterHtml.includes('aria-expanded'))
assert('FilterBar dropdown trigger for city', filterHtml.includes('همه شهرها'))
assert('FilterBar dropdown trigger for status', filterHtml.includes('همه وضعیت‌ها'))
assert('FilterBar dropdown trigger for date', filterHtml.includes('همه تاریخ‌ها'))
assert('FilterBar dropdown trigger for sort', filterHtml.includes('نزدیک‌ترین'))
assert('FilterBar has refresh button', filterHtml.includes('بروزرسانی'))
assert('FilterBar has reset button', filterHtml.includes('پاک کردن فیلترها'))
assert('FilterBar shows result count', filterHtml.includes('7') && filterHtml.includes('نتیجه'))

// Dropdown component
const dropdownHtml = renderToString(
  React.createElement(Dropdown, {
    value: 'آمل',
    onChange: () => {},
    options: [
      { value: 'all', label: 'همه' },
      { value: 'آمل', label: 'آمل' },
      { value: 'بابل', label: 'بابل' },
    ],
    label: 'شهر',
  })
)
assert('Dropdown renders label', dropdownHtml.includes('شهر'))
assert('Dropdown shows selected value', dropdownHtml.includes('آمل'))
assert('Dropdown has aria-haspopup', dropdownHtml.includes('aria-haspopup'))
assert('Dropdown has aria-expanded', dropdownHtml.includes('aria-expanded'))
assert('Dropdown trigger has dark styling', dropdownHtml.includes('bg-white/5'))
assert('Dropdown has chevron icon', dropdownHtml.includes('points="9 18 15 12 9 6"'))
assert('Dropdown menu is closed by default (no role=listbox)', !dropdownHtml.includes('role="listbox"'))
assert('Dropdown placeholder styling for empty value', renderToString(
  React.createElement(Dropdown, {
    value: '',
    onChange: () => {},
    options: [{ value: 'a', label: 'A' }],
  })
).includes('text-slate-500'))

// OutageCard
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
assert('Card shows status badge', cardHtml.includes('پیش‌رو') || cardHtml.includes('در جریان') || cardHtml.includes('پایان‌یافته'))
assert('Card shows start time block', cardHtml.includes('شروع قطعی'))
assert('Card shows end time block', cardHtml.includes('پایان قطعی'))
assert('Card has expand button', cardHtml.includes('جزئیات'))

// WeatherProviderSelector
const selHtml = renderToString(
  React.createElement(WeatherProviderSelector, {
    value: 'open-meteo',
    onChange: () => {},
  })
)
assert('Selector lists Open-Meteo', selHtml.includes('Open-Meteo'))
assert('Selector lists Met.no', selHtml.includes('Met.no'))
assert('Selector does not list 7Timer!', !selHtml.includes('7Timer!'))
assert('Selector mentions no API key', selHtml.includes('کلید'))

// App
const appHtml = renderToString(React.createElement(App))
assert('App renders without errors', appHtml.length > 1000)
assert('App includes dropdown triggers', appHtml.includes('aria-haspopup'))

console.log(`\n=== ${pass} passed, ${fail} failed ===\n`)
fs.unlinkSync(tmpFile)
process.exit(fail === 0 ? 0 : 1)
