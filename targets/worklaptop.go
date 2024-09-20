package targets

import (
	"fmt"
	"log"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

func backupWorkLaptopToExternal(cfg config.Config) error {

	// note that here we dont want to --delete because we need to save old projects
	command := `-av --copy-links -e ssh
	--exclude=personal
	--exclude=tools
	worklaptop:/home/brun0/Desktop
	worklaptop:/home/brun0/.ssh
	/mnt/external/worklaptop_backup`

	log.Printf("[worklaptop backup info] starting rsync command worklatop -> external\n")
	if err := commands.RsyncCommand(command, "toExternal", cfg.Targets[0].Instance, cfg.Pushgateway.Url); err != nil {
		log.Printf("[worklaptop backup error] could not perform RsyncCommand in worklaptop to external: %s\n", err)
		return err
	}
	log.Printf("[worklaptop backup info] completed backup of worklaptop to external\n")

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
		return fmt.Errorf("[worklaptop backup error] the target %s is not alive: %s\n", cfg.Targets[0].Instance, err)
	}
	return nil
}
