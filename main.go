package main

import (
	"flag"
	"log"
	"os"
	"time"
)

func main() {
	partition := flag.String("partition", "", "Path to the partition to monitor (e.g., /)")
	targetDir := flag.String("targetDir", "", "Path to the directory to clean up (e.g., /var/log/app)")
	minFreePercent := flag.Float64("minFreePercent", 10.0, "Minimum percentage of free space to maintain")
	checkInterval := flag.Duration("checkInterval", 1*time.Minute, "How often to check disk usage")
	dryRun := flag.Bool("dryRun", false, "Simulate deletion without actually removing files")

	flag.Parse()

	if *partition == "" || *targetDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Starting Partition Cleaner Daemon")
	log.Printf("Monitoring partition: %s", *partition)
	log.Printf("Target directory: %s", *targetDir)
	log.Printf("Minimum free space: %.2f%%", *minFreePercent)
	log.Printf("Check interval: %v", *checkInterval)
	if *dryRun {
		log.Printf("DRY RUN MODE ENABLED")
	}

	ticker := time.NewTicker(*checkInterval)
	defer ticker.Stop()

	// Run once immediately
	checkAndClean(*partition, *targetDir, *minFreePercent, *dryRun)

	for range ticker.C {
		checkAndClean(*partition, *targetDir, *minFreePercent, *dryRun)
	}
}

func checkAndClean(partition, targetDir string, minFreePercent float64, dryRun bool) {
	usage, err := GetDiskUsage(partition)
	if err != nil {
		log.Printf("Error getting disk usage for %s: %v", partition, err)
		return
	}

	freePercent := (float64(usage.Free) / float64(usage.Total)) * 100
	log.Printf("Disk Usage: Total=%d, Free=%d (%.2f%%), Used=%d", usage.Total, usage.Free, freePercent, usage.Used)

	if freePercent < minFreePercent {
		log.Printf("Free space (%.2f%%) is below minimum (%.2f%%). Initiating cleanup...", freePercent, minFreePercent)

		// Calculate how many bytes we need to free to reach the target
		// Target free bytes = Total * (minFreePercent / 100)
		targetFreeBytes := uint64(float64(usage.Total) * (minFreePercent / 100))

		err := CleanUp(targetDir, targetFreeBytes, usage.Free, dryRun)
		if err != nil {
			log.Printf("Error during cleanup: %v", err)
		} else {
			log.Println("Cleanup completed successfully.")
		}
	} else {
		log.Println("Free space is sufficient.")
	}
}
