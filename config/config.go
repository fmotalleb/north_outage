package config

import (
	"bytes"
	"encoding/json"
	"net/url"
	"time"

	// Autoload .env file.
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	HTTPListenAddr string `mapstructure:"http_listen" env:"HTTP_LISTEN"`

	Telegram Telegram `mapstructure:"telegram"`

	DatabaseConnection string `mapstructure:"database" env:"DATABASE" default:"sqlite:///outage.db" validate:"required,uri"`

	CollectCycle    string        `mapstructure:"collect_cycle" env:"COLLECT_CRON" default:"0 0 * * * *" validate:"required,cron"`
	CollectTimeout  time.Duration `mapstructure:"collect_timeout" env:"COLLECT_TIMEOUT" default:"1h"`
	CollectorConfig Collector     `mapstructure:"collector"`

	CollectOnStart          *bool         `mapstructure:"collect_on_start" env:"COLLECT_ON_START" default:"true"`
	CollectOnStartThreshold time.Duration `mapstructure:"collect_on_start_threshold" env:"COLLECT_ON_START_THRESHOLD" default:"10m"`

	RotateAfter time.Duration `mapstructure:"max_age" env:"MAX_AGE" default:"1h"`

	NotifyBefore time.Duration `mapstructure:"notify_before" env:"NOTIFY_BEFORE" default:"15m"`

	Weather Weather `mapstructure:"weather"`

	Tracing Tracing `mapstructure:"tracing"`
}

// Tracing configures OpenTelemetry trace export. URL selects the protocol
// and transport security from its scheme: http:// and https:// use the OTLP
// HTTP exporter, grpc:// and grpcs:// use the OTLP gRPC exporter. When URL is
// empty no tracing is exported. Rate is the sampling ratio (0..1) applied to
// API and Telegram interactions only; collector traces are always exported.
type Tracing struct {
	URL     string         `mapstructure:"url" env:"TRACING_URL"`
	Rate    float64        `mapstructure:"rate" env:"TRACING_RATE" default:"1"`
	Headers TracingHeaders `mapstructure:"headers" env:"TRACING_HEADERS"`
}

// TracingHeaders holds extra headers sent to the OTLP collector with every
// export request (used for auth tokens such as Tempo's X-Scope-OrgID or
// Honeycomb's X-Honeycomb-Team). It is written as a TOML table in the config
// file or as a JSON object in TRACING_HEADERS, e.g.
//
//	TRACING_HEADERS='{"X-Scope-OrgID":"acme"}'
type TracingHeaders map[string]string

// UnmarshalText parses a JSON object into the header map, so the value can
// come from an environment variable or a JSON string in the config file.
func (t *TracingHeaders) UnmarshalText(text []byte) error {
	if len(bytes.TrimSpace(text)) == 0 {
		*t = nil
		return nil
	}
	var headers map[string]string
	if err := json.Unmarshal(text, &headers); err != nil {
		return err
	}
	*t = headers
	return nil
}

type Collector struct {
	Endpoint  string        `mapstructure:"endpoint" env:"COLLECTOR_ENDPOINT" default:"https://khamooshi.maztozi.ir/api/outages"`
	Timeout   time.Duration `mapstructure:"timeout" env:"COLLECTOR_TIMEOUT" default:"30s" validate:"required"`
	Proxy     *url.URL      `mapstructure:"proxy" env:"COLLECTOR_PROXY"`
	SSLVerify bool          `mapstructure:"ssl_verify" env:"COLLECTOR_SSL_VERIFY" default:"false"` // default explicitly set to false due to current state of the api ssl

	PlannedDuration   time.Duration `mapstructure:"planned_duration" env:"COLLECTOR_PLANNED_DURATION" default:"2h"`
	UnPlannedDuration time.Duration `mapstructure:"unplanned_duration" env:"COLLECTOR_UNPLANNED_DURATION" default:"5h"`
}

type Telegram struct {
	BotKey   string        `mapstructure:"key" env:"TELEGRAM_BOT"`
	Timeout  time.Duration `mapstructure:"timeout" env:"TELEGRAM_BOT_TIMEOUT" default:"30s" validate:"required"`
	Proxy    *url.URL      `mapstructure:"proxy" env:"TELEGRAM_BOT_PROXY"`
	Endpoint url.URL       `mapstructure:"api" env:"TELEGRAM_BOT_ENDPOINT" default:"https://api.telegram.org" validate:"required"`
	// MessageTTL controls how long menu-like bot messages (search results,
	// lists, confirmations) are kept before they are deleted, unless the user
	// keeps interacting with them.
	MessageTTL time.Duration `mapstructure:"message_ttl" env:"TELEGRAM_MESSAGE_TTL" default:"1m"`
}

type Weather struct {
	Proxy  *url.URL `mapstructure:"proxy" env:"WEATHER_PROXY"`
	Notify bool     `mapstructure:"notify" env:"WEATHER_NOTIFY" default:"false"`
}
