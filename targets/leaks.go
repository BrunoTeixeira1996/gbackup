package targets

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal"
)

/*
   Here I only tar the folder that is already present in the external hard drive
   and then copy that tar to the HDD
   Finaly I delete the archives that are 15 days old from external hard drive
   and from the hdd that is on proxmox
*/

// Returns number of days from the current time
// to the modificationTime
func getDateDiff(modificationTime string) int {
	timeFormat := "2006-01-02"
	t, _ := time.Parse(timeFormat, modificationTime)
	duration := time.Now().Sub(t)

	return int(duration.Hours() / 24)
}

// Function that compresses folder to a tar
// format with the current date (yy/mm/dd)
// and returns the current location
func compressFolder() (string, error) {
	timeNow := time.Now()
	timeNowCorrectFormat := timeNow.Format("2006-01-02")

	tarN := "/mnt/pve/external/leaks_backup/leak-" + timeNowCorrectFormat + ".tar"

	cmd := exec.Command("tar", "-cvf", tarN, "/mnt/pve/external/leaks")

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return tarN, nil
}

// Function that deletes archives older than 15 days
// in external hard drive and in HDD that reside on proxmox
func clean(directoryToClean string) error {
	files, err := os.ReadDir(directoryToClean)
	if err != nil {
		return err
	}

	for _, file := range files {

		p := directoryToClean + file.Name()

		// gather info from file
		fileInfo, err := os.Stat(p)
		if err != nil {
			return err
		}

		modificationTime := strings.Split(fileInfo.ModTime().String(), " ")[0]

		daysDiff := getDateDiff(modificationTime)

		// delete all files that are older than 15 days
		if daysDiff >= 14 {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}

	return nil
}

// Function that copies tar file to the
// HDD located in proxmox
func backupLeaksToHDD(cfg internal.Config) error {
	var (
		externalLocation string
		err              error
	)

	// tar in external hard drive
	if externalLocation, err = compressFolder(); err != nil {
		return err
	}

	rCmd := []string{"-av", externalLocation, "/storagepool/backups/leaks_backup/"}
	if err = internal.ExecCmdToProm("rsync", rCmd, "rsync", cfg.Targets[3].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

func ExecuteLeaksBackup(cfg internal.Config) error {
	if err := backupLeaksToHDD(cfg); err != nil {
		return err
	}

	// cleaning files older than 15 days
	dirsToBeCleaned := []string{"/storagepool/backups/leaks_backup/", "/mnt/pve/external/leaks_backup/"}
	for _, d := range dirsToBeCleaned {
		if err := clean(d); err != nil {
			return err
		}
	}
	return nil
}
