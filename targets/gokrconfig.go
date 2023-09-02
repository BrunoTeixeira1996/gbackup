package targets

import (
	"github.com/BrunoTeixeira1996/gbackup/internal"
)

/*
   Here I only copy from external hard drive to HDD because my laptop
   is running the rsync command to copy the gokrazy folder to the proxmox
   instance
   I rather run a cronjob in my laptop than open a ssh connection
*/

// Function that copies the backed up gokrazy file
// that holds all useful information about brun0-pi instance
// to HDD present in proxmox
func backupGokrConfToHDD(cfg internal.Config) error {
	c := []string{"-r", "/mnt/pve/external/gokrazy_backup/brun0-pi", "/storagepool/backups/gokrazy_backup"}
	err := internal.ExecCmdToProm("cp", c, "cmd", cfg.Targets[5].Instance, cfg.Pushgateway.Host)
	if err != nil {
		return err
	}

	return nil
}

// Function that handles both backups
func ExecuteGokrConfBackup(cfg internal.Config) error {
	if err := backupGokrConfToHDD(cfg); err != nil {
		return err
	}

	return nil
}
