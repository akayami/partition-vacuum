package main

import "testing"

func TestParseBytes(t *testing.T) {
	tests := []struct {
		input       string
		expected    uint64
		shouldError bool
	}{
		{"1024", 1024, false},
		{"1024B", 1024, false},
		{"1KB", 1024, false},
		{"1 KB", 1024, false},
		{"1K", 1024, false},
		{"1MB", 1024 * 1024, false},
		{"1M", 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"1G", 1024 * 1024 * 1024, false},
		{"1TB", 1024 * 1024 * 1024 * 1024, false},
		{"1T", 1024 * 1024 * 1024 * 1024, false},
		{"1.5GB", uint64(1.5 * 1024 * 1024 * 1024), false},
		{"500MB", 500 * 1024 * 1024, false},
		{"10GB", 10 * 1024 * 1024 * 1024, false},
		{"  10GB  ", 10 * 1024 * 1024 * 1024, false},
		{"", 0, true},
		{"invalid", 0, true},
		{"GB", 0, true},
	}

	for _, test := range tests {
		result, err := parseBytes(test.input)
		if test.shouldError {
			if err == nil {
				t.Errorf("parseBytes(%q) should have returned an error", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("parseBytes(%q) returned unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("parseBytes(%q) = %d, expected %d", test.input, result, test.expected)
			}
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1023, "1023 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1024 * 1024 * 1024 * 1024, "1.00 TB"},
		{1024 * 1024 * 1024 * 1024 * 1024, "1.00 PB"},
	}

	for _, test := range tests {
		result := formatBytes(test.input)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s; want %s", test.input, result, test.expected)
		}
	}
}
