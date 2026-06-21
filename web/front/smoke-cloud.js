// Verify cloud coverage + provider list after the changes
import { WEATHER_PROVIDERS, getProvider } from './src/utils/weatherProviders.js'
import { getCoords } from './src/data/cityCoordinates.js'

const r = await fetch('http://127.0.0.1:8081/events')
const outages = await r.json()
const sample = outages[2]
const coords = getCoords(sample.city)

console.log('=== Provider list ===')
WEATHER_PROVIDERS.forEach((p) => console.log(`  ${p.id.padEnd(12)} ${p.name}`))
console.log(`  Total: ${WEATHER_PROVIDERS.length}\n`)

console.log('=== Cloud coverage test (آمل outage) ===')
for (const p of WEATHER_PROVIDERS) {
  try {
    const w = await p.fetch({
      city: sample.city,
      lat: coords.latitude,
      lon: coords.longitude,
      startISO: sample.start,
      endISO: sample.end,
    })
    if (w.available) {
      console.log(
        `  ${p.name.padEnd(12)} → temp ${w.avgTemp?.toFixed(1)}°C, humidity ${w.avgHumidity?.toFixed(0)}%, cloud ${w.avgCloud?.toFixed(0)}% (peak ${w.peakCloud?.toFixed(0)}%, min ${w.minCloud?.toFixed(0)}%)`
      )
      console.log(`  sample cloud values: ${w.sampled.slice(0, 3).map(s => s.cloud).join(', ')}`)
    } else {
      console.log(`  ${p.name.padEnd(12)} → unavailable (${w.reason})`)
    }
  } catch (e) {
    console.log(`  ${p.name.padEnd(12)} → error: ${e.message}`)
  }
}
