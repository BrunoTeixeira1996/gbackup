package main

import (
	"log"
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
			log.Println(err)
		}

	case "gokr_perm_backup":
		if err := targets.ExecuteGokrPermBackup(cfg); err != nil {
			log.Println(err)
		}

	case "gokr_config_backup":
		if err := targets.ExecuteGokrConfBackup(cfg); err != nil {
			log.Println(err)
		}

	case "syncthing_backup":
		if err := targets.ExecuteSyncthingBackup(cfg); err != nil {
			log.Println(err)
		}

	case "monitoring_backup":
		if err := targets.ExecuteMonitoringBackup(cfg); err != nil {
			log.Println(err)
		}
	case "leaks_backup":
		if err := targets.ExecuteLeaksBackup(cfg); err != nil {
			log.Println(err)
		}
	}

	return nil
}

func main() {
	var (
		cfg internal.Config
		err error
		wg  sync.WaitGroup
	)

	if cfg, err = internal.ReadTomlFile(); err != nil {
		log.Fatal(err)
	}

	for _, t := range supportedTargets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			log.Printf("Starting %s\n\n", t)
			if err := getExecutionFunction(t, cfg); err != nil {
				log.Println(err)
			}
		}(t)
	}
	wg.Wait()
}
