package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

// ── Data types ──────────────────────────────────────────────────────────────

// WeatherData holds the result of a weather query, matching the frontend shape.
type WeatherData struct {
	Source      string   `json:"source"`
	Available   bool     `json:"available"`
	AvgTemp     *float64 `json:"avg_temp,omitempty"`
	AvgHumidity *float64 `json:"avg_humidity,omitempty"`
	AvgCloud    *float64 `json:"avg_cloud,omitempty"`
	Note        string   `json:"note,omitempty"`
}

// ── HTTP client (configurable via Init) ───────────────────────────────────

var httpClient = &http.Client{Timeout: 15 * time.Second}

// Init configures the weather package, including the HTTP client used for API
// calls. If proxyURL is non-nil the client will route requests through the
// given SOCKS5 (or other proxy.Dialer-compatible) proxy.
func Init(proxyURL *url.URL) {
	if proxyURL == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
		return
	}
	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		// Fall back to direct if proxy setup fails.
		httpClient = &http.Client{Timeout: 15 * time.Second}
		return
	}
	transport := &http.Transport{Dial: dialer.Dial}
	httpClient = &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}
}

// ── In-memory cache (by city) ──────────────────────────────────────────────

var (
	cacheMu    sync.Mutex
	cache      = make(map[string]*cacheEntry)
	cacheTTL   = 30 * time.Minute
)

type cacheEntry struct {
	data      *WeatherData
	expiresAt time.Time
}

// GetWeather returns weather data for the given city and time range.
// It uses an in-memory cache keyed by city name to avoid redundant API calls.
// Only successful responses are cached; transient errors are retried on the next call.
func GetWeather(ctx context.Context, city string, start, end time.Time) *WeatherData {
	cacheMu.Lock()
	if entry, ok := cache[city]; ok && time.Now().Before(entry.expiresAt) {
		cacheMu.Unlock()
		return entry.data
	}
	cacheMu.Unlock()

	data := fetchOpenMeteo(ctx, city, start, end)

	// Only cache successful responses — transient errors get retried.
	if data.Available {
		cacheMu.Lock()
		cache[city] = &cacheEntry{data: data, expiresAt: time.Now().Add(cacheTTL)}
		cacheMu.Unlock()
	}

	return data
}

// FormatWeatherLine returns a one-line string summarising the weather data,
// suitable for appending to a notification message. Returns "" when no data.
func FormatWeatherLine(w *WeatherData) string {
	if w == nil || !w.Available {
		return ""
	}
	parts := ""
	if w.AvgTemp != nil {
		parts += fmt.Sprintf(" 🌡%.0f°C", *w.AvgTemp)
	}
	if w.AvgHumidity != nil {
		parts += fmt.Sprintf(" 💧%.0f%%", *w.AvgHumidity)
	}
	if w.AvgCloud != nil {
		parts += fmt.Sprintf(" ☁%.0f%%", *w.AvgCloud)
	}
	return parts
}

// ── City coordinates (Mazandaran province, Iran) ───────────────────────────

// cityCoords mirrors web/front/src/data/cityCoordinates.js
var cityCoords = map[string]struct {
	Lat float64
	Lon float64
}{
	"آمل":         {36.4696, 52.3507},
	"بابل":        {36.5513, 52.6789},
	"بابلسر":      {36.6921, 52.6872},
	"بهشهر":       {36.6924, 53.5526},
	"جویبار":      {36.6411, 52.9125},
	"ساری":        {36.5633, 53.0600},
	"سوادکوه شمالی": {36.2919, 52.8806},
	"سوادکوه":     {36.1167, 53.0500},
	"سیمرغ":       {36.5806, 52.9000},
	"فریدونکنار":  {36.6867, 52.5225},
	"قایمشهر":     {36.4631, 52.8595},
	"گلوگاه":      {36.7278, 53.8083},
	"میاندرود":    {36.5640, 53.1930},
	"نکا":         {36.6519, 53.2990},
}

func getCoords(city string) (lat, lon float64, ok bool) {
	c, found := cityCoords[city]
	if !found {
		return 0, 0, false
	}
	return c.Lat, c.Lon, true
}

// ── Open-Meteo fetcher ─────────────────────────────────────────────────────

const (
	forecastAPI = "https://api.open-meteo.com/v1/forecast"
	archiveAPI  = "https://archive-api.open-meteo.com/v1/archive"
)

func toDateStr(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

func fetchOpenMeteo(ctx context.Context, city string, start, end time.Time) *WeatherData {
	now := time.Now()

	lat, lon, found := getCoords(city)
	if !found {
		return &WeatherData{Source: "Open-Meteo", Available: false, Note: "مختصاتی برای این شهر یافت نشد"}
	}

	startStr := toDateStr(start)
	endStr := toDateStr(end)

	// Decide forecast vs archive.
	hoursAhead := end.Sub(now).Hours()
	isFuture := hoursAhead > 24
	hoursBehind := now.Sub(end).Hours()
	isPastTooFar := hoursBehind > 6*24

	base := forecastAPI
	if !isFuture && isPastTooFar {
		base = archiveAPI
	}

	u := fmt.Sprintf("%s?latitude=%.4f&longitude=%.4f&hourly=temperature_2m,relative_humidity_2m,cloud_cover&start_date=%s&end_date=%s&timezone=auto",
		base, lat, lon, startStr, endStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return &WeatherData{Source: "Open-Meteo", Available: false, Note: err.Error()}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return &WeatherData{Source: "Open-Meteo", Available: false, Note: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &WeatherData{Source: "Open-Meteo", Available: false, Note: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
	}

	var result struct {
		Hourly struct {
			Times                []string  `json:"time"`
			Temperature2m        []float64 `json:"temperature_2m"`
			RelativeHumidity2m   []float64 `json:"relative_humidity_2m"`
			CloudCover           []float64 `json:"cloud_cover"`
		} `json:"hourly"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &WeatherData{Source: "Open-Meteo", Available: false, Note: err.Error()}
	}

	times := result.Hourly.Times
	temps := result.Hourly.Temperature2m
	hums := result.Hourly.RelativeHumidity2m
	clouds := result.Hourly.CloudCover

	if len(times) == 0 {
		return &WeatherData{Source: "Open-Meteo", Available: false, Note: "empty response"}
	}

	startMs := start.UnixMilli()
	endMs := end.UnixMilli()

	var sampledTemps, sampledHums, sampledClouds []float64
	for i, t := range times {
		parsed, err := time.Parse(time.RFC3339, t)
		if err != nil {
			// Try ISO 8601
			parsed, err = time.Parse("2006-01-02T15:04", t)
			if err != nil {
				continue
			}
		}
		ts := parsed.UnixMilli()
		if ts >= startMs-600000 && ts <= endMs+600000 {
			if i < len(temps) {
				sampledTemps = append(sampledTemps, temps[i])
			}
			if i < len(hums) {
				sampledHums = append(sampledHums, hums[i])
			}
			if i < len(clouds) {
				sampledClouds = append(sampledClouds, clouds[i])
			}
		}
	}

	avg := func(vals []float64) *float64 {
		if len(vals) == 0 {
			return nil
		}
		sum := 0.0
		for _, v := range vals {
			sum += v
		}
		r := sum / float64(len(vals))
		return &r
	}

	return &WeatherData{
		Source:      "Open-Meteo",
		Available:   true,
		AvgTemp:     avg(sampledTemps),
		AvgHumidity: avg(sampledHums),
		AvgCloud:    avg(sampledClouds),
	}
}
