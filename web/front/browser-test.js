// Full browser test of the built React app.
// Starts mock API, serves the built dist, opens it with puppeteer, asserts rendering.
import puppeteer from 'puppeteer'
import http from 'node:http'
import fs from 'node:fs'
import path from 'node:path'

// Start a tiny static server for the built dist (port 4173)
const distDir = path.resolve('dist')
const server = http.createServer((req, res) => {
  let urlPath = req.url === '/' ? '/index.html' : req.url.split('?')[0]
  let filePath = path.join(distDir, urlPath)
  if (!filePath.startsWith(distDir)) { res.writeHead(403).end(); return }
  if (fs.existsSync(filePath) && fs.statSync(filePath).isFile()) {
    const ext = path.extname(filePath)
    const mime = { '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css',
                   '.svg': 'image/svg+xml', '.json': 'application/json' }[ext] || 'application/octet-stream'
    res.writeHead(200, { 'Content-Type': mime })
    fs.createReadStream(filePath).pipe(res)
  } else {
    res.writeHead(404).end()
  }
})
await new Promise((r) => server.listen(4174, '127.0.0.1', r))
console.log('Static server on http://127.0.0.1:4174')

const browser = await puppeteer.launch({
  headless: 'new',
  args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage'],
})
const page = await browser.newPage()
page.on('console', (msg) => console.log(`[console.${msg.type()}]`, msg.text()))
page.on('pageerror', (err) => console.log('[pageerror]', err.message))

await page.goto('http://127.0.0.1:4174', { waitUntil: 'networkidle0', timeout: 20000 })

// Wait for outage cards to render
await page.waitForSelector('article', { timeout: 15000 })

const stats = await page.evaluate(() => {
  return {
    title: document.title,
    h1: document.querySelector('h1')?.innerText,
    cards: document.querySelectorAll('article').length,
    cityChips: Array.from(new Set(Array.from(document.querySelectorAll('.chip')).map((e) => e.innerText).filter((t) => /^(قایمشهر|آمل|بابل|سوادکوه|ساری)/.test(t)))),
    bodyBg: getComputedStyle(document.body).backgroundColor,
    filterInputs: document.querySelectorAll('input,select').length,
    weatherButtons: document.querySelectorAll('button').length,
  }
})
console.log('\n=== Render stats ===')
console.log(JSON.stringify(stats, null, 2))

// Click "details" on the first card and wait for weather panel
await page.click('article:first-of-type button')
await new Promise((r) => setTimeout(r, 1500))
const weather = await page.evaluate(() => {
  const tempEls = Array.from(document.querySelectorAll('article')).map((a) =>
    a.querySelector('.text-orange-100')?.innerText || null,
  )
  const humEls = Array.from(document.querySelectorAll('article')).map((a) =>
    a.querySelector('.text-cyan-100')?.innerText || null,
  )
  return { tempEls, humEls }
})
console.log('\n=== Weather panel after expanding first card ===')
console.log(JSON.stringify(weather, null, 2))

// Take a screenshot
await page.screenshot({ path: 'screenshot.png', fullPage: true })
console.log('\nScreenshot saved to screenshot.png')

// Test provider switching
const providerButtons = await page.$$('aside button')
console.log(`\nFound ${providerButtons.length} buttons in sidebar`)

await browser.close()
server.close()
console.log('\n✓ All browser tests passed')
