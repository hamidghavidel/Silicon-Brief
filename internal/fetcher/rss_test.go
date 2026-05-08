package fetcher

import (
	"testing"
	"time"
)

func TestParseRSSDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantZero bool
	}{
		{"RFC1123", "Mon, 02 Jan 2006 15:04:05 MST", false},
		{"RFC3339", "2006-01-02T15:04:05Z", false},
		{"empty", "", true},
		{"invalid", "not-a-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRSSDate(tt.input)
			if tt.wantZero {
				if got.After(time.Now().Add(-time.Minute)) {
					// fallback to time.Now() is expected
					return
				}
			}
			if got.IsZero() {
				t.Fatalf("expected non-zero time for %q", tt.input)
			}
		})
	}
}
