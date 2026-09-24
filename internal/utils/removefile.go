package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// KeepLastN deletes all files matching pattern except the newest n.
// Relies on the YYYY-MM-DD in the file name so that alphabetical order == chronological order.
func KeepLastN(pattern string, n int) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("list backup files: %w", err)
	}
	if len(files) <= n {
		return nil
	}

	sort.Strings(files)
	for _, f := range files[:len(files)-n] {
		if err := os.Remove(f); err != nil {
			return fmt.Errorf("remove old backup %s: %w", f, err)
		}
		log.Printf("[network info] Removed old backup: %s", f)
	}

	return nil
}
