package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// FileInfo holds minimal info needed for sorting and deletion
type FileInfo struct {
	Path string
	Size int64
	Age  int64 // Unix timestamp of modification time
}

// CleanUp deletes oldest files in dir until currentFreeBytes >= targetFreeBytes
// It also removes any directories that become empty.
func CleanUp(dir string, targetFreeBytes uint64, currentFreeBytes uint64, dryRun bool) error {
	// 1. Collect all files
	var files []FileInfo

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil // Skip files we can't stat
		}

		files = append(files, FileInfo{
			Path: path,
			Size: info.Size(),
			Age:  info.ModTime().Unix(),
		})
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	// 2. Sort by Age ascending (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Age < files[j].Age
	})

	// 3. Delete files until target reached
	bytesNeeded := targetFreeBytes - currentFreeBytes
	var bytesDeleted uint64 = 0

	// Only delete if we actually need space
	if currentFreeBytes < targetFreeBytes {
		for _, file := range files {
			if bytesDeleted >= bytesNeeded {
				break
			}

			if dryRun {
				fmt.Printf("[DRY RUN] Would delete %s (size: %d)\n", file.Path, file.Size)
			} else {
				err := os.Remove(file.Path)
				if err != nil {
					fmt.Printf("Failed to delete %s: %v\n", file.Path, err)
					continue
				}
				fmt.Printf("Deleted %s (size: %d)\n", file.Path, file.Size)
			}
			bytesDeleted += uint64(file.Size)
		}
	}

	// 4. Remove empty directories
	if err := removeEmptyDirs(dir, dryRun); err != nil {
		fmt.Printf("Error removing empty directories: %v\n", err)
	}

	if currentFreeBytes < targetFreeBytes && bytesDeleted < bytesNeeded {
		return fmt.Errorf("deleted all eligible files but still need %d bytes", bytesNeeded-bytesDeleted)
	}

	return nil
}

func removeEmptyDirs(root string, dryRun bool) error {
	var dirs []string

	// Collect all directories
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Ignore errors accessing paths
		}
		if d.IsDir() && path != root {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Sort directories by length descending (deepest first)
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	// Try to remove each directory
	for _, d := range dirs {
		if dryRun {
			// In dry run, we can't easily know if a directory WOULD be empty because we didn't actually delete files.
			// However, we can check if it IS empty now. But that might be misleading if it contains files we WOULD have deleted.
			// A simple approximation is to just say we would check/remove it.
			// Or, strictly speaking, we should only say we'd remove it if it's currently empty.
			// Let's stick to checking if it's empty now.
			isEmpty, _ := isDirEmpty(d)
			if isEmpty {
				fmt.Printf("[DRY RUN] Would remove empty directory: %s\n", d)
			}
		} else {
			// os.Remove fails if directory is not empty, which is exactly what we want
			err := os.Remove(d)
			if err == nil {
				fmt.Printf("Removed empty directory: %s\n", d)
			}
		}
	}

	return nil
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == nil {
		return false, nil // Found at least one entry
	}
	return true, nil // EOF or error (assume empty or inaccessible)
}
