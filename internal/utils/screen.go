package utils

import (
	"log"
	"time"
)

func Header(version string) {
	currentDate := time.Now().Format("2006/01/02 15:04:05")

	log.Println("=====================================")
	log.Println("Start Gbackup:", currentDate)
	log.Println("=====================================\n")
}

func Body(txt string) {
	currentDate := time.Now().Format("2006/01/02 15:04:05")

	log.Println("=====================================")
	log.Printf("%s: %s\n", txt, currentDate)
	log.Println("=====================================\n")
}

func Footer(version string) {
	currentDate := time.Now().Format("2006/01/02 15:04:05")

	log.Println("=====================================")
	log.Println("End Gbackup:", currentDate)
	log.Println("=====================================\n\n\n")
	log.Println("------------------------------------------------------------------------------\n\n\n")

}
