package targets

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

type target struct {
	Name    string
	Command string
}

// Function that keeps the last two backups (newest)
func keepLastTwo() error {
	// list directories in /mnt/datastore/backupExternal
	output, err := exec.Command("ssh", "nas1", "ls", "-la", "/mnt/datastore/backupExternal").Output()
	if err != nil {
		log.Printf("[external backup error] error while listing: %s (%s)\n", output, err)
		return err
	}

	// grab only folders using regex
	regexPattern := `\d{4}-\d{2}-\d{2}`
	re := regexp.MustCompile(regexPattern)

	folderNames := re.FindAllString(string(output), -1)

	// verify if there's at least 3 folders
	if len(folderNames) < 3 {
		log.Printf("[external backup info] skipping this because there is only %d folder(s)\n", len(folderNames))
		return nil
	}

	oldestFolder := fmt.Sprintf("/mnt/datastore/backupExternal/%s", folderNames[0])

	output, err = exec.Command("ssh", "nas1", "sudo", "rm", "-r", oldestFolder).Output()
	if err != nil {
		log.Printf("[external backup error] error while deleting the oldest folder (%s): %s (%s)\n", oldestFolder, output, err)
		return err
	}
	log.Printf("[external backup info] successfully deleted oldest folder %s\n", oldestFolder)

	return nil
}

// Function that backups /external folder to NAS
func ExecuteExternalToNASBackup(cfg config.Config) error {
	t := []target{
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
	}

	// I need to give different name to "external hard drive" so I can grab both rsync commands on prometheus
	for _, t := range t {
		instance := fmt.Sprintf("external-hard-drive-%s", t.Name)
		log.Printf("[external backup info] starting rsync command external -> NAS (%s) - %s\n", cfg.NAS.Name, t.Name)
		if err := commands.RsyncCommand(t.Command, "toNAS", instance, cfg.Pushgateway.Url); err != nil {
			log.Printf("[external backup error] could not perform RsyncCommand in external to NAS: %s\n", err)
			return err
		}
	}
	log.Printf("[external backup info] completed backup of external to NAS (%s)\n", cfg.NAS.Name)

	log.Printf("[external backup info] verifying len of backup folders with keepLastTwo mechanism\n")
	if err := keepLastTwo(); err != nil {
		return err
	}
	log.Printf("[external backup info] completed the house clean to keep 2 backup folders (newest)\n")

	return nil
}
