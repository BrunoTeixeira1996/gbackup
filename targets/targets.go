package targets

import "github.com/BrunoTeixeira1996/gbackup/internal/utils"

type BackupResult struct {
	TargetName string
	Elapsed    utils.ElapsedTime
	TargetSize utils.TargetSize
	Err        error
}
