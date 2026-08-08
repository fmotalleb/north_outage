package config

import (
	"reflect"
	"testing"

	"github.com/fmotalleb/go-tools/decoder"
	"github.com/fmotalleb/go-tools/defaulter"
)

// decodeAndApply mirrors the flow used by ReadConfig so the tests exercise
// the same decoder hooks and env handling without needing a config file.
func decodeAndApply(t *testing.T, raw map[string]any) (*Config, error) {
	t.Helper()
	cfg := &Config{}
	if err := decoder.Decode(cfg, raw); err != nil {
		return nil, err
	}
	if err := defaulter.ApplyDefaults(cfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func TestTracingHeadersFromEnv(t *testing.T) {
	t.Setenv("TRACING_HEADERS", `{"X-Scope-OrgID":"acme","Authorization":"Bearer token"}`)
	cfg, err := decodeAndApply(t, map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{
		"X-Scope-OrgID": "acme",
		"Authorization": "Bearer token",
	}
	if got := map[string]string(cfg.Tracing.Headers); !reflect.DeepEqual(got, want) {
		t.Fatalf("headers mismatch:\n got %v\nwant %v", got, want)
	}
}

func TestTracingHeadersFromRawMap(t *testing.T) {
	raw := map[string]any{
		"tracing": map[string]any{
			"url":     "http://localhost:4318",
			"headers": map[string]any{"X-Scope-OrgID": "acme"},
		},
	}
	cfg, err := decodeAndApply(t, raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{"X-Scope-OrgID": "acme"}
	if got := map[string]string(cfg.Tracing.Headers); !reflect.DeepEqual(got, want) {
		t.Fatalf("headers mismatch:\n got %v\nwant %v", got, want)
	}
}

func TestTracingHeadersInvalidJSON(t *testing.T) {
	t.Setenv("TRACING_HEADERS", `{not json`)
	if _, err := decodeAndApply(t, map[string]any{}); err == nil {
		t.Fatal("expected an error for invalid TRACING_HEADERS json")
	}
}
