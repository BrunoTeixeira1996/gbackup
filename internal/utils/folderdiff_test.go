package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFolderSize(t *testing.T) {
	// Setup
	tmpDir := "./testfolder"
	err := os.Mkdir(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Error creating test directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files
	content := []byte("This is a test file.")
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	err = os.WriteFile(file1, content, 0644)
	if err != nil {
		t.Fatalf("Error creating file1: %v", err)
	}
	err = os.WriteFile(file2, content, 0644)
	if err != nil {
		t.Fatalf("Error creating file2: %v", err)
	}

	// Manually sum the actual file sizes like GetFolderSize does
	var totalBytes int64
	for _, path := range []string{file1, file2} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Could not stat file %s: %v", path, err)
		}
		totalBytes += info.Size()
	}
	expectedSize := roundFloat(float64(totalBytes)/(1<<20), 2)

	// Call function under test
	gotSize, err := GetFolderSize(tmpDir)
	if err != nil {
		t.Fatalf("Error getting folder size: %v", err)
	}

	// Compare
	if gotSize != expectedSize {
		t.Errorf("Expected size %.2f MB, but got %.2f MB", expectedSize, gotSize)
	}
}
