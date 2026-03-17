package user

import (
	"fmt"
	"time"
)

var supportedFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
	"2006-01-02",
}

func parseTime(s string) (time.Time, error) {
	for _, format := range supportedFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %s (use RFC3339)", s)
}
