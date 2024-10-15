package proxmox

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

type PBS struct {
	API Boilerplate
}

type QueryBackup struct {
	Total      int      `json:"total"`
	DataBackup []Backup `json:"data"`
}

type Backup struct {
	Upid       string  `json:"upid"`
	Node       string  `json:"node"`
	Pid        float64 `json:"pid"`
	Pstart     float64 `json:"pstart"`
	Starttime  float64 `json:"starttime"`
	WorkerType string  `json:"worker_type"`
	WordID     string  `json:"worker_id"`
	User       string  `json:"user"`
	EndTime    float64 `json:"endtime"`
	Status     string  `json:"status"`
}

func (p *PBS) Init() error {
	tokenID := os.Getenv("PBS_TOKENID")
	secret := os.Getenv("PBS_SECRET")

	p.API.TokenID = tokenID
	p.API.Secret = secret
	p.API.Url = "https://192.168.30.200:8007/api2/json"
	p.API.Node = "localhost"
	p.API.Authorization = fmt.Sprintf("PBSAPIToken=%s:%s", p.API.TokenID, p.API.Secret)

	return nil
}

// Loops all backups and prune jobs and waits
// for all to finish so gbackup can proceed
func (p *PBS) checkBackupStatus(totalObjects int) error {
	var (
		epoch       int64 = utils.Epoch() // epoch time of 12 PM for the current day
		response    []byte
		backups     QueryBackup
		tempBackups []Backup
		err         error
		sleepTime   int64 = 20                    // Sleep time between checks in seconds
		completed         = make(map[string]bool) // Map to track completed backups by "Upid"
	)

	// Loop until all backup and prune jobs are completed
	for {
		log.Println("[pbs info] checking backup status...")

		// Fetch backup and prune jobs since the epoch
		apiPath := fmt.Sprintf("tasks?since=%d", epoch)
		if response, err = p.API.request("GET", apiPath); err != nil {
			return fmt.Errorf("[pbs error] could not check backup status: %s\n", err)
		}

		if err = json.Unmarshal(response, &backups); err != nil {
			return fmt.Errorf("[pbs error] could not unmarshal response to Backup: %s\n", err)
		}

		// Process backups and avoid duplicates using the map
		for _, b := range backups.DataBackup {
			// If we reach the expected number of jobs, exit the loop
			if len(completed) == totalObjects*2 {
				log.Printf("[pbs error] all %d backup and prune jobs completed\n", len(completed))
				break
			}

			// Only process "prune" and "backup" jobs with "OK" status
			if (b.WorkerType == "prune" || b.WorkerType == "backup") && b.Status == "OK" {
				// Check if the backup job (identified by "Upid") is already completed
				if _, exists := completed[b.Upid]; !exists {
					// Add to tempBackups and mark the job as completed in the map
					tempBackups = append(tempBackups, b)
					completed[b.Upid] = true
					log.Printf("[pbs info] added backup job %s (type: %s)\n", b.Upid, b.WorkerType)
				}
			}
		}

		// Check if all backups are done
		if len(completed) == totalObjects*2 {
			log.Printf("[pbs info] successfully completed %d backup and prune jobs.\n", len(completed))
			break
		}

		// Sleep before retrying
		log.Printf("[pbs info] incomplete jobs (%d/%d), sleeping for %d seconds...\n", len(completed), totalObjects*2, sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	// Final check to ensure all backups have "OK" status
	for _, b := range tempBackups {
		if b.Status != "OK" {
			return fmt.Errorf("[pbs error] backup: %s was not OK\n", b.Upid)
		}
	}

	return nil
}

func CheckPBSBackupStatus() error {
	var (
		pve          = &PVE{}
		pbs          = &PBS{}
		totalObjects int
		err          error
	)

	log.Println("[proxmox info] initializing PBS")
	if err = pbs.Init(); err != nil {
		return err
	}

	log.Println("[proxmox info] initializing PVE")
	if err = pve.Init(); err != nil {
		return err
	}

	log.Println("[proxmox info] gathering all objects from PVE")
	if err = pve.getAllObjects(); err != nil {
		return err
	}

	totalObjects = len(pve.LXCs) + len(pve.VMs)
	log.Printf("[proxmox info] total objects: %d\n", totalObjects)

	log.Println("[proxmox info] checking backup status")
	if err := pbs.checkBackupStatus(totalObjects); err != nil {
		return err
	}
	log.Printf("[proxmox info] all backups completed successfully and have 'OK' status.\n")

	return nil
}
