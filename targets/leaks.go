package targets

import "github.com/BrunoTeixeira1996/gbackup/internal"

/*
   Here I only tar the folder that is already present in the external hard drive
   and then copy that tar to the HDD
   Finaly I delete the archives that are 15 days old from external hard drive
   and from the hdd that is on proxmox
*/

/*
   get current date in format d/mm/yyyy
   tar folder /mnt/pve/external/leaks with the date
   copy the tar to /storagepool/backups/leaks_backup
   delete archives that are 15 days old from external hard drive and hdd
*/

// Function that compresses folder to a tar
// format with the current date
func compressFolder() {
}

// Function that deletes archives older than 15 days
// in external hard drive and in HDD that reside on proxmox
func clean() {

}

// Function that copies tar file to the
// HDD located in proxmox
func backupLeaksToHDD(cfg internal.Config) error {
	return nil
}

func ExecuteLeaksBackup(cfg internal.Config) error {
	if err := backupLeaksToHDD(cfg); err != nil {
		return err
	}
	return nil
}
