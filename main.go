package main

import (
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

	// if err := targets.ExecutePostgreSQLBackup(cfg); err != nil {
	// 	log.Println(err)
	// }

	// if err := targets.ExecuteGokrPermBackup(cfg); err != nil {
	// 	log.Println(err)
	// }

	// if err := targets.ExecuteSyncthingBackup(cfg); err != nil {
	// 	log.Println(err)
	// }

	if err := targets.ExecuteGokrConfBackup(cfg); err != nil {
		log.Println(err)
	}
}
