package scheduler

import (
	"testing"
)

func TestRemoveFromSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		remove   string
		expected []string
	}{
		{"remove middle", []string{"a", "b", "c"}, "b", []string{"a", "c"}},
		{"remove first", []string{"a", "b", "c"}, "a", []string{"b", "c"}},
		{"remove last", []string{"a", "b", "c"}, "c", []string{"a", "b"}},
		{"remove nonexistent", []string{"a", "b"}, "z", []string{"a", "b"}},
		{"remove from single", []string{"a"}, "a", []string{}},
		{"remove from empty", []string{}, "a", []string{}},
		{"remove from nil", nil, "a", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeFromSlice(tt.input, tt.remove)
			if len(result) != len(tt.expected) {
				t.Errorf("expected len %d, got %d (%v)", len(tt.expected), len(result), result)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("at index %d: expected %q, got %q", i, tt.expected[i], result[i])
				}
			}
		})
	}
}
