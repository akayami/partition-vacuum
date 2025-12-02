package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanUp(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cleaner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create dummy files with different modification times
	files := []struct {
		name    string
		size    int64
		modTime time.Time
	}{
		{"file1.txt", 100, time.Now().Add(-3 * time.Hour)},        // Oldest, in root
		{"subdir/file2.txt", 200, time.Now().Add(-2 * time.Hour)}, // Middle, in subdir
		{"file3.txt", 300, time.Now().Add(-1 * time.Hour)},        // Newest, in root
	}

	for _, f := range files {
		path := filepath.Join(tempDir, f.name)
		err := os.WriteFile(path, make([]byte, f.size), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
		// Set modification time
		err = os.Chtimes(path, time.Now(), f.modTime)
		if err != nil {
			t.Fatalf("Failed to set mtime for %s: %v", path, err)
		}
	}

	// Total size: 600.
	// Target free: 150.
	// Need to delete file1 (100) and file2 (200) to reach target free?
	// Wait, logic is: delete until currentFree >= targetFree.
	// Start free: 0. Target: 150.
	// Delete file1 (100) -> free 100.
	// Delete file2 (200) -> free 300. Stop.

	targetFree := uint64(150)
	currentFree := uint64(0)

	err = CleanUp([]string{tempDir}, targetFree, currentFree, false, false)
	if err != nil {
		t.Fatalf("CleanUp failed: %v", err)
	}

	// Verify file1 and file2 are gone, file3 remains
	if _, err := os.Stat(filepath.Join(tempDir, "file1.txt")); !os.IsNotExist(err) {
		t.Errorf("file1.txt should have been deleted")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "subdir/file2.txt")); !os.IsNotExist(err) {
		t.Errorf("subdir/file2.txt should have been deleted")
	}
	// Verify subdir is gone (it should be empty after file2 is deleted)
	if _, err := os.Stat(subDir); !os.IsNotExist(err) {
		t.Errorf("subdir should have been removed")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "file3.txt")); err != nil {
		t.Errorf("file3.txt should still exist")
	}
}

func TestCleanUp_MultipleDirectories(t *testing.T) {
	tempDir := t.TempDir()
	dir1 := filepath.Join(tempDir, "dir1")
	dir2 := filepath.Join(tempDir, "dir2")

	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatal(err)
	}

	// Create files:
	// dir1/old.txt (age 3h, size 100)
	// dir2/mid.txt (age 2h, size 100)
	// dir1/new.txt (age 1h, size 100)

	files := []struct {
		path string
		age  time.Duration
	}{
		{filepath.Join(dir1, "old.txt"), 3 * time.Hour},
		{filepath.Join(dir2, "mid.txt"), 2 * time.Hour},
		{filepath.Join(dir1, "new.txt"), 1 * time.Hour},
	}

	for _, f := range files {
		if err := os.WriteFile(f.path, make([]byte, 100), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(f.path, time.Now(), time.Now().Add(-f.age)); err != nil {
			t.Fatal(err)
		}
	}

	// Total size: 300.
	// Target free: 150 (needs to free 150+ bytes).
	// Start free: 0.
	// Should delete old.txt (100) -> free 100.
	// Should delete mid.txt (100) -> free 200. Stop.
	// new.txt should remain.

	err := CleanUp([]string{dir1, dir2}, 150, 0, false, false)
	if err != nil {
		t.Fatalf("CleanUp failed: %v", err)
	}

	if _, err := os.Stat(files[0].path); !os.IsNotExist(err) {
		t.Errorf("old.txt should be deleted")
	}
	if _, err := os.Stat(files[1].path); !os.IsNotExist(err) {
		t.Errorf("mid.txt should be deleted")
	}
	if _, err := os.Stat(files[2].path); os.IsNotExist(err) {
		t.Errorf("new.txt should exist")
	}
}
