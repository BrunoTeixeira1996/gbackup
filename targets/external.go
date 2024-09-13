package targets

import (
	"log"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
)

// Function that backups /external folder to NAS
func ExecuteExternalToNASBackup(cfg config.Config) error {
	// TODO: calculate time
	//start := time.Now()

	command := `-av --delete -e ssh
	--exclude=template
	--exclude=snippets
	--exclude=private
	--exclude=lost+found
	--exclude=images
	--exclude=dump
	/mnt/external nas1:/mnt/datastore/backupExternal`

	log.Printf("[backup info] starting rsync command external -> NAS (%s)\n", cfg.NAS.Name)
	if err := commands.RsyncCommand(command, "toNAS", "external hard drive", cfg.Pushgateway.Url); err != nil {
		log.Printf("[backup error] could not perform RsyncCommand in external to NAS: %s\n", err)
		return err
	}
	log.Printf("[backup info] completed backup of external to NAS (%s)\n", cfg.NAS.Name)

	// Calculate run time
	// end := time.Now()
	// el.Target = cfg.Targets[1].Name
	// el.Elapsed = end.Sub(start).Seconds()

	return nil
}
