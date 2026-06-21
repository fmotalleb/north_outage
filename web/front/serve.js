// Tiny zero-dependency static server for the single-file build.
// Usage: node serve.js [port]
import http from 'node:http'
import fs from 'node:fs'
import path from 'node:path'

const PORT = Number(process.argv[2]) || 4175
const DIST = path.resolve('dist')

const server = http.createServer((req, res) => {
  // Always serve index.html (it's a single-file SPA)
  const filePath = path.join(DIST, 'index.html')
  if (!filePath.startsWith(DIST)) {
    res.writeHead(403).end('Forbidden')
    return
  }
  const buf = fs.readFileSync(filePath)
  res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8', 'Cache-Control': 'no-store' })
  res.end(buf)
})

server.listen(PORT, '127.0.0.1', () => {
  console.log(`\n  Outage Tracker running at:\n  http://127.0.0.1:${PORT}/\n`)
  console.log('  Press Ctrl+C to stop.\n')
})
