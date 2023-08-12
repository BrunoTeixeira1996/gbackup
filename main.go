package main

import (
	"fmt"
	"sync"

	"github.com/BrunoTeixeira1996/gbackup/internal"
	"github.com/BrunoTeixeira1996/gbackup/targets"
)

var supportedTargets = []string{
	"leaks_backup",
	"postgresql_backup",
	"gokr_perm_backup",
	"gokr_config_backup",
	"syncthing_backup",
	"monitoring_backup",
}

// Function that executes backup based on target type
func getExecutionFunction(target string, cfg internal.Config) error {
	switch target {
	case "postgresql_backup":
		if err := targets.ExecutePostgreSQLBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}

	case "gokr_perm_backup":
		if err := targets.ExecuteGokrPermBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}

	case "gokr_config_backup":
		if err := targets.ExecuteGokrConfBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}

	case "syncthing_backup":
		if err := targets.ExecuteSyncthingBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}

	case "monitoring_backup":
		if err := targets.ExecuteMonitoringBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}
	case "leaks_backup":
		if err := targets.ExecuteLeaksBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}
	}

	return nil
}

func logic() error {

	var (
		cfg     internal.Config
		err     error
		wg      sync.WaitGroup
		success int
	)

	if cfg, err = internal.ReadTomlFile(); err != nil {
		internal.Logger.Fatal(err)
	}

	for _, t := range supportedTargets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			internal.Logger.Printf("Starting %s\n\n", t)
			if err := getExecutionFunction(t, cfg); err != nil {
				internal.Logger.Println(err)
			} else {
				success += 1
			}
		}(t)
	}
	wg.Wait()

	res := fmt.Sprintf("Total backups: %d\nTotal backup completed successfully: %d\n\n", len(supportedTargets), success)

	if err := internal.SendEmail(res); err != nil {
		internal.Logger.Printf(err.Error())
	}

	internal.Logger.Printf(res)

	return nil
}

func main() {
	if err := logic(); err != nil {
		internal.Logger.Printf(err.Error())
	}
}
