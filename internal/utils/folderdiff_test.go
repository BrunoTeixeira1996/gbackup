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

	// Create files (big enough to round to a nonzero MB value)
	content := make([]byte, 50*1024)
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

func TestGetFolderSize_EmptyFolder(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "empty")
	if err != nil {
		t.Fatalf("Error creating test directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gotSize, err := GetFolderSize(tmpDir)
	if err != nil {
		t.Fatalf("Error getting folder size: %v", err)
	}
	if gotSize != 0 {
		t.Errorf("Expected size 0 for empty folder, got %.2f MB", gotSize)
	}
}

func TestGetFolderSize_NestedSubdirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nested")
	if err != nil {
		t.Fatalf("Error creating test directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Error creating subdirectory: %v", err)
	}

	// needs to be big enough to round to a nonzero MB value
	content := make([]byte, 50*1024)
	if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), content, 0644); err != nil {
		t.Fatalf("Error creating nested file: %v", err)
	}

	expectedSize := roundFloat(float64(len(content))/(1<<20), 2)

	gotSize, err := GetFolderSize(tmpDir)
	if err != nil {
		t.Fatalf("Error getting folder size: %v", err)
	}
	if gotSize != expectedSize {
		t.Errorf("Expected size %.2f MB, but got %.2f MB", expectedSize, gotSize)
	}
}

func TestGetFolderSize_UnreadableSubdirectoryIsSkippedNotFatal(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root, permission checks don't apply")
	}

	tmpDir, err := os.MkdirTemp("", "unreadable")
	if err != nil {
		t.Fatalf("Error creating test directory: %v", err)
	}
	defer func() {
		os.Chmod(filepath.Join(tmpDir, "locked"), 0755) // restore so RemoveAll can clean up
		os.RemoveAll(tmpDir)
	}()

	content := make([]byte, 50*1024)
	if err := os.WriteFile(filepath.Join(tmpDir, "readable.txt"), content, 0644); err != nil {
		t.Fatalf("Error creating readable file: %v", err)
	}

	lockedDir := filepath.Join(tmpDir, "locked")
	if err := os.Mkdir(lockedDir, 0755); err != nil {
		t.Fatalf("Error creating locked directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lockedDir, "hidden.txt"), content, 0644); err != nil {
		t.Fatalf("Error creating file in locked directory: %v", err)
	}
	if err := os.Chmod(lockedDir, 0); err != nil {
		t.Fatalf("Error removing permissions from locked directory: %v", err)
	}

	// the locked dir's contents can't be walked, but that shouldn't be
	// fatal - it just gets skipped, and the readable file still counts
	expectedSize := roundFloat(float64(len(content))/(1<<20), 2)

	gotSize, err := GetFolderSize(tmpDir)
	if err != nil {
		t.Fatalf("GetFolderSize() returned unexpected error: %v", err)
	}
	if gotSize != expectedSize {
		t.Errorf("Expected size %.2f MB (only the readable file), but got %.2f MB", expectedSize, gotSize)
	}
}

func TestGetFolderSize_NonexistentFolder(t *testing.T) {
	_, err := GetFolderSize("/path/does/not/exist/hopefully")
	if err == nil {
		t.Error("Expected an error for a nonexistent folder, got nil")
	}
}

func TestRoundFloat(t *testing.T) {
	tests := []struct {
		name      string
		val       float64
		precision uint
		want      float64
	}{
		{"rounds down", 1.2340, 2, 1.23},
		{"rounds up", 1.2360, 2, 1.24},
		{"zero precision", 1.6, 0, 2},
		{"already rounded", 5.5, 1, 5.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roundFloat(tt.val, tt.precision); got != tt.want {
				t.Errorf("roundFloat(%v, %d) = %v, want %v", tt.val, tt.precision, got, tt.want)
			}
		})
	}
}
