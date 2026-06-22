// Verify the CityInstaller logic for manifest URL generation
console.log('=== Manifest URL generation ===')

function generateManifestUrl(city) {
  const name = encodeURIComponent(city)
  // Double-encode the inner Persian so that the URL parser + Go's
  // c.QueryParam() decode leaves us with "?city=URL_ENCODED_PERSIAN"
  // — and start_url "/" + query is then re-decoded by the browser on
  // app launch, giving us the correct "?city=آمل" URL.
  const query = encodeURIComponent(`?city=${encodeURIComponent(city)}`)
  return `/manifest.json?name=${name}&query=${query}`
}

// Test with various cities
const cities = ['آمل', 'ساری', 'بابلسر', 'سوادکوه شمالی', 'فریدون‌کنار']
for (const city of cities) {
  const url = generateManifestUrl(city)
  console.log(`  ${city.padEnd(20)} → ${url}`)

  // Verify the URL decodes correctly through one round-trip
  const parsed = new URL(`http://x${url}`)
  const name = parsed.searchParams.get('name')
  const query = parsed.searchParams.get('query')
  // name should decode fully to Persian
  // query should decode to "?city=<URL_ENCODED_PERSIAN>" (one level of
  // encoding remains — that's intentional)
  const isNameOk = name === city
  const isQueryOk = query === `?city=${encodeURIComponent(city)}`
  console.log(`     decoded name:  ${name}`)
  console.log(`     decoded query: ${query}`)
  console.log(`     ${isNameOk && isQueryOk ? '✓ PASS' : '✗ FAIL'}`)
}

// Test with mock browser
console.log('\n=== Simulated browser flow ===')

const { JSDOM } = await import('jsdom')
const dom = new JSDOM('<!DOCTYPE html><html><head></head><body></body></html>', { url: 'http://localhost/' })

global.window = dom.window
global.document = dom.window.document
global.navigator = dom.window.navigator

// Simulate a <link rel="manifest"> in the DOM
const link = document.createElement('link')
link.rel = 'manifest'
link.href = '/manifest.json'
document.head.appendChild(link)
console.log(`Initial manifest href: ${link.href}`)

// Simulate clicking "Install for آمل"
const city = 'آمل'
const newHref = generateManifestUrl(city)
link.href = newHref
console.log(`After click: ${link.href}`)

// Verify what would be fetched
console.log(`Would request: http://localhost${newHref}`)

// Simulate the server response
console.log('\n=== Mock manifest response ===')
const mockManifest = {
  name: `قطعی برق ${city}`,
  short_name: 'قطعی برق شمال',
  description: 'نمایش زنده قطعی‌های برنامه‌ریزی‌شده برق مازندران با اطلاعات آب و هوا',
  start_url: `/${decodeURIComponent(new URL(`http://x${newHref}`).searchParams.get('query') || '')}`,
  scope: '/',
  display: 'standalone',
  orientation: 'any',
  background_color: '#070b14',
  theme_color: '#0c1220',
  lang: 'fa',
  dir: 'rtl',
  categories: ['utilities', 'productivity'],
  icons: [
    { src: '/icon-192.svg', sizes: '192x192', type: 'image/svg+xml', purpose: 'any' },
    { src: '/icon-512.svg', sizes: '512x512', type: 'image/svg+xml', purpose: 'any maskable' },
  ],
}

console.log(JSON.stringify(mockManifest, null, 2))

console.log('\n=== Done ===')
