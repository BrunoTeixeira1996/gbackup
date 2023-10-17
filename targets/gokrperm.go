package targets

import (
	"github.com/BrunoTeixeira1996/gbackup/internal"
)

// Function that backups /perm partition in gokrazy
// to external hard drive
func backupGokrPermToExternal(cfg internal.Config) error {
	waiwCmd := []string{"-av", "--delete", "-e", "ssh", "rsync://waiw-backup/waiw", "/mnt/pve/external/gokrazy_backup/waiw_backup"}
	if err := internal.ExecCmdToProm("rsync", waiwCmd, "rsync", cfg.Targets[1].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}
	gmahCmd := []string{"-av", "--delete", "-e", "ssh", "rsync://gmah-backup/gmah", "/mnt/pve/external/gokrazy_backup/gmah_backup"}
	if err := internal.ExecCmdToProm("rsync", gmahCmd, "rsync", cfg.Targets[7].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that copies backed up /perm partition to
// HDD present in proxmox instance
func backupGokrPermToHDD(cfg internal.Config) error {
	c := []string{"-r", "/mnt/pve/external/gokrazy_backup/waiw_backup", "/storagepool/backups/gokrazy_backup/"}
	if err := internal.ExecCmdToProm("cp", c, "cmd", cfg.Targets[1].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	g := []string{"-r", "/mnt/pve/external/gokrazy_backup/gmah_backup", "/storagepool/backups/gokrazy_backup/"}
	if err := internal.ExecCmdToProm("cp", g, "cmd", cfg.Targets[7].Instance, cfg.Pushgateway.Host); err != nil {
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
