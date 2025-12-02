package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Config represents the top-level configuration
type Config struct {
	Global    GlobalConfig     `toml:"global"`
	Locations []LocationConfig `toml:"location"`
}

// GlobalConfig contains default settings
type GlobalConfig struct {
	CheckInterval  duration `toml:"check_interval"`
	DryRun         bool     `toml:"dry_run"`
	HumanReadable  bool     `toml:"human_readable"`
	MinFreePercent float64  `toml:"min_free_percent"`
}

// LocationConfig defines a specific partition to monitor and directories to clean
type LocationConfig struct {
	TargetDirs     []string  `toml:"target_dirs"`
	MinFreePercent *float64  `toml:"min_free_percent"` // Optional override
	CheckInterval  *duration `toml:"check_interval"`   // Optional override
	DryRun         *bool     `toml:"dry_run"`          // Optional override
}

// duration is a wrapper around time.Duration to support TOML string decoding
type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// LoadConfig loads configuration from the given path (file or directory)
func LoadConfig(path string) (*Config, error) {
	config := &Config{
		Global: GlobalConfig{
			CheckInterval:  duration{1 * time.Minute},
			MinFreePercent: 10.0,
		},
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config path %s: %w", path, err)
	}

	if info.IsDir() {
		// Load all .toml files in the directory
		files, err := filepath.Glob(filepath.Join(path, "*.toml"))
		if err != nil {
			return nil, fmt.Errorf("failed to glob config dir %s: %w", path, err)
		}

		for _, file := range files {
			log.Printf("Loading config file: %s", file)
			var partialConfig Config
			if _, err := toml.DecodeFile(file, &partialConfig); err != nil {
				return nil, fmt.Errorf("failed to decode config file %s: %w", file, err)
			}
			// Merge
			if partialConfig.Global.CheckInterval.Duration != 0 {
				config.Global.CheckInterval = partialConfig.Global.CheckInterval
			}
			if partialConfig.Global.MinFreePercent != 0 {
				config.Global.MinFreePercent = partialConfig.Global.MinFreePercent
			}
			if partialConfig.Global.DryRun {
				config.Global.DryRun = true
			}
			if partialConfig.Global.HumanReadable {
				config.Global.HumanReadable = true
			}
			config.Locations = append(config.Locations, partialConfig.Locations...)
		}
	} else {
		// Load single file
		if _, err := toml.DecodeFile(path, config); err != nil {
			return nil, fmt.Errorf("failed to decode config file %s: %w", path, err)
		}
	}

	return config, nil
}
