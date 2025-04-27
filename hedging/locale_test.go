package hedging

import (
	"regexp"
	"testing"
)

func TestGetLocale(t *testing.T) {
	locale, err := getLocale()
	if err != nil {
		t.Fatalf("getLocale failed: %s", err)
	}

	expected := regexp.MustCompile("[a-z]{2}[-_]{1}[A-Z]{2}")
	if !expected.MatchString(locale) {
		t.Fatalf("locale format mismatch: %s", locale)
	}
}

func TestGetPrinter(t *testing.T) {
	printer, err := GetPrinter()
	if err != nil {
		t.Fatalf("GetPrinter failed: %s", err)
	}
	if printer == nil {
		t.Fatalf("GetPrinter unexpectedly returned nil")
	}
}
