package targets

import (
	"github.com/BrunoTeixeira1996/gbackup/internal"
)

// Function that backups /perm partition in gokrazy
// to external hard drive
func backupGokrPermToExternal(cfg internal.Config) error {
	rCmd := []string{"-av", "-e", "ssh", "rsync://waiw-backup/waiw", "/mnt/pve/external/gokrazy_backup/waiw_backup"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "rsync", cfg.Targets[1].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that copies backed up /perm partition to
// HDD present in proxmox instance
func backupGokrPermToHDD(cfg internal.Config) error {
	c := []string{"-r", "/mnt/pve/external/gokrazy_backup/waiw_backup", "/storagepool/backups/gokrazy_backup/"}
	err := internal.ExecCmdToProm("cp", c, "cmd", cfg.Targets[1].Instance, cfg.Pushgateway.Host)
	if err != nil {
		return err
	}

	return nil
}

// Function that handles both backups
func ExecuteGokrPermBackup(cfg internal.Config) error {
	if err := backupGokrPermToExternal(cfg); err != nil {
		return err
	}

	if err := backupGokrPermToHDD(cfg); err != nil {
		return err
	}

	return nil
}
