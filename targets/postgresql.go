package targets

import "github.com/BrunoTeixeira1996/gbackup/internal"

const (
	host     = "192.168.30.171"
	keypath  = "/home/brun0/.ssh/id_ed25519_postgresql"
	instance = "brun0:test123"
)

func backupPostgresqlToExternal() error {
	cmd := "pg_dump waiw > waiw.sql && pg_dump leaks > leaks.sql"

	err := internal.ExecuteCmdSSH(cmd, host, keypath)
	if err != nil {
		return err
	}

	rCmd := []string{"bot:/root/*.sql", "/tmp/l"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "rsync", instance); err != nil {
		return err
	}

	return nil
}

func backupPostgresqlToHDD() error {
	c := []string{"/tmp/l/waiw.sql", "/tmp/l/leaks.sql", "/tmp/a"}
	err := internal.ExecCmdToProm("cp", c, "cmd", instance)
	if err != nil {
		return err
	}

	return nil
}

func ExecutePostgreSQLBackup() error {
	if err := backupPostgresqlToExternal(); err != nil {
		return err
	}

	if err := backupPostgresqlToHDD(); err != nil {
		return err
	}

	return nil
}
