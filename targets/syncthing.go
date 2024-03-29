package targets

import (
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal"
)

// Function that backups Syncthing folder
// to external hard drive
func backupSyncthingToExternal(cfg internal.Config) error {
	rCmd := []string{"-av", "--delete", "-e", "ssh", "syncthing:/root/Sync", "/mnt/pve/external/syncthing_backup"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "toExternal", cfg.Targets[2].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that copies backed up Syncthing folder
// HDD present in proxmox instance
func backupSyncthingToHDD(cfg internal.Config) error {
	c := []string{"-av", "--delete", "/mnt/pve/external/syncthing_backup/Sync", "/storagepool/backups/syncthing_backup"}
	err := internal.ExecCmdToProm("rsync", c, "toStoragePool", cfg.Targets[2].Instance, cfg.Pushgateway.Host)
	if err != nil {
		return err
	}

	return nil
}

// Function that handles both backups
func ExecuteSyncthingBackup(cfg internal.Config, el *internal.ElapsedTime) error {
	start := time.Now()

	if err := backupSyncthingToExternal(cfg); err != nil {
		return err
	}

	if err := backupSyncthingToHDD(cfg); err != nil {
		return err
	}

	// Calculate run time
	end := time.Now()
	el.Target = cfg.Targets[2].Name
	el.Elapsed = end.Sub(start).Seconds()

	return nil
}
