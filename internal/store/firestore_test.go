package store

import (
	"testing"
)

func TestHashURL(t *testing.T) {
	a := hashURL("https://example.com")
	b := hashURL("https://example.com")
	c := hashURL("https://example.org")

	if a != b {
		t.Fatal("expected same URL to produce same hash")
	}
	if a == c {
		t.Fatal("expected different URLs to produce different hashes")
	}
	if len(a) != 64 {
		t.Fatalf("expected SHA-256 hex length 64, got %d", len(a))
	}
}
