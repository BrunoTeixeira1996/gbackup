package main

import (
	"log"

	"github.com/BrunoTeixeira1996/gbackup/targets"
)

func main() {
	if err := targets.ExecutePostgreSQLBackup(); err != nil {
		log.Fatal(err)
	}
}
