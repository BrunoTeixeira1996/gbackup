package targets

import (
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

type TargetWrapper struct {
	config.Target
}

type BackupResult struct {
	TargetName string
	Elapsed    utils.ElapsedTime
	TargetSize utils.TargetSize
	Err        error
}
