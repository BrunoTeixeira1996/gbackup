package targets

import (
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal"
)

// Function that backups Gocam folder
// to external hard drive
func backupGocamToExternal(cfg internal.Config) error {
	rCmd := []string{"-av", "-e", "ssh", "gocam:/root/gocam/", "/mnt/pve/external/gocam_backup"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "toExternal", cfg.Targets[8].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that copies backed up Gocam folder
// HDD present in proxmox instance
func backupGocamToHDD(cfg internal.Config) error {
	c := []string{"-av", "/mnt/pve/external/gocam_backup/", "/storagepool/backups/gocam_backup"}
	err := internal.ExecCmdToProm("rsync", c, "toStoragePool", cfg.Targets[8].Instance, cfg.Pushgateway.Host)
	if err != nil {
		return err
	}

	return nil
}

// Function that handles both backups
func ExecuteGocamBackup(cfg internal.Config, el *internal.ElapsedTime) error {
	start := time.Now()

	if err := backupGocamToExternal(cfg); err != nil {
		return err
	}

	if err := backupGocamToHDD(cfg); err != nil {
		return err
	}

	// Calculate run time
	end := time.Now()
	el.Target = cfg.Targets[8].Name
	el.Elapsed = end.Sub(start).Seconds()

	return nil
}
