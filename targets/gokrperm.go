package targets

import (
	"log"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

type supportedTargets struct {
	Name    string
	Command string
}

// Function that backups /perm partition in gokrazy
// to external hard drive
func backupGokrPermToExternal(cfg config.Config) error {
	var e error

	sT := []supportedTargets{
		{
			Name:    "waiw",
			Command: "-av --delete -e ssh rsync://waiw-backup/waiw /mnt/external/gokrazy_backup/waiw_backup",
		},
		{
			Name:    "gmah",
			Command: "-av --delete -e ssh rsync://gmah-backup/gmah /mnt/external/gokrazy_backup/gmah_backup",
		},
	}

	log.Printf("[backup info] starting rsync command gokrazy /perm -> external\n")
	for _, t := range sT {
		log.Printf("[backup info] performing backup of %s\n", t.Name)
		if err := commands.RsyncCommand(t.Command, "toExternal", cfg.Targets[1].Instance, cfg.Pushgateway.Url); err != nil {
			log.Printf("[backup error] could not perform RsyncCommand in %s: %s\n", t.Name, err)
			e = err
		}
	}
	return e
}

// Function that backups gokrazy perm folder to external hard drive
func ExecuteGokrPermBackup(cfg config.Config, el *utils.ElapsedTime) error {
	start := time.Now()
	if err := backupGokrPermToExternal(cfg); err != nil {
		return err
	}

	// Calculate run time
	end := time.Now()
	el.Target = cfg.Targets[1].Name
	el.Elapsed = end.Sub(start).Seconds()

	return nil
}
