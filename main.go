package main

import (
	"fmt"
	"log"

	"github.com/BrunoTeixeira1996/gbackup/internal"
	"github.com/BrunoTeixeira1996/gbackup/targets"
)

func main() {
	var (
		cfg internal.Config
		err error
	)

	if cfg, err = internal.ReadTomlFile(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting Postgresql backup\n")
	if err := targets.ExecutePostgreSQLBackup(cfg); err != nil {
		log.Println(err)
	}
	fmt.Printf("-------------------------------\n")

	fmt.Printf("Starting gokr perm partition backup\n")
	if err := targets.ExecuteGokrPermBackup(cfg); err != nil {
		log.Println(err)
	}
	fmt.Printf("-------------------------------\n")

	fmt.Printf("Starting Syncthing backup\n")
	if err := targets.ExecuteSyncthingBackup(cfg); err != nil {
		log.Println(err)
	}
	fmt.Printf("-------------------------------\n")

	fmt.Printf("Starting gokr config backup\n")
	if err := targets.ExecuteGokrConfBackup(cfg); err != nil {
		log.Println(err)
	}
	fmt.Printf("-------------------------------\n")

	fmt.Printf("Starting Leaks backup\n")
	if err := targets.ExecuteLeaksBackup(cfg); err != nil {
		log.Println(err)
	}
	fmt.Printf("-------------------------------\n")

	fmt.Printf("Starting monitoring backup\n")
	if err := targets.ExecuteMonitoringBackup(cfg); err != nil {
		log.Println(err)
	}
	fmt.Printf("-------------------------------\n")
}
