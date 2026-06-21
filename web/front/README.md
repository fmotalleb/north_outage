# قطعی برق · Outage Tracker

A modern React front-end for displaying planned power-outage events (Mazandaran, Iran).
Fetches data from `/api/events` (relative URL — works with any backend), and overlays
free, no-key weather data (temperature + humidity + cloud cover) onto each outage
in a modal popup with an animated sun/cloud icon.

## Features

- 🇮🇷 Full Persian (RTL) UI with the Vazirmatn font
- 🌗 Modern dark UI with gradient accents, glassmorphism, smooth transitions
- 🔎 Live filtering by city, status, search, date range, and sort
- 🔗 Filters persist in the URL — shareable links restore the exact view
- 💾 Last filter state restored on next visit via `localStorage`
- 📱 Installable as a PWA — `manifest.json` + standalone display mode
- 🌡️ Weather per outage (temp + humidity + cloud) from free, no-key providers:
  - **Open-Meteo** (default, best historical + forecast coverage)
  - **Met.no** (Norwegian Meteorological Institute, ~3 day forecast)
- 🪟 Weather opens in a modal with an animated sun/cloud icon

## Build

```bash
npm install
npm run build
```

Produces a `dist/` folder:

```
dist/
├── index.html     218 KB  (React + Tailwind + everything inlined)
├── manifest.json          (PWA manifest)
├── icon-192.svg           (PWA icon)
└── icon-512.svg           (PWA maskable icon)
```

## Run

```bash
# Serve the build with the included zero-dependency server
node serve.js              # http://127.0.0.1:4175/

# Or with vite dev mode (with API proxy)
npm run dev                # http://127.0.0.1:5173/

# Or with any other static server
cd dist && python3 -m http.server 8000
```

The server falls back to `dist/index.html` for any unknown route, so client-side
URL state (e.g. `/?city=آمل&status=active`) works correctly.

## Backend integration

Your HTTP server must serve:
- `GET /api/events` → JSON array of outage events (see `mock-server.js` for sample shape)
- `GET /manifest.json` → the PWA manifest
- `GET /icon-192.svg`, `/icon-512.svg` → PWA icons
- `GET /*` (any other route) → `dist/index.html` (SPA fallback)

### Expected `/api/events` response shape

```jsonc
[
  {
    "id": 2,
    "unique_hash": "907908047657215d24b8f9ce7ba5d5475a82cb4b6bb915e9d245236ee78f755a",
    "city": "قایمشهر",
    "address": "حد فاصل ابتدای جاده جویبار تا روستای نوده…",
    "start": "2026-06-21T11:00:00+03:30",   // RFC3339
    "end":   "2026-06-21T13:00:00+03:30",
    "created_at": "2026-06-20T08:52:07+03:30"
  }
]
```

## File layout

```
outage-app/
├── index.html                 # HTML shell (refs manifest, icons, fonts)
├── manifest.json              # PWA manifest (also copied to dist/)
├── public/                    # assets copied to dist/ on build
│   ├── manifest.json
│   ├── icon-192.svg
│   └── icon-512.svg
├── src/
│   ├── main.jsx
│   ├── App.jsx                # Root component
│   ├── index.css
│   ├── components/
│   │   ├── Header.jsx
│   │   ├── FilterBar.jsx
│   │   ├── OutageList.jsx
│   │   ├── OutageCard.jsx
│   │   ├── WeatherModal.jsx   # The sun/cloud icon lives here
│   │   ├── WeatherProviderSelector.jsx
│   │   └── Dropdown.jsx       # Custom styled select
│   ├── hooks/
│   │   ├── useOutages.js      # fetches /api/events
│   │   ├── useWeather.js      # fetches + caches weather
│   │   └── useUrlState.js     # URL + localStorage sync
│   ├── utils/
│   │   ├── weatherProviders.js  # Open-Meteo + Met.no adapters
│   │   └── dateUtils.js
│   └── data/
│       └── cityCoordinates.js # Fallback for ~23 Mazandaran cities
├── vite.config.js             # vite-plugin-singlefile config
├── tailwind.config.js
├── postcss.config.js
├── serve.js                   # zero-dep static server (also serves manifest/icons)
└── mock-server.js             # sample API for local testing
```

## Notes on weather data

- **Open-Meteo** — geocodes the city, then queries `archive-api.open-meteo.com`
  (past outages) or `api.open-meteo.com` (upcoming outages) for hourly
  temperature, humidity and cloud cover. No API key required.
- **Met.no** — requires coordinates. Returns ~3 day forecasts. Past outages
  may have no data. Cloud cover is already in percent.

All three providers require no API key. The frontend talks to them directly
from the browser; nothing is proxied or stored on a server.
