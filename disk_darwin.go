package main

import (
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
	free := uint64(stat.Bavail) * uint64(stat.Bsize)
	total := uint64(stat.Blocks) * uint64(stat.Bsize)
	used := total - free

	return DiskUsage{
		Total: total,
		Free:  free,
		Used:  used,
	}, nil
}
