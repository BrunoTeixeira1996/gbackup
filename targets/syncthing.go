package targets

import "github.com/BrunoTeixeira1996/gbackup/internal"

// Function that backups Syncthing folder
// to external hard drive
func backupSyncthingToExternal(cfg internal.Config) error {
	rCmd := []string{"-av", "-e","ssh", "syncthing:/root/config/Sync", "/mnt/pve/external/syncthing_backup"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "rsync", cfg.Targets[2].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that copies backed up Syncthing folder
// HDD present in proxmox instance
func backupSyncthingToHDD(cfg internal.Config) error {
	c := []string{"-r", "/mnt/pve/external/syncthing_backup/Sync", "/storagepool/backups/syncthing_backup"}
	err := internal.ExecCmdToProm("cp", c, "cmd", cfg.Targets[2].Instance, cfg.Pushgateway.Host)
	if err != nil {
		return err
	}

	return nil
}

// Function that handles both backups
func ExecuteSyncthingBackup(cfg internal.Config) error {
	if err := backupSyncthingToExternal(cfg); err != nil {
		return err
	}

	if err := backupSyncthingToHDD(cfg); err != nil {
		return err
	}

	return nil
}
