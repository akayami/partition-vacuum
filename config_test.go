package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig_SingleFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	content := `
[global]
check_interval = "5m"
min_free_percent = 20.0
human_readable = true
dry_run = true

[[location]]
target_dirs = ["/tmp"]
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.Global.CheckInterval.Duration != 5*time.Minute {
		t.Errorf("Expected CheckInterval 5m, got %v", config.Global.CheckInterval.Duration)
	}
	if config.Global.MinFreePercent != 20.0 {
		t.Errorf("Expected MinFreePercent 20.0, got %f", config.Global.MinFreePercent)
	}
	if !config.Global.HumanReadable {
		t.Errorf("Expected HumanReadable true")
	}
	if !config.Global.DryRun {
		t.Errorf("Expected DryRun true")
	}
	if len(config.Locations) != 1 {
		t.Fatalf("Expected 1 location, got %d", len(config.Locations))
	}
	if len(config.Locations[0].TargetDirs) != 1 || config.Locations[0].TargetDirs[0] != "/tmp" {
		t.Errorf("Expected target dir /tmp, got %v", config.Locations[0].TargetDirs)
	}
}

func TestLoadConfig_Directory(t *testing.T) {
	tempDir := t.TempDir()

	// File 1: Global settings + 1 location
	file1 := filepath.Join(tempDir, "01_main.toml")
	content1 := `
[global]
check_interval = "10m"

[[location]]
target_dirs = ["/tmp"]
`
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	// File 2: More global settings + another location
	file2 := filepath.Join(tempDir, "02_extra.toml")
	content2 := `
[global]
min_free_percent = 15.0

[[location]]
target_dirs = ["/home/user/cache"]
`
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	config, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify merging
	if config.Global.CheckInterval.Duration != 10*time.Minute {
		t.Errorf("Expected CheckInterval 10m, got %v", config.Global.CheckInterval.Duration)
	}
	if config.Global.MinFreePercent != 15.0 {
		t.Errorf("Expected MinFreePercent 15.0, got %f", config.Global.MinFreePercent)
	}
	if len(config.Locations) != 2 {
		t.Fatalf("Expected 2 locations, got %d", len(config.Locations))
	}

	// Order depends on file globbing, usually alphabetical
	// We expect "/" then "/home" because 01_main < 02_extra
	if len(config.Locations[0].TargetDirs) != 1 || config.Locations[0].TargetDirs[0] != "/tmp" {
		t.Errorf("Expected first target dir /tmp, got %v", config.Locations[0].TargetDirs)
	}
	if len(config.Locations[1].TargetDirs) != 1 || config.Locations[1].TargetDirs[0] != "/home/user/cache" {
		t.Errorf("Expected second target dir /home/user/cache, got %v", config.Locations[1].TargetDirs)
	}
}

func TestLoadConfig_MinFreeBytes(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	content := `
[global]
check_interval = "5m"
min_free_bytes = "10GB"

[[location]]
target_dirs = ["/tmp"]
min_free_bytes = "5GB"
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	expectedGlobalBytes := uint64(10 * 1024 * 1024 * 1024)
	if config.Global.MinFreeBytes.Bytes != expectedGlobalBytes {
		t.Errorf("Expected Global.MinFreeBytes %d, got %d", expectedGlobalBytes, config.Global.MinFreeBytes.Bytes)
	}

	expectedLocationBytes := uint64(5 * 1024 * 1024 * 1024)
	if config.Locations[0].MinFreeBytes == nil {
		t.Fatalf("Expected Location[0].MinFreeBytes to be set")
	}
	if config.Locations[0].MinFreeBytes.Bytes != expectedLocationBytes {
		t.Errorf("Expected Location[0].MinFreeBytes %d, got %d", expectedLocationBytes, config.Locations[0].MinFreeBytes.Bytes)
	}
}
