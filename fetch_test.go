package fetch

import (
	"testing"
)

func TestParseHeaderName(t *testing.T) {
	scenarios := map[string]bool{
		"":                false,
		" ":               false,
		"content-type":    true,
		"Accept-Language": true,
		" Content-Type":   false,
		":":               false,
		"\r":              false,
		"\n":              false,
	}
	for name, expectedSuccess := range scenarios {
		t.Run(name, func(t *testing.T) {
			_, err := parseHeaderName(name)
			if expectedSuccess && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !expectedSuccess && err == nil {
				t.Error("expected an error")
			}
		})
	}
}
