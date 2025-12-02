package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var version = "dev"

func main() {
	partition := flag.String("partition", "", "Path to the partition to monitor (e.g., /)")
	targetDir := flag.String("targetDir", "", "Path to the directory to clean up (e.g., /var/log/app)")
	minFreePercent := flag.Float64("minFreePercent", 10.0, "Minimum percentage of free space to maintain")
	checkInterval := flag.Duration("checkInterval", 1*time.Minute, "How often to check disk usage")
	dryRun := flag.Bool("dryRun", false, "Simulate deletion without actually removing files")

	human := flag.Bool("h", false, "Show output in human-readable format")
	v := flag.Bool("v", false, "Print version and exit")

	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	if *v {
		fmt.Printf("Partition Vacuum version %s\n", version)
		os.Exit(0)
	}

	// Check if we should run in config mode
	useConfig := *configPath != ""
	if !useConfig {
		// Check default config locations
		// 1. ~/.config/partition-vacuum/
		homeDir, err := os.UserHomeDir()
		if err == nil {
			userConfig := filepath.Join(homeDir, ".config", "partition-vacuum")
			if _, err := os.Stat(userConfig); err == nil {
				useConfig = true
				*configPath = userConfig
			}
		}

		// 2. /etc/partition-vacuum/
		if !useConfig {
			if _, err := os.Stat("/etc/partition-vacuum"); err == nil {
				useConfig = true
				*configPath = "/etc/partition-vacuum"
			}
		}
	}

	if useConfig {
		runConfigMode(*configPath)
	} else {
		runLegacyMode(*partition, *targetDir, *minFreePercent, *checkInterval, *dryRun, *human)
	}
}

func runLegacyMode(partition, targetDir string, minFreePercent float64, checkInterval time.Duration, dryRun, humanReadable bool) {
	if partition == "" || targetDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Starting Partition Vacuum Daemon (Legacy Mode)")
	log.Printf("Monitoring partition: %s", partition)
	log.Printf("Target directory: %s", targetDir)
	log.Printf("Minimum free space: %.2f%%", minFreePercent)
	log.Printf("Check interval: %v", checkInterval)
	if dryRun {
		log.Printf("DRY RUN MODE ENABLED")
	}

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// Run once immediately
	checkAndClean(partition, []string{targetDir}, minFreePercent, dryRun, humanReadable)

	for range ticker.C {
		checkAndClean(partition, []string{targetDir}, minFreePercent, dryRun, humanReadable)
	}
}

func runConfigMode(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	log.Printf("Using configuration source: %s", absPath)

	config, err := LoadConfig(path)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting Partition Vacuum Daemon (Config Mode)")
	if len(config.Locations) == 0 {
		log.Fatalf("No locations defined in configuration")
	}

	// Channel to keep main goroutine alive
	done := make(chan bool)

	for i, loc := range config.Locations {
		// Apply defaults if not set
		minFree := config.Global.MinFreePercent
		if loc.MinFreePercent != nil {
			minFree = *loc.MinFreePercent
		}

		interval := config.Global.CheckInterval.Duration
		if loc.CheckInterval != nil {
			interval = loc.CheckInterval.Duration
		}

		dryRun := config.Global.DryRun
		if loc.DryRun != nil {
			dryRun = *loc.DryRun
		}

		humanReadable := config.Global.HumanReadable

		log.Printf("Starting monitor for partition %s (Directories: %v)", loc.Partition, loc.TargetDirs)

		go func(idx int, l LocationConfig, mf float64, iv time.Duration, dr, hr bool) {
			ticker := time.NewTicker(iv)
			defer ticker.Stop()

			// Run once immediately
			checkAndClean(l.Partition, l.TargetDirs, mf, dr, hr)

			for range ticker.C {
				checkAndClean(l.Partition, l.TargetDirs, mf, dr, hr)
			}
		}(i, loc, minFree, interval, dryRun, humanReadable)
	}

	// Wait forever
	<-done
}

func checkAndClean(partition string, targetDirs []string, minFreePercent float64, dryRun, humanReadable bool) {
	usage, err := GetDiskUsage(partition)
	if err != nil {
		log.Printf("Error getting disk usage for %s: %v", partition, err)
		return
	}

	freePercent := (float64(usage.Free) / float64(usage.Total)) * 100

	if humanReadable {
		log.Printf("[%s] Disk Usage: Total=%s, Free=%s (%.2f%%), Used=%s",
			partition, formatBytes(usage.Total), formatBytes(usage.Free), freePercent, formatBytes(usage.Used))
	} else {
		log.Printf("[%s] Disk Usage: Total=%d, Free=%d (%.2f%%), Used=%d", partition, usage.Total, usage.Free, freePercent, usage.Used)
	}

	if freePercent < minFreePercent {
		log.Printf("[%s] Free space (%.2f%%) is below minimum (%.2f%%). Initiating cleanup...", partition, freePercent, minFreePercent)

		// Calculate how many bytes we need to free to reach the target
		// Target free bytes = Total * (minFreePercent / 100)
		targetFreeBytes := uint64(float64(usage.Total) * (minFreePercent / 100))

		err := CleanUp(targetDirs, targetFreeBytes, usage.Free, dryRun, humanReadable)
		if err != nil {
			log.Printf("[%s] Error during cleanup: %v", partition, err)
		} else {
			log.Printf("[%s] Cleanup completed successfully.", partition)
		}
	} else {
		log.Printf("[%s] Free space is sufficient.", partition)
	}
}
