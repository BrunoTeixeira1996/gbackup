package internal

import (
	"math"
	"os"
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

// Returns folder size in Megabytes
func GetFolderSize(folderPath string) (float64, error) {
	var totalSize float64
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			totalSize += float64(info.Size())
		}
		return nil
	})
	if err != nil {
		return 0.0, err
	}

	return roundFloat((totalSize / (1 << 20)), 2), nil
}
