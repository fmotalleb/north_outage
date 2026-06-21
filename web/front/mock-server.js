// Tiny mock server to simulate the user's API at http://127.0.0.1:8081/events
// Run with: node mock-server.js
import http from 'node:http'

const SAMPLE = [
  {
    id: 2,
    unique_hash: '907908047657215d24b8f9ce7ba5d5475a82cb4b6bb915e9d245236ee78f755a',
    city: 'قایمشهر',
    address: 'حد فاصل ابتدای جاده جویبار تا روستای نوده سمت چپ خیابان-چمازکتی',
    start: '2026-06-21T11:00:00+03:30',
    end: '2026-06-21T13:00:00+03:30',
    created_at: '2026-06-20T08:52:07.303667882+03:30',
  },
  {
    id: 3,
    unique_hash: 'cf1a1003bd7d11387283df28d301c7d9cef851822cb85e34d3849edf34849082',
    city: 'آمل',
    address: 'شهرک صنعتی امامزاده عبدالله شرکت سارو',
    start: '2026-06-20T08:00:00+03:30',
    end: '2026-06-20T23:55:00+03:30',
    created_at: '2026-06-20T08:52:07.303813653+03:30',
  },
  {
    id: 5,
    unique_hash: '0452071bee53ea553d68cdf7113660d913cb6a85ab9a164a85e6801878473833',
    city: 'آمل',
    address: 'شهرک صنعتی امامزاده عبدالله شرکت میناگران',
    start: '2026-06-20T08:00:00+03:30',
    end: '2026-06-20T23:55:00+03:30',
    created_at: '2026-06-20T08:52:07.30402412+03:30',
  },
  {
    id: 6,
    unique_hash: '960b90c29e332fdd29ab466b633ad9a584c06436c4a334d720f533073e018e4c',
    city: 'آمل',
    address: 'شهرک صنعتی امامزاده عبدالله شرکت صالح',
    start: '2026-06-20T08:00:00+03:30',
    end: '2026-06-20T23:55:00+03:30',
    created_at: '2026-06-20T08:52:07.304126243+03:30',
  },
  {
    id: 8,
    unique_hash: '860a649a14892dc168985b442ec5a0d44b53ff491c8808d5cbf321e64db3a7ad',
    city: 'بابل',
    address: 'شرکت پازوار بابل ملامین',
    start: '2026-06-20T08:00:00+03:30',
    end: '2026-06-20T23:59:00+03:30',
    created_at: '2026-06-20T08:52:07.304302937+03:30',
  },
  {
    id: 10,
    unique_hash: '714c661cfe2f63b37e7267e5381588b3493b2446f3311eb9e7c74f4d273144aa',
    city: 'بابل',
    address: 'روستاهای گله کلا -شمشیرمحله - مشهدی کلا - باریکلا - بوله کلا - میانرود- درویش خیل -منطقه شهری حاجی کلای رودبار و امیرکبیر',
    start: '2026-06-21T11:00:00+03:30',
    end: '2026-06-21T13:00:00+03:30',
    created_at: '2026-06-20T08:52:07.304469938+03:30',
  },
  {
    id: 13,
    unique_hash: 'e7f375b4db0e65e869bd0cd9c53d80a4a631edd6614423befd90fc998c5c1cd7',
    city: 'سوادکوه',
    address: 'ابتدای جاده خطیرکوه به سمت سر چلشک و روستاهای آریم و کنگلو - ملرد و امافت تا لردخطیرکوه و کلیه صنایع شن و ماسه تنگه خطیرکوه',
    start: '2026-06-21T13:00:00+03:30',
    end: '2026-06-21T15:00:00+03:30',
    created_at: '2026-06-20T08:52:07.304795303+03:30',
  },
]

const server = http.createServer((req, res) => {
  const cors = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET, OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type',
  }
  if (req.method === 'OPTIONS') {
    res.writeHead(204, cors).end()
    return
  }
  if (req.url === '/events' || req.url.startsWith('/events?')) {
    res.writeHead(200, { 'Content-Type': 'application/json; charset=utf-8', ...cors })
    res.end(JSON.stringify(SAMPLE))
    return
  }
  res.writeHead(404, cors).end('Not Found')
})

server.listen(8081, '127.0.0.1', () => {
  console.log('Mock outage server listening on http://127.0.0.1:8081/events')
})
