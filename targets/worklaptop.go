package targets

import (
	"fmt"
	"log"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

/*
TODO: 
backup ~/Desktop ~/.ssh /etc
execute pipx list and save that for future (pipx list and save output)
execute apt list installed and save that for future (dpkg --get-selections and save output)
*/
func backupWorkLaptopToExternal(cfg config.Config) error {
	command := "-av -e ssh worklaptop:/home/brun0/Desktop/work worklaptop:/home/brun0/Desktop/shared_folder /mnt/pve/external/worklaptop_backup"

	log.Printf("[backup info] starting rsync command worklatop -> external\n")
	if err := commands.RsyncCommand(command, "toExternal", "worklaptop", cfg.Pushgateway.Url); err != nil {
		log.Printf("[backup error] could not perform RsyncCommand in worklaptop to external: %s\n", err)
		return err
	}
	log.Printf("[backup info] completed backup of worklaptop to external\n")

	return nil
}

// Function that handles both backups
func ExecuteWorkLaptopBackup(cfg config.Config, el *utils.ElapsedTime) error {
	isAlive, err := utils.IsAlive(cfg.Targets[0].MAC)
	if err != nil {
		return err
	}
	if isAlive {
		start := time.Now()
		if err := backupWorkLaptopToExternal(cfg); err != nil {
			return err
		}
		// Calculate run time
		end := time.Now()
		el.Target = cfg.Targets[0].Name
		el.Elapsed = end.Sub(start).Seconds()
	} else {
		return fmt.Errorf("[backup error] the target %s is not alive: %w", cfg.Targets[0].Instance, err)
	}
	return nil
}
