// jsdom render test - override fetch + serve dist files inline
import jsdomPkg from 'jsdom'
const { JSDOM } = jsdomPkg
import fs from 'node:fs'
import path from 'node:path'
import http from 'node:http'

const distDir = path.resolve('dist')

// Serve dist on a local port
const server = http.createServer((req, res) => {
  let filePath = path.join(distDir, req.url === '/' ? '/index.html' : req.url.split('?')[0])
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
await new Promise((r) => server.listen(4180, '127.0.0.1', r))

const SAMPLE = JSON.parse(fs.readFileSync('sample-data.json', 'utf8'))

// Build the URL the app will use. We need to make /events resolve to our SAMPLE.
// Strategy: use a JSDOM option `beforeParsing` to inject a fetch monkey-patch
// via a `<script>` tag into the HTML before jsdom parses it.
const html = fs.readFileSync(path.join(distDir, 'index.html'), 'utf8')
const patch = `
<script>
  (function(){
    const SAMPLE = ${JSON.stringify(SAMPLE)};
    // Define a fetch stub (jsdom may not have one)
    function stubFetch(input, init) {
      const u = typeof input === 'string' ? input : (input && input.url) || '';
      if (u.endsWith('/events') || u.includes('/events?')) {
        return Promise.resolve({
          ok: true, status: 200, statusText: 'OK',
          json: () => Promise.resolve(SAMPLE),
          text: () => Promise.resolve(JSON.stringify(SAMPLE)),
        });
      }
      // Try the real fetch if it exists, else 503
      if (typeof window.__realFetch === 'function') {
        return Promise.race([
          window.__realFetch(input, init),
          new Promise((_, rej) => setTimeout(() => rej(new Error('timeout')), 8000)),
        ]).catch((e) => ({ ok: false, status: 503, statusText: 'timeout' }));
      }
      return Promise.resolve({ ok: false, status: 503, statusText: 'no-fetch' });
    }
    window.fetch = stubFetch;
  })();
</script>
`
const patchedHtml = html.replace('</head>', patch + '</head>')

const dom = new JSDOM(patchedHtml, {
  url: `http://127.0.0.1:4180/`,
  runScripts: 'dangerously',
  pretendToBeVisual: true,
  resources: 'usable',
})

dom.window.addEventListener('error', (e) => {
  console.log('[window.error]', e.message, e.filename, e.lineno)
})
dom.window.console.error = (...args) => console.log('[console.error]', ...args)
dom.window.console.warn = (...args) => console.log('[console.warn]', ...args)
dom.window.console.log = (...args) => console.log('[console.log]', ...args)

// Wait for React to render
await new Promise((r) => setTimeout(r, 6000))

const doc = dom.window.document
const articles = doc.querySelectorAll('article')
const h1 = doc.querySelector('h1')?.innerText
const inputs = doc.querySelectorAll('input, select')
const asideButtons = doc.querySelectorAll('aside button')

console.log('=== jsdom render report ===')
console.log('Title:', doc.title)
console.log('H1:', h1)
console.log('Article count:', articles.length)
console.log('Form inputs:', inputs.length)
console.log('Sidebar buttons:', asideButtons.length)
console.log()
const results = {
  articlesRendered: articles.length > 0,
  titleSet: doc.title.includes('قطعی'),
  headerRendered: !!(h1 && h1.includes('قطعی')),
  filterInputs: inputs.length >= 4,
  providerSelector: asideButtons.length >= 3,
}
for (const [k, v] of Object.entries(results)) {
  console.log(`  ${v ? '✓ PASS' : '✗ FAIL'}  ${k}`)
}

if (articles.length) {
  console.log('\nFirst card text (first 250 chars):')
  console.log('  ', articles[0].textContent.replace(/\s+/g, ' ').slice(0, 250))
}

const chips = Array.from(doc.querySelectorAll('.chip')).map((e) => e.innerText.trim())
console.log('\nSample chips:', chips.slice(0, 12).join(' | '))

dom.window.close()
server.close()
process.exit(Object.values(results).every(Boolean) ? 0 : 1)
