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

type targets struct {
	Name    string
	Command string
}

// Function that keeps the last two backups (newest)
func keepLastTwo() error {
	// list directories in /mnt/datastore/backupExternal
	output, err := exec.Command("ssh", "nas1", "ls", "-la", "/mnt/datastore/backupExternal").Output()
	if err != nil {
		log.Printf("[keepLastTwo] error while listing: %s (%s)\n", output, err)
		return err
	}

	// grab only folders using regex
	regexPattern := `\d{4}-\d{2}-\d{2}`
	re := regexp.MustCompile(regexPattern)

	folderNames := re.FindAllString(string(output), -1)

	// verify if there's at least 3 folders
	if len(folderNames) < 3 {
		log.Printf("[keepLastTwo] skipping this because there is only %d folder(s)\n", len(folderNames))
		return nil
	}

	oldestFolder := fmt.Sprintf("/mnt/datastore/backupExternal/%s", folderNames[0])

	output, err = exec.Command("ssh", "nas1", "rm -r", oldestFolder).Output()
	if err != nil {
		log.Printf("[keepLastTwo] error while deleting the oldest folder: %s (%s)\n", output, err)
		return err
	}
	log.Printf("[keepLastTwo] successfully deleted oldest folder %s\n", oldestFolder)

	return nil
}

// Function that backups /external folder to NAS
func ExecuteExternalToNASBackup(cfg config.Config) error {
	// TODO: calculate time
	//start := time.Now()

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
			Command: "-av -e ssh /mnt/external/worklaptop_backup nas1:/mnt/datastore/backupExternal/" + utils.CurrentTime() + "/external",
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

	log.Printf("[backup info] verifying len of backup folders with keepLastTwo mechanism\n")
	if err := keepLastTwo(); err != nil {
		return err
	}
	log.Printf("[backup info] completed the house clean to keep 2 backup folders (newest)\n")

	// Calculate run time
	// end := time.Now()
	// el.Target = cfg.Targets[1].Name
	// el.Elapsed = end.Sub(start).Seconds()

	return nil
}
