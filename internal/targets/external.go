package targets

import (
	"log"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/nas"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

type External struct {
	Name          string                `toml:"name"`
	ExternalPath  string                `toml:"external_path"`
	RsyncCommands []config.RsyncCommand `toml:"rsync_commands"`
}

// Verify the external hard drive size based on the operation
func (e *External) VerifyExternalSize(operation string, ts *utils.TargetSize) {
	var err error

	switch operation {
	case "before":
		ts.Before, err = utils.GetFolderSize(e.ExternalPath)
		if err != nil {
			log.Printf("[run error] could not get folder size for external (before): %s\n", err)
		}
	case "after":
		ts.After, err = utils.GetFolderSize(e.ExternalPath)
		if err != nil {
			log.Printf("[run error] could not get folder size for external (before): %s\n", err)
		}
	default:
		log.Println("[validateExternal error] unknown operation")
	}
}

// Initializes the external from the config package.
func InitExternal(cfg config.Config) External {
	return External{
		Name:          cfg.External.Name,
		ExternalPath:  cfg.External.ExternalPath,
		RsyncCommands: cfg.External.RsyncCommands,
	}
}

// Wrap function that backups external hard drive to NAS
// starts the backup based on the commands provided on the toml file for the external object
// after the backup is done, it will clean the third folder (if exists) so we keep 2 folders
// to be prepare for a disaster recovery
func ExecuteExternalToNASBackup(external External, cfg config.Config) error {
	var err error

	for _, rsyncCommand := range external.RsyncCommands {
		if err = commands.RsyncCommand(rsyncCommand.Command, "toNAS", rsyncCommand.Name, cfg.Pushgateway.Url); err != nil {
			log.Printf("[external backup error] could not perform RsyncCommand in external to NAS: %s\n", err)
			return err
		}
	}
	log.Printf("[external backup info] completed backup of external to NAS (%s)\n", cfg.NAS.Name)

	log.Printf("[external backup info] verifying len of backup folders with keepLastTwo mechanism\n")
	if err = nas.KeepLastTwo(); err != nil {
		return err
	}
	log.Printf("[external backup info] completed the house clean to keep 2 backup folders (newest)\n")

	return nil
}
