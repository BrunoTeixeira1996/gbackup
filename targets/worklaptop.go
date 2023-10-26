package targets

import (
	"fmt"

	"github.com/BrunoTeixeira1996/gbackup/internal"
)

func backupWorkLaptopToExternal(cfg internal.Config) error {
	rCmd := []string{"-av", "--delete", "-e", "ssh", "worklaptop:/home/brun0/Desktop/work", "worklaptop:/home/brun0/Desktop/shared_folder", "/mnt/pve/external/worklaptop_backup"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "toExternal", cfg.Targets[6].Instance, cfg.Pushgateway.Host); err != nil {
		return err

	}

	return nil
}

func backupWorkLaptopToHDD(cfg internal.Config) error {
	c := []string{"-av", "--delete", "/mnt/pve/external/worklaptop_backup/", "/storagepool/backups/"}
	err := internal.ExecCmdToProm("rsync", c, "toStoragePool", cfg.Targets[6].Instance, cfg.Pushgateway.Host)
	if err != nil {
		return err
	}

	return nil
}

// Function that handles both backups
func ExecuteWorkLaptopBackup(cfg internal.Config) error {
	isAlive, err := internal.IsAlive(cfg.Targets[6].MAC)
	if err != nil {
		return err
	}
	if isAlive {
		if err := backupWorkLaptopToExternal(cfg); err != nil {
			return err
		}

		if err := backupWorkLaptopToHDD(cfg); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("The target %s is not alive: %w", cfg.Targets[6].Instance, err)
	}
	return nil
}
