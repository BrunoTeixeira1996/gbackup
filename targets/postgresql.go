package targets

import (
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal"
)

// Function that dumps databases in postgresql server
// and then backups up to external hard drive
func backupPostgresqlToExternal(cfg internal.Config) error {
	//cmd := "pg_dump waiw > waiw.sql && pg_dump leaks > leaks.sql"
	// note that in order to pg_dump without password you need to export PGPASSWORD="yourpass"
	cmd := "pg_dump waiw > waiw.sql"

	err := internal.ExecuteCmdSSH(cmd, cfg.Targets[0].Host, cfg.Targets[0].Keypath)
	if err != nil {
		return err
	}

	rCmd := []string{"-av", "--delete", "-e", "ssh", "database:/root/*.sql", "/mnt/pve/external/postgresql_backup"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "toExternal", cfg.Targets[0].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that copies previous database dump to
// HDD present in proxmox instance
func backupPostgresqlToHDD(cfg internal.Config) error {
	//	c := []string{"/mnt/pve/external/postgresql_backup/waiw.sql", "/mnt/pve/external/postgresql_backup/leaks.sql", "/storagepool/backups/postgresql_backup"}
	c := []string{"-av", "--delete", "/mnt/pve/external/postgresql_backup/waiw.sql", "/storagepool/backups/postgresql_backup"}

	err := internal.ExecCmdToProm("rsync", c, "toStoragePool", cfg.Targets[0].Instance, cfg.Pushgateway.Host)
	if err != nil {
		return err
	}

	return nil
}

// Function that handles both backups
func ExecutePostgreSQLBackup(cfg internal.Config, el *internal.ElapsedTime) error {
	start := time.Now()
	if err := backupPostgresqlToExternal(cfg); err != nil {
		return err
	}

	if err := backupPostgresqlToHDD(cfg); err != nil {
		return err
	}

	// Calculate run time
	end := time.Now()
	el.Target = cfg.Targets[0].Name
	el.Elapsed = end.Sub(start).Seconds()

	return nil
}
