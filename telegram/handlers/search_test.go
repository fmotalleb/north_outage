package handlers

import "testing"

func TestTruncateQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "short stays", query: "خیابان آزادی", want: "خیابان آزادی"},
		{name: "empty stays", query: "", want: ""},
		{name: "trimmed", query: "  foo  ", want: "foo"},
		{name: "truncated", query: "12345678901234567890ABCDE", want: "12345678901234567890"},
		{name: "truncated multibyte", query: "الفبپتثجچحخدذرزژسشصضطظع", want: "الفبپتثجچحخدذرزژسشصض"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := truncateQuery(tt.query); got != tt.want {
				t.Fatalf("truncateQuery(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}
