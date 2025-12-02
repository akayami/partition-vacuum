package main

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

// GetDiskUsage returns the disk usage for the filesystem containing the given path.
func GetDiskUsage(path string) (DiskUsage, error) {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")

	var freeBytesAvailableToCaller, totalNumberOfBytes, totalNumberOfFreeBytes int64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return DiskUsage{}, err
	}

	r1, _, err := c.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailableToCaller)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	if r1 == 0 {
		return DiskUsage{}, err
	}

	return DiskUsage{
		Total: uint64(totalNumberOfBytes),
		Free:  uint64(freeBytesAvailableToCaller), // Use available to caller, not total free
		Used:  uint64(totalNumberOfBytes - freeBytesAvailableToCaller),
	}, nil
}

// SameFilesystem checks if all paths are on the same filesystem (volume).
func SameFilesystem(paths []string) error {
	if len(paths) < 2 {
		return nil
	}

	var firstVol string
	for i, path := range paths {
		vol := filepath.VolumeName(path)
		if vol == "" {
			// If VolumeName returns empty (e.g. relative path), we might need absolute path
			abs, err := filepath.Abs(path)
			if err == nil {
				vol = filepath.VolumeName(abs)
			}
		}

		if i == 0 {
			firstVol = vol
		} else {
			if vol != firstVol {
				return fmt.Errorf("paths %s and %s are on different volumes", paths[0], path)
			}
		}
	}
	return nil
}
