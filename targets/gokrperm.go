package targets

import (
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
	if err := internal.ExecCmdToProm("rsync", gmahCmd, "toExternal", cfg.Targets[7].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that copies backed up /perm partition to
// HDD present in proxmox instance
func backupGokrPermToHDD(cfg internal.Config) error {
	c := []string{"-av", "--delete", "/mnt/pve/external/gokrazy_backup/waiw_backup", "/storagepool/backups/gokrazy_backup/"}
	if err := internal.ExecCmdToProm("rsync", c, "toStoragePool", cfg.Targets[1].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	g := []string{"-av", "--delete", "/mnt/pve/external/gokrazy_backup/gmah_backup", "/storagepool/backups/gokrazy_backup/"}
	if err := internal.ExecCmdToProm("rsync", g, "toStoragePool", cfg.Targets[7].Instance, cfg.Pushgateway.Host); err != nil {
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
