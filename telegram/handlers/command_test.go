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
		{name: "plain", text: searchCommand, command: searchCommand, want: false},
		{name: "basic", text: "/search", command: searchCommand, want: true},
		{name: "with args", text: "/search foo", command: searchCommand, want: true},
		{name: "with mention", text: "/search@north_outage_bot foo", command: searchCommand, want: true},
		{name: "different command", text: "/help", command: searchCommand, want: false},
		{name: "mid text", text: "hello /search foo", command: searchCommand, want: false},
	}

	for _, tt := range tests {
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
		{name: "simple", text: "/search foo bar", command: searchCommand, want: "foo bar"},
		{name: "leading space", text: "  /search   foo", command: searchCommand, want: "foo"},
		{name: "mention", text: "/search@north_outage_bot foo", command: searchCommand, want: "foo"},
		{name: "no args", text: "/search", command: searchCommand, want: ""},
		{name: "wrong command", text: "/help foo", command: searchCommand, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := commandArgument(tt.text, tt.command)
			if got != tt.want {
				t.Fatalf("commandArgument(%q, %q) = %q, want %q", tt.text, tt.command, got, tt.want)
			}
		})
	}
}
