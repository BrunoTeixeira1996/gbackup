package proxmox

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
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

// https://forum.proxmox.com/threads/pbs-api.154610/
func (p *PBS) Init() error {
	tokenID := os.Getenv("PBS_TOKENID")
	secret := os.Getenv("PBS_SECRET")

	p.API.TokenID = tokenID
	p.API.Secret = secret
	p.API.Url = "https://nas1.lan:8007/api2/json"
	p.API.Node = "localhost"
	p.API.Authorization = fmt.Sprintf("PBSAPIToken=%s:%s", p.API.TokenID, p.API.Secret)

	return nil
}

// vmidFromWorkerID pulls the vmid (e.g. "ct-101") out of a PBS worker_id,
// which looks like "<datastore>:<vmid>" (e.g. "backupProxmox:ct-101").
func vmidFromWorkerID(workerID string) string {
	_, vmid, found := strings.Cut(workerID, ":")
	if !found {
		return workerID
	}
	return vmid
}

// friendlyName looks up the configured name for a job's vmid, falling back
// to the vmid itself when it's not in the configured list.
func friendlyName(workerID string, idToName map[string]string) string {
	vmid := vmidFromWorkerID(workerID)
	if name, ok := idToName[vmid]; ok {
		return fmt.Sprintf("%s (%s)", name, vmid)
	}
	return vmid
}

// Loops all backups and prune jobs and waits
// for all to finish so gbackup can proceed
func (p *PBS) CheckBackupStatus(objects []config.ProxmoxObject) error {
	var (
		epoch            int64 = utils.Epoch() // epoch time of 12 PM for the current day
		response         []byte
		backups          QueryBackup
		tempBackups      []Backup
		err              error
		totalObjects           = len(objects)
		sleepTime        int64 = 20                    // Sleep time between checks in seconds
		completed              = make(map[string]bool) // Map to track completed backups by "Upid"
		maximumSleepTime int64 = 1800                  // waits 30 minutes before continuing with the program
	)

	idToName := make(map[string]string, len(objects))
	for _, o := range objects {
		idToName[o.ID] = o.Name
	}

	// Loop until all backup and prune jobs are completed
	for {
		// The PBS backup can break (it shouldn't but it might) so I can warn the telegram bot and end the program instead of staying on an infinite loop
		maximumSleepTime -= 20
		if maximumSleepTime == 0 {
			return fmt.Errorf("[pbs info] 30 minutes passed and no PBS backup was completed, so ignoring the PBS backup but please check this")
		}

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
				log.Printf("[pbs info] all %d backup and prune jobs completed\n", len(completed))
				break
			}

			// Only process "prune" and "backup" jobs with "OK" status
			if (b.WorkerType == "prune" || b.WorkerType == "backup") && b.Status == "OK" {
				// Check if the backup job (identified by "Upid") is already completed
				if _, exists := completed[b.Upid]; !exists {
					// Add to tempBackups and mark the job as completed in the map
					tempBackups = append(tempBackups, b)
					completed[b.Upid] = true
					log.Printf("[pbs info] added %s job for %s\n", b.WorkerType, friendlyName(b.WordID, idToName))
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

// var so it can be mocked in tests
var CheckPBSBackupStatus = func(objects []config.ProxmoxObject) error {
	var (
		pve = &PVE{}
		pbs = &PBS{}
		err error
	)

	log.Println("[proxmox info] initializing PBS")
	if err = pbs.Init(); err != nil {
		return err
	}

	// Purely informational: lets you compare against len(objects) in the
	// logs to notice when a new LXC/VM hasn't been added to config.toml yet.
	log.Println("[proxmox info] initializing PVE")
	if err = pve.Init(); err != nil {
		return err
	}

	log.Println("[proxmox info] gathering all objects from PVE")
	if err = pve.GetAllObjects(); err != nil {
		return err
	}

	log.Printf("[proxmox info] total objects on proxmox: %d, configured for backup: %d\n", len(pve.LXCs)+len(pve.VMs), len(objects))

	log.Println("[proxmox info] checking backup status")
	if err := pbs.CheckBackupStatus(objects); err != nil {
		return err
	}
	log.Printf("[proxmox info] all backups completed successfully and have 'OK' status.\n")

	return nil
}
