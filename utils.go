package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// byteSize is a wrapper around uint64 to support TOML string decoding of byte sizes
type byteSize struct {
	Bytes uint64
}

func (b *byteSize) UnmarshalText(text []byte) error {
	parsed, err := parseBytes(string(text))
	if err != nil {
		return err
	}
	b.Bytes = parsed
	return nil
}

// parseBytes converts a human-readable byte string (e.g., "10GB", "500 MB", "1.5TB") to bytes
func parseBytes(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty byte size string")
	}

	// Match number (with optional decimal) followed by optional unit
	re := regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*(B|KB|MB|GB|TB|PB|K|M|G|T|P)?$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid byte size format: %s", s)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number in byte size: %s", s)
	}

	unit := strings.ToUpper(matches[2])
	var multiplier float64 = 1

	switch unit {
	case "", "B":
		multiplier = 1
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "P", "PB":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	}

	return uint64(value * multiplier), nil
}

// formatBytes converts bytes to a human-readable string (e.g., "1.23 GB")
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
