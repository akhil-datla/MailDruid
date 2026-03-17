package user

import (
	"testing"
)

func TestParseTimeRFC3339(t *testing.T) {
	result, err := parseTime("2025-01-15T10:30:00Z")
	if err != nil {
		t.Fatalf("parseTime error: %v", err)
	}
	if result.Year() != 2025 || result.Month() != 1 || result.Day() != 15 {
		t.Errorf("unexpected time: %v", result)
	}
}

func TestParseTimeDateOnly(t *testing.T) {
	result, err := parseTime("2025-06-15")
	if err != nil {
		t.Fatalf("parseTime error: %v", err)
	}
	if result.Year() != 2025 || result.Month() != 6 || result.Day() != 15 {
		t.Errorf("unexpected time: %v", result)
	}
}

func TestParseTimeInvalidFormat(t *testing.T) {
	_, err := parseTime("not-a-date")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestParseTimeDateTimeNoTimezone(t *testing.T) {
	result, err := parseTime("2025-03-10T14:00:00")
	if err != nil {
		t.Fatalf("parseTime error: %v", err)
	}
	if result.Hour() != 14 {
		t.Errorf("expected hour 14, got %d", result.Hour())
	}
}
