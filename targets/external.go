package targets

import (
	"log"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

// Function that backups /external folder to NAS
func ExecuteExternalToNASBackup(cfg config.Config, el *utils.ElapsedTime) error {
	start := time.Now()

	/*
		TODO:
		exclude template
		exclude snippets
		exclude private
		exclude lost+found
		exclude images
		exclude dump
	*/
	command := "-av --delete -e ssh /mnt/external  nas1:/mnt/datastore/backupExternal"

	log.Printf("[backup info] starting rsync command external -> NAS\n")
	if err := commands.RsyncCommand(command, "toNAS", "external hard drive", cfg.Pushgateway.Url); err != nil {
		log.Printf("[backup error] could not perform RsyncCommand in external to NAS: %s\n", err)
		return err
	}

	// Calculate run time
	end := time.Now()
	el.Target = cfg.Targets[1].Name
	el.Elapsed = end.Sub(start).Seconds()

	return nil
}
