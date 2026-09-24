package utils

import (
	"fmt"
	"io"
	"os"
)

func DownloadFile(fileName string, body io.ReadCloser) error {
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("create backup file: %w", err)
	}

	defer f.Close()

	if _, err := io.Copy(f, body); err != nil {
		return fmt.Errorf("write backup file: %w", err)
	}

	return nil
}
