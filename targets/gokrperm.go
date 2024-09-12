package targets

import (
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal"
)

// Function that backups /perm partition in gokrazy
// to external hard drive
func backupGokrPermToExternal(cfg internal.Config) error {
	waiwCmd := []string{"-av", "--delete", "-e", "ssh", "rsync://waiw-backup/waiw", "/mnt/pve/external/gokrazy_backup/waiw_backup"}
	if err := internal.ExecCmdToProm("rsync", waiwCmd, "toExternal", cfg.Targets[1].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}
	gmahCmd := []string{"-av", "--delete", "-e", "ssh", "rsync://gmah-backup/gmah", "/mnt/pve/external/gokrazy_backup/gmah_backup"}
	if err := internal.ExecCmdToProm("rsync", gmahCmd, "toExternal", cfg.Targets[1].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that handles both backups
func ExecuteGokrPermBackup(cfg internal.Config, el *internal.ElapsedTime) error {
	start := time.Now()
	if err := backupGokrPermToExternal(cfg); err != nil {
		return err
	}

	// Calculate run time
	end := time.Now()
	el.Target = cfg.Targets[1].Name
	el.Elapsed = end.Sub(start).Seconds()

	return nil
}
