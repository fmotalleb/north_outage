// End-to-end smoke test of the React app's data pipeline:
// 1) Fetch /events (via running mock server)
// 2) Pass each outage through the same weather fetch logic
// 3) Print the resolved weather summary
import { getProvider } from './src/utils/weatherProviders.js'
import { getCoords } from './src/data/cityCoordinates.js'

const r = await fetch('http://127.0.0.1:8081/events')
const outages = await r.json()
console.log('Fetched', outages.length, 'outages\n')

const provider = getProvider('open-meteo')

for (const o of outages) {
  const coords = getCoords(o.city)
  console.log(`\n=== ${o.city} | ${o.address.slice(0, 60)}…`)
  console.log(`  start: ${o.start}`)
  console.log(`  end:   ${o.end}`)
  console.log(`  coords: ${coords ? `${coords.latitude}, ${coords.longitude}` : 'unknown'}`)
  try {
    const w = await provider.fetch({
      city: o.city,
      lat: coords?.latitude,
      lon: coords?.longitude,
      startISO: o.start,
      endISO: o.end,
    })
    if (w.available) {
      console.log(`  ✓ ${w.source} → avg temp ${w.avgTemp?.toFixed(1)}°C, avg humidity ${w.avgHumidity?.toFixed(0)}%`)
      console.log(`  samples: ${w.sampled.length}`)
    } else {
      console.log(`  ✗ ${w.source} unavailable: ${w.reason}`)
    }
  } catch (e) {
    console.log(`  ✗ error: ${e.message}`)
  }
}
