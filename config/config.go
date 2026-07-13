package config

import (
	"net/url"
	"time"

	// Autoload .env file.
	_ "github.com/joho/godotenv/autoload"

	sc "github.com/fmotalleb/scrapper-go/config"
)

type Config struct {
	HTTPListenAddr string `mapstructure:"http_listen" env:"HTTP_LISTEN"`

	Telegram   Telegram   `mapstructure:"telegram"`
	Mattermost Mattermost `mapstructure:"mattermost"`

	DatabaseConnection string `mapstructure:"database" env:"DATABASE" default:"sqlite:///outage.db" validate:"required,uri"`

	CollectCycle    string             `mapstructure:"collect_cycle" env:"COLLECT_CRON" default:"0 0 * * * *" validate:"required,cron"`
	CollectTimeout  time.Duration      `mapstructure:"collect_timeout" env:"COLLECT_TIMEOUT" default:"1h"`
	CollectorConfig sc.ExecutionConfig `mapstructure:"collector"`

	CollectOnStart          *bool         `mapstructure:"collect_on_start" env:"COLLECT_ON_START" default:"true"`
	CollectOnStartThreshold time.Duration `mapstructure:"collect_on_start_threshold" env:"COLLECT_ON_START_THRESHOLD" default:"10m"`

	RotateAfter time.Duration `mapstructure:"max_age" env:"MAX_AGE" default:"1h"`

	NotifyBefore time.Duration `mapstructure:"notify_before" env:"NOTIFY_BEFORE" default:"15m"`

	Weather Weather `mapstructure:"weather"`
}

type Telegram struct {
	BotKey   string        `mapstructure:"key" env:"TELEGRAM_BOT"`
	Timeout  time.Duration `mapstructure:"timeout" env:"TELEGRAM_BOT_TIMEOUT" default:"30s" validate:"required"`
	Proxy    *url.URL      `mapstructure:"proxy" env:"TELEGRAM_BOT_PROXY"`
	Endpoint url.URL       `mapstructure:"api" env:"TELEGRAM_BOT_ENDPOINT" default:"https://api.telegram.org" validate:"required"`
}

type Mattermost struct {
	BotToken     string        `mapstructure:"bot_token" env:"MATTERMOST_BOT_TOKEN"`
	ServerURL    string        `mapstructure:"server_url" env:"MATTERMOST_SERVER_URL" validate:"omitempty,url"`
	PublicURL    string        `mapstructure:"public_url" env:"MATTERMOST_PUBLIC_URL" validate:"omitempty,url"`
	CommandToken string        `mapstructure:"command_token" env:"MATTERMOST_COMMAND_TOKEN"`
	Timeout      time.Duration `mapstructure:"timeout" env:"MATTERMOST_TIMEOUT" default:"30s" validate:"required"`
}

type Weather struct {
	Proxy  *url.URL `mapstructure:"proxy" env:"WEATHER_PROXY"`
	Notify bool     `mapstructure:"notify" env:"WEATHER_NOTIFY" default:"false"`
}
