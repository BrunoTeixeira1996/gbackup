package targets

import (
	"log"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

type targets struct {
	Name    string
	Command string
}

/*
TODO:
Implement function that keeps 2 backups and replace the older one with the new one
*/
func keepTwo() error {
	/*
	checks /mnt/datastore/backupExternal for folders and grab the folder names
	validates the older date
	delete that folder after the backup is executed in case of disaster during the backup
	*/

	return nil
}

// Function that backups /external folder to NAS
func ExecuteExternalToNASBackup(cfg config.Config) error {
	// TODO: calculate time
	//start := time.Now()

	// TODO: i dont want to --delete the worklaptop target
	t := []targets{
		{
			Name: "all (minus worklaptop)",
			Command: `-av --delete -e ssh
			--exclude=template
			--exclude=snippets
			--exclude=private
			--exclude=lost+found
			--exclude=images
			--exclude=dump
			--exclude=worklaptop_backup
			/mnt/external nas1:/mnt/datastore/backupExternal/` + utils.CurrentTime() + `/`,
		},
		{
			Name:    "worklaptop",
			Command: "-av -e ssh /mnt/external/worklaptop_backup nas1:/mnt/datastore/backupExternal/" + utils.CurrentTime() +"/external",
		},
	}

	for _, t := range t {
		log.Printf("[backup info] starting rsync command external -> NAS (%s) - %s\n", cfg.NAS.Name, t.Name)
		if err := commands.RsyncCommand(t.Command, "toNAS", "external hard drive", cfg.Pushgateway.Url); err != nil {
			log.Printf("[backup error] could not perform RsyncCommand in external to NAS: %s\n", err)
			return err
		}
	}
	log.Printf("[backup info] completed backup of external to NAS (%s)\n", cfg.NAS.Name)

	// Calculate run time
	// end := time.Now()
	// el.Target = cfg.Targets[1].Name
	// el.Elapsed = end.Sub(start).Seconds()

	return nil
}
