package targets

import (
	"fmt"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal"
)

func backupWorkLaptopToExternal(cfg internal.Config) error {
	rCmd := []string{"-av", "-e", "ssh", "worklaptop:/home/brun0/Desktop/work", "worklaptop:/home/brun0/Desktop/shared_folder", "/mnt/pve/external/worklaptop_backup"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "toExternal", cfg.Targets[0].Instance, cfg.Pushgateway.Host); err != nil {
		return err

	}

	return nil
}

// Function that handles both backups
func ExecuteWorkLaptopBackup(cfg internal.Config, el *internal.ElapsedTime) error {
	isAlive, err := internal.IsAlive(cfg.Targets[0].MAC)
	if err != nil {
		return err
	}
	if isAlive {
		start := time.Now()
		if err := backupWorkLaptopToExternal(cfg); err != nil {
			return err
		}
		// Calculate run time
		end := time.Now()
		el.Target = cfg.Targets[0].Name
		el.Elapsed = end.Sub(start).Seconds()
	} else {
		return fmt.Errorf("[ERROR] the target %s is not alive: %w", cfg.Targets[0].Instance, err)
	}
	return nil
}
