package targets

import (
	"log"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

// Function that backups pinute to external hard drive
func ExecutePinuteBackup(cfg config.Config, el *utils.ElapsedTime) error {
	start := time.Now()
	var e error

	command := `-av --delete
    /home/brun0/nut
    /home/brun0/src
    /home/brun0/.ssh
    /home/brun0/.bash_profile
    /home/brun0/.bashrc
    /mnt/external/pinute_backup`

	log.Printf("[pinute backup info] starting rsync command pinute -> external\n")
	if err := commands.RsyncCommand(command, "toExternal", cfg.Targets[2].Instance, cfg.Pushgateway.Url); err != nil {
		log.Printf("[pinute backup error] could not perform RsyncCommand: %s\n", err)
		e = err

	}

	// Calculate run time
	end := time.Now()
	el.Target = cfg.Targets[1].Name
	el.Elapsed = end.Sub(start).Seconds()

	return e
}
