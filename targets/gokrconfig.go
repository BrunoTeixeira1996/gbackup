package targets

import (
	"github.com/BrunoTeixeira1996/gbackup/internal"
)

// Function that backups gokrazy config
// to external hard drive
func backupGokrConfToExternal(cfg internal.Config) error {
	rCmd := []string{"-av", "--delete", "-e", "ssh", "gkconfig:/root/gokrazy/brun0-pi", "/mnt/pve/external/gokrazy_backup/"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "toExternal", cfg.Targets[5].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that copies the backed up gokrazy file
// that holds all useful information about brun0-pi instance
// to HDD present in proxmox
func backupGokrConfToHDD(cfg internal.Config) error {
	c := []string{"-av", "--delete", "/mnt/pve/external/gokrazy_backup/brun0-pi", "/storagepool/backups/gokrazy_backup"}
	err := internal.ExecCmdToProm("rsync", c, "toStoragePool", cfg.Targets[5].Instance, cfg.Pushgateway.Host)
	if err != nil {
		return err
	}

	return nil
}

// Function that handles both backups
func ExecuteGokrConfBackup(cfg internal.Config) error {
	if err := backupGokrConfToExternal(cfg); err != nil {
		return err
	}

	if err := backupGokrConfToHDD(cfg); err != nil {
		return err
	}

	return nil
}
