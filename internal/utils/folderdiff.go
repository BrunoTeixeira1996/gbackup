package utils

import (
	"fmt"
	"io/fs"
	"log"
	"math"
	"path/filepath"
)

type TargetSize struct {
	Name   string
	Before float64
	After  float64
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// GetFolderSize returns the total folder size in megabytes (MB)
func GetFolderSize(folderPath string) (float64, error) {
	var totalSize float64

	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Gracefully handle permission or read errors
			log.Printf("[folderdiff warning] skipping %s: %v\n", path, err)
			return nil
		}

		if d == nil || d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			// Skip files we can't stat
			log.Printf("[folderdiff warning] could not read info for %s: %v\n", path, err)
			return nil
		}

		totalSize += float64(info.Size())
		return nil
	})

	if err != nil {
		return 0.0, fmt.Errorf("[folderdiff error] could not perform filepath.WalkDir: %w", err)
	}

	final := roundFloat((totalSize / (1 << 20)), 2)
	return final, nil
}
