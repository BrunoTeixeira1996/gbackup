package proxmox

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

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

type PBS struct {
	TokenID       string
	Secret        string
	APIUrl        string
	Node          string
	Authorization string
}

func (p *PBS) Init() error {
	tokenID := os.Getenv("PBS_TOKENID")
	secret := os.Getenv("PBS_SECRET")

	if tokenID == "" || secret == "" {
		return fmt.Errorf("[pbs error] please provide the PBS token and secret env vars\n")
	}

	p.TokenID = tokenID
	p.Secret = secret
	p.APIUrl = "https://192.168.30.200:8007/api2/json"
	p.Node = "localhost"
	p.Authorization = fmt.Sprintf("PBSAPIToken=%s:%s", p.TokenID, p.Secret)

	return nil
}

func request(pbs PBS, rType, apiPath string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}

	requestBody := fmt.Sprintf("%s/nodes/%s/%s", pbs.APIUrl, pbs.Node, apiPath)

	req, err := http.NewRequest(rType, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("[pbs error] could not create new request:%s\n", err)
	}

	req.Header.Set("Authorization", pbs.Authorization)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[pbs error] could not perform client.Do: %s\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[pbs error] could not perform %s request to: %s - status code: %d\n", rType, requestBody, resp.StatusCode)
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[pbs error] could not read response body: %s\n", err)
	}

	return res, nil
}

func checkBackupStatus(pbs PBS) error {
	var (
		epoch       int64
		response    []byte
		backups     QueryBackup
		tempBackups []Backup
		counter     int
		isDup       bool
		err         error
	)

	// get epoch time for 12 PM from current day
	epoch = utils.Epoch()

	// https://192.168.30.200:8007/api2/json/nodes/localhost/tasks?since=1726657200
	path := fmt.Sprintf("tasks?since=%d", epoch)
	if response, err = request(pbs, "GET", path); err != nil {
		return fmt.Errorf("[checkBackupStatus] could not check backup status: %s\n", err)
	}

	if err = json.Unmarshal(response, &backups); err != nil {
		return fmt.Errorf("[checkBackupStatus] could not unmarshal response to Backup: %s\n", err)
	}

	// FIXME: I dont like to rely on a fixed number ... i need a better approach to this
	// if I add another vm or LXC i need to come here and update this again ...
	// a good approach would be to find the number of vm's + lxc dynamicaly
	// this is possible using PVE API instead
	// https://192.168.30.3:8006/api2/json/nodes/localhost/lxc -> lxc
	// https://192.168.30.3:8006/api2/json/nodes/localhost/qemu -> VMs
	for _, b := range backups.DataBackup {
		// if there's 16 entries in tempBackups that means all backups were executed
		if len(tempBackups) == 16 {
			break
		}

		// we only care about prune jobs and backups
		// FIXME: there could be a bug in here since I dont want dup backups in tempBackups
		if b.WorkerType == "prune" || b.WorkerType == "backup" && b.Status == "OK" {
			for _, tempB := range tempBackups {
				if tempB.Starttime == b.Starttime {
					isDup = true
					break
				}
			}
			if !isDup {
				tempBackups = append(tempBackups, b)
			}
		}

		time.Sleep(2 * time.Minute)
		isDup = false
	}

	for _, b := range tempBackups {
		if b.Status != "OK" {
			log.Printf("[checkBackupStatus] backup: %s was not OK\n", b.Upid)
		}
		counter++
	}

	if counter == len(tempBackups) {
		log.Printf("[checkBackupStatus] all backups are OK\n")
	}

	/*
		TODO: get current date epoch time
		if result, err = request(pbs, "GET", "tasks?since=epochtime"); err != nil {
			log.Println(err)
		}

		validate that return contains 16 entries (number of vms and lxc backup 8 and prune 8)
		if not, sleep 2 mins and try again
		when it has 16 entries, check the status
		if all status OK then the pbs finished backup and everything worked fine
	*/

	return nil
}

func Test() {
	var (
		pbs PBS
		err error
	)

	if err = pbs.Init(); err != nil {
		log.Println(err)
		return
	}

	if err := checkBackupStatus(pbs); err != nil {
		log.Println(err)
	}
}
