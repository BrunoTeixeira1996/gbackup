package utils

import (
	"os"
	"strings"
)

// FIXME: add dynamic mountPoint string
func IsExternalMounted() bool {
	data, _ := os.ReadFile("/proc/mounts")

	// check if the mount point exists in the data
	return strings.Contains(string(data), "/mnt/external")
}
