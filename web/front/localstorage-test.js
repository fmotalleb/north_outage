// Verify useLocalStorage hook behavior
import { JSDOM } from 'jsdom'

const dom = new JSDOM('<!DOCTYPE html><html><body></body></html>', { url: 'http://localhost/' })

global.window = dom.window
global.document = dom.window.document
global.navigator = dom.window.navigator
global.localStorage = dom.window.localStorage
// No history needed anymore
delete global.history

const React = (await import('react')).default
const { useEffect } = React
const ReactDOM = (await import('react-dom/client')).default

const { useLocalState } = await import('./src/hooks/useLocalStorage.js?ls=1')

console.log('=== Test 1: read default state (empty localStorage) ===')
function Test1() {
  const [state, update] = useLocalState('test.key', {
    city: { default: 'all' },
    status: { default: 'all', values: ['all', 'active'] },
  })
  global.__s1 = state
  global.__u1 = update
  return null
}
const r1 = ReactDOM.createRoot(document.body)
r1.render(React.createElement(Test1))
await new Promise((r) => setTimeout(r, 100))
console.log('state:', JSON.stringify(global.__s1))
console.log('Expected: {} (no value stored)')
console.log(global.__s1 && Object.keys(global.__s1).length === 0 ? '✓ PASS' : '✗ FAIL')

console.log('\n=== Test 2: write to state, persist to localStorage ===')
global.__u1({ city: 'آمل', status: 'active' })
await new Promise((r) => setTimeout(r, 100))
console.log('localStorage value:', localStorage.getItem('test.key'))
console.log('state:', JSON.stringify(global.__s1))
const stored = localStorage.getItem('test.key')
const expected = JSON.stringify({ city: 'آمل', status: 'active' })
console.log(stored === expected ? '✓ PASS' : '✗ FAIL')

console.log('\n=== Test 3: read from localStorage on re-mount ===')
function Test3() {
  const [state] = useLocalState('test.key', {
    city: { default: 'all' },
    status: { default: 'all', values: ['all', 'active'] },
  })
  global.__s3 = state
  return null
}
const r3 = ReactDOM.createRoot(document.body)
r3.render(React.createElement(Test3))
await new Promise((r) => setTimeout(r, 100))
console.log('state:', JSON.stringify(global.__s3))
console.log(global.__s3?.city === 'آمل' && global.__s3?.status === 'active' ? '✓ PASS' : '✗ FAIL')

console.log('\n=== Test 4: schema validation drops invalid values ===')
localStorage.setItem('test.key2', JSON.stringify({ city: 'بابل', status: 'INVALID_VALUE', sort: 'BAD' }))
function Test4() {
  const [state] = useLocalState('test.key2', {
    city: { default: 'all' },
    status: { default: 'all', values: ['all', 'active', 'past'] },
    sort: { default: 'start_asc', values: ['start_asc', 'start_desc'] },
  })
  global.__s4 = state
  return null
}
const r4 = ReactDOM.createRoot(document.body)
r4.render(React.createElement(Test4))
await new Promise((r) => setTimeout(r, 100))
console.log('state:', JSON.stringify(global.__s4))
console.log('Expected: city preserved, status/sort dropped')
const ok4 = global.__s4?.city === 'بابل' && !('status' in global.__s4) && !('sort' in global.__s4)
console.log(ok4 ? '✓ PASS' : '✗ FAIL')

console.log('\n=== Test 5: NO URL is being touched ===')
// Reset
localStorage.clear()
function Test5() {
  const [state, update] = useLocalState('test.key5', {})
  global.__u5 = update
  return null
}
const r5 = ReactDOM.createRoot(document.body)
r5.render(React.createElement(Test5))
await new Promise((r) => setTimeout(r, 50))
const urlBefore = window.location.href
global.__u5({ city: 'آمل' })
await new Promise((r) => setTimeout(r, 100))
const urlAfter = window.location.href
console.log('URL before:', urlBefore)
console.log('URL after :', urlAfter)
console.log(urlBefore === urlAfter ? '✓ PASS (no URL change)' : '✗ FAIL')

console.log('\n=== Test 6: cross-tab sync via storage event ===')
localStorage.clear()
function Test6() {
  const [state] = useLocalState('test.key6', {})
  global.__s6 = state
  return null
}
const r6 = ReactDOM.createRoot(document.body)
r6.render(React.createElement(Test6))
await new Promise((r) => setTimeout(r, 100))
console.log('initial state:', JSON.stringify(global.__s6))
// Simulate another tab writing
localStorage.setItem('test.key6', JSON.stringify({ city: 'ساری', status: 'past' }))
// Fire the storage event manually (jsdom doesn't fire it for setItem)
const SE = dom.window.StorageEvent
window.dispatchEvent(new SE('storage', {
  key: 'test.key6',
  newValue: JSON.stringify({ city: 'ساری', status: 'past' }),
}))
await new Promise((r) => setTimeout(r, 100))
console.log('after cross-tab update:', JSON.stringify(global.__s6))
const ok6 = global.__s6?.city === 'ساری' && global.__s6?.status === 'past'
console.log(ok6 ? '✓ PASS' : '✗ FAIL')

console.log('\n=== All localStorage tests done ===')
