package internal

import (
	"log"
	"os"
)

// For now this works for safe printing in goroutines
var Logger = log.New(os.Stdout, "", 0)
