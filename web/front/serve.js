// Tiny zero-dependency static server for the single-file build.
// Usage: node serve.js [port]
//
// Serves:
//   • /manifest.json              → dist/manifest.json  (PWA)
//   • /icon-192.svg, /icon-512.svg → dist/*.svg
//   • any other static file in dist/
//   • everything else             → dist/index.html     (SPA fallback so
//                                                       client-side routing
//                                                       still works)
import http from 'node:http'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const PORT = Number(process.argv[2]) || 4175
const DIST = path.resolve(__dirname, 'dist')

const MIME = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.webp': 'image/webp',
  '.json': 'application/json; charset=utf-8',
  '.ico': 'image/x-icon',
  '.txt': 'text/plain; charset=utf-8',
}

function send(res, status, body, contentType) {
  res.writeHead(status, {
    'Content-Type': contentType || 'application/octet-stream',
    'Cache-Control': 'no-store',
    'Content-Length': Buffer.byteLength(body),
  })
  res.end(body)
}

const server = http.createServer((req, res) => {
  let urlPath = decodeURIComponent(req.url.split('?')[0])
  if (urlPath === '/') urlPath = '/index.html'

  // Try to serve an actual file from dist/
  const safePath = path.normalize(urlPath).replace(/^(\.\.[\/\\])+/, '')
  const filePath = path.join(DIST, safePath)

  if (filePath.startsWith(DIST) && fs.existsSync(filePath) && fs.statSync(filePath).isFile()) {
    const ext = path.extname(filePath).toLowerCase()
    const buf = fs.readFileSync(filePath)
    return send(res, 200, buf, MIME[ext])
  }

  // SPA fallback → always serve index.html
  const indexPath = path.join(DIST, 'index.html')
  if (fs.existsSync(indexPath)) {
    const buf = fs.readFileSync(indexPath)
    return send(res, 200, buf, MIME['.html'])
  }

  send(res, 404, 'Not Found', 'text/plain; charset=utf-8')
})

server.listen(PORT, '127.0.0.1', () => {
  console.log(`\n  Outage Tracker running at:\n  http://127.0.0.1:${PORT}/\n`)
  console.log('  Press Ctrl+C to stop.\n')
})
