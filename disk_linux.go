package main

import (
	"fmt"
	"syscall"
)

// GetDiskUsage returns the disk usage for the filesystem containing the given path.
func GetDiskUsage(path string) (DiskUsage, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return DiskUsage{}, err
	}

	// Available blocks * size per block = available space in bytes
	// Bavail is free blocks available to unprivileged user
	free := stat.Bavail * uint64(stat.Bsize)
	total := stat.Blocks * uint64(stat.Bsize)
	used := total - free

	return DiskUsage{
		Total: total,
		Free:  free,
		Used:  used,
	}, nil
}

// SameFilesystem checks if all paths are on the same filesystem (device).
func SameFilesystem(paths []string) error {
	if len(paths) < 2 {
		return nil
	}

	var firstDev uint64
	for i, path := range paths {
		var stat syscall.Stat_t
		if err := syscall.Stat(path, &stat); err != nil {
			return fmt.Errorf("failed to stat %s: %w", path, err)
		}

		if i == 0 {
			firstDev = stat.Dev
		} else {
			if stat.Dev != firstDev {
				return fmt.Errorf("paths %s and %s are on different filesystems", paths[0], path)
			}
		}
	}
	return nil
}
