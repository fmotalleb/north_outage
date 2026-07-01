package handlers

import (
	"testing"

	"github.com/go-telegram/bot/models"
)

func TestIsCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		text    string
		command string
		want    bool
	}{
		{name: "plain", text: "search", command: "search", want: false},
		{name: "basic", text: "/search", command: "search", want: true},
		{name: "with args", text: "/search foo", command: "search", want: true},
		{name: "with mention", text: "/search@north_outage_bot foo", command: "search", want: true},
		{name: "different command", text: "/help", command: "search", want: false},
		{name: "mid text", text: "hello /search foo", command: "search", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isCommand(&models.Update{
				Message: &models.Message{Text: tt.text},
			}, tt.command)
			if got != tt.want {
				t.Fatalf("isCommand(%q, %q) = %v, want %v", tt.text, tt.command, got, tt.want)
			}
		})
	}
}

func TestCommandArgument(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		text    string
		command string
		want    string
	}{
		{name: "simple", text: "/search foo bar", command: "search", want: "foo bar"},
		{name: "leading space", text: "  /search   foo", command: "search", want: "foo"},
		{name: "mention", text: "/search@north_outage_bot foo", command: "search", want: "foo"},
		{name: "no args", text: "/search", command: "search", want: ""},
		{name: "wrong command", text: "/help foo", command: "search", want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := commandArgument(tt.text, tt.command)
			if got != tt.want {
				t.Fatalf("commandArgument(%q, %q) = %q, want %q", tt.text, tt.command, got, tt.want)
			}
		})
	}
}
