// Simulate the React app lifecycle: load data, then render FilterBar
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
const FilterBar = mod.FilterBar
const React = (await import('react')).default
const { renderToString } = await import('react-dom/server')

const html = renderToString(
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

console.log('=== Dropdown renders in FilterBar ===')
console.log(`HTML length: ${html.length} bytes`)
console.log(`aria-haspopup count: ${(html.match(/aria-haspopup/g) || []).length}`)
console.log(`listbox count: ${(html.match(/listbox/g) || []).length}`)
console.log(`chevron SVG count: ${(html.match(/6 9 12 15 18 9/g) || []).length}`)
console.log(`label count: ${(html.match(/<label/g) || []).length}`)
console.log('')
console.log('=== Trigger button styles ===')
const triggerStyles = html.match(/bg-white\/5[^"]*/g) || []
console.log(`bg-white/5 occurrences: ${triggerStyles.length}`)
console.log('')
console.log('=== Selected values ===')
['همه شهرها', 'همه وضعیت‌ها', 'همه تاریخ‌ها', 'نزدیک‌ترین'].forEach((s) => {
  console.log(`  "${s}": ${html.includes(s) ? '✓' : '✗'}`)
})

fs.unlinkSync(tmpFile)
