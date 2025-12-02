package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanUpDryRun(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cleaner_dryrun_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file
	filePath := filepath.Join(tempDir, "file1.txt")
	err = os.WriteFile(filePath, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Run CleanUp with dryRun=true
	// Target free space high enough to trigger deletion
	targetFree := uint64(1000)
	currentFree := uint64(0)

	err = CleanUp([]string{tempDir}, targetFree, currentFree, true, false)
	if err != nil && err.Error()[:28] != "deleted all eligible files b" {
		t.Fatalf("CleanUp failed with unexpected error: %v", err)
	}

	// Verify file still exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("file1.txt should NOT have been deleted in dry run")
	}
}
