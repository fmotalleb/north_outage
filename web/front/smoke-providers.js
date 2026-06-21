// Test all 3 weather providers with the same outage
import { WEATHER_PROVIDERS, getProvider } from './src/utils/weatherProviders.js'
import { getCoords } from './src/data/cityCoordinates.js'

const r = await fetch('http://127.0.0.1:8081/events')
const outages = await r.json()
const sample = outages[2] // آمل outage
const coords = getCoords(sample.city)

console.log(`Testing all providers with: ${sample.city} (${sample.start} → ${sample.end})`)
console.log(`coords: ${coords.latitude}, ${coords.longitude}\n`)

for (const p of WEATHER_PROVIDERS) {
  process.stdout.write(`  ${p.name.padEnd(12)} … `)
  try {
    const w = await p.fetch({
      city: sample.city,
      lat: coords.latitude,
      lon: coords.longitude,
      startISO: sample.start,
      endISO: sample.end,
    })
    if (w.available) {
      console.log(`✓ temp ${w.avgTemp?.toFixed(1)}°C, humidity ${w.avgHumidity?.toFixed(0)}% (${w.sampled.length} samples)`)
    } else {
      console.log(`✗ unavailable: ${w.reason}`)
    }
  } catch (e) {
    console.log(`✗ ${e.message}`)
  }
}
