// Coordinates for cities in Mazandaran province, Iran (IRST, +03:30).
// Used as a fallback when geocoding fails. Keys MUST match the `city` field
// returned by the /api/events endpoint (Persian).
export const CITY_COORDS = {
  "آمل": {
    "latitude": 36.4696,
    "longitude": 52.3507,
    "name_en": "Amol"
  },
  "بابل": {
    "latitude": 36.5513,
    "longitude": 52.6789,
    "name_en": "Babol"
  },
  "بابلسر": {
    "latitude": 36.6921,
    "longitude": 52.6872,
    "name_en": "Babolsar"
  },
  "بهشهر": {
    "latitude": 36.6924,
    "longitude": 53.5526,
    "name_en": "Behshahr"
  },
  "جویبار": {
    "latitude": 36.6411,
    "longitude": 52.9125,
    "name_en": "Juybar"
  },
  "ساری": {
    "latitude": 36.5633,
    "longitude": 53.0600,
    "name_en": "Sari"
  },
  "سوادکوه شمالی": {
    "latitude": 36.2919,
    "longitude": 52.8806,
    "name_en": "Savadkuh Shomali"
  },
  "سوادکوه": {
    "latitude": 36.1167,
    "longitude": 53.0500,
    "name_en": "Savadkuh"
  },
  "سیمرغ": {
    "latitude": 36.5806,
    "longitude": 52.9000,
    "name_en": "Simorgh"
  },
  "فریدونکنار": {
    "latitude": 36.6867,
    "longitude": 52.5225,
    "name_en": "Fereydunkenar"
  },
  "قایمشهر": {
    "latitude": 36.4631,
    "longitude": 52.8595,
    "name_en": "Qaemshahr"
  },
  "گلوگاه": {
    "latitude": 36.7278,
    "longitude": 53.8083,
    "name_en": "Galugah"
  },
  "میاندرود": {
    "latitude": 36.5640,
    "longitude": 53.1930,
    "name_en": "Miandorud"
  },
  "نکا": {
    "latitude": 36.6519,
    "longitude": 53.2990,
    "name_en": "Neka"
  }
}

// Returns an array of all known cities, sorted by Persian collation.
export function getKnownCities() {
  return Object.keys(CITY_COORDS).sort((a, b) => a.localeCompare(b, 'fa'))
}

export function getCoords(city) {
  if (!city) return null
  return CITY_COORDS[city] || null
}
