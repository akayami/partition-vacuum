package main

// DiskUsage contains usage statistics for a filesystem
type DiskUsage struct {
	Total uint64
	Free  uint64
	Used  uint64
}
