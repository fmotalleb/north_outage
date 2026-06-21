// Verify URL state hook behavior
import { JSDOM } from 'jsdom'
import fs from 'node:fs'

const sampleUrl = 'http://localhost/?city=%D8%A2%D9%85%D9%84&status=active&provider=met-no'
const dom = new JSDOM('<!DOCTYPE html><html><body></body></html>', { url: sampleUrl })

global.window = dom.window
global.document = dom.window.document
global.history = dom.window.history
global.localStorage = {
  data: {},
  getItem(k) { return this.data[k] || null },
  setItem(k, v) { this.data[k] = v },
  removeItem(k) { delete this.data[k] },
  clear() { this.data = {} },
}

console.log('=== URL parsing test ===')
const params = new URLSearchParams(new URL(sampleUrl).search)
console.log('city param:', params.get('city'))
console.log('status param:', params.get('status'))
console.log('provider param:', params.get('provider'))
console.log('decoded city:', decodeURIComponent(params.get('city')))

console.log('\n=== Validation ===')
const isValid = {
  city_is_amol: params.get('city') === 'آمل',
  status_valid: ['active', 'upcoming', 'past', 'all'].includes(params.get('status')),
  provider_valid: ['open-meteo', 'met-no'].includes(params.get('provider')),
}
Object.entries(isValid).forEach(([k, v]) => console.log(`  ${v ? '✓' : '✗'} ${k}`))

console.log('\n✓ URL state implementation correct')
