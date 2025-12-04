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
	minFreeBytes := flag.String("minFreeBytes", "", "Minimum absolute free space to maintain (e.g., 10GB, 500MB)")
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
		var minFreeBytesValue uint64
		if *minFreeBytes != "" {
			var err error
			minFreeBytesValue, err = parseBytes(*minFreeBytes)
			if err != nil {
				log.Fatalf("Invalid minFreeBytes value: %v", err)
			}
		}
		runLegacyMode(*partition, *targetDir, *minFreePercent, minFreeBytesValue, *checkInterval, *dryRun, *human)
	}
}

func runLegacyMode(partition, targetDir string, minFreePercent float64, minFreeBytes uint64, checkInterval time.Duration, dryRun, humanReadable bool) {
	if partition == "" || targetDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Starting Partition Vacuum Daemon (Legacy Mode)")
	log.Printf("Monitoring partition: %s", partition)
	log.Printf("Target directory: %s", targetDir)
	log.Printf("Minimum free space: %.2f%%", minFreePercent)
	if minFreeBytes > 0 {
		log.Printf("Minimum free bytes: %s", formatBytes(minFreeBytes))
	}
	log.Printf("Check interval: %v", checkInterval)
	if dryRun {
		log.Printf("DRY RUN MODE ENABLED")
	}

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// Run once immediately
	checkAndClean([]string{targetDir}, minFreePercent, minFreeBytes, dryRun, humanReadable)

	for range ticker.C {
		checkAndClean([]string{targetDir}, minFreePercent, minFreeBytes, dryRun, humanReadable)
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

	// Count active locations
	activeLocations := 0

	for i, loc := range config.Locations {
		// Apply defaults if not set
		minFree := config.Global.MinFreePercent
		if loc.MinFreePercent != nil {
			minFree = *loc.MinFreePercent
		}

		minFreeBytes := config.Global.MinFreeBytes.Bytes
		if loc.MinFreeBytes != nil {
			minFreeBytes = loc.MinFreeBytes.Bytes
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

		if len(loc.TargetDirs) == 0 {
			log.Printf("Location %d has no target directories, skipping", i)
			continue
		}

		if err := SameFilesystem(loc.TargetDirs); err != nil {
			log.Printf("Location %d configuration error: %v", i, err)
			continue
		}

		log.Printf("Starting monitor for directories: %v", loc.TargetDirs)
		activeLocations++

		go func(idx int, l LocationConfig, mf float64, mfb uint64, iv time.Duration, dr, hr bool) {
			ticker := time.NewTicker(iv)
			defer ticker.Stop()

			// Run once immediately
			checkAndClean(l.TargetDirs, mf, mfb, dr, hr)

			for range ticker.C {
				checkAndClean(l.TargetDirs, mf, mfb, dr, hr)
			}
		}(i, loc, minFree, minFreeBytes, interval, dryRun, humanReadable)
	}

	if activeLocations == 0 {
		log.Fatalf("No valid locations to monitor")
	}

	// Wait forever
	<-done
}

func checkAndClean(targetDirs []string, minFreePercent float64, minFreeBytes uint64, dryRun, humanReadable bool) {
	if len(targetDirs) == 0 {
		return
	}

	// Use the first directory to check disk usage (we verified they are on the same FS)
	partition := targetDirs[0]

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

	// Calculate target free bytes from percentage
	targetFreeByPercent := uint64(float64(usage.Total) * (minFreePercent / 100))

	// Use the larger of percentage-based or absolute minimum
	targetFreeBytes := targetFreeByPercent
	if minFreeBytes > targetFreeBytes {
		targetFreeBytes = minFreeBytes
	}

	// Check if we need to clean up
	needsCleanup := false
	if minFreePercent > 0 && freePercent < minFreePercent {
		needsCleanup = true
	}
	if minFreeBytes > 0 && usage.Free < minFreeBytes {
		needsCleanup = true
	}

	if needsCleanup {
		if minFreeBytes > 0 {
			log.Printf("[%s] Free space (%.2f%% / %s) is below minimum (%.2f%% / %s). Initiating cleanup...",
				partition, freePercent, formatBytes(usage.Free), minFreePercent, formatBytes(minFreeBytes))
		} else {
			log.Printf("[%s] Free space (%.2f%%) is below minimum (%.2f%%). Initiating cleanup...", partition, freePercent, minFreePercent)
		}

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
