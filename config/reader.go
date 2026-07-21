package config

import (
	"context"

	"github.com/go-playground/validator/v10"

	parser "github.com/fmotalleb/go-tools/config"
	"github.com/fmotalleb/go-tools/decoder"
	"github.com/fmotalleb/go-tools/defaulter"
)

func ReadConfig(ctx context.Context, conf string) (*Config, error) {
	cfg := &Config{}
	var err error
	var raw map[string]any
	if raw, err = parser.ReadAndMergeConfig(ctx, conf); err != nil {
		return nil, err
	}
	if err = decoder.Decode(cfg, raw); err != nil {
		return nil, err
	}
	if err := defaulter.ApplyDefaults(cfg, cfg); err != nil {
		return nil, err
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err = validate.Struct(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
