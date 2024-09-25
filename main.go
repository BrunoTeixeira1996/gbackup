package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/nas"
	"github.com/BrunoTeixeira1996/gbackup/internal/proxmox"
	"github.com/BrunoTeixeira1996/gbackup/internal/setup"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"

	"github.com/BrunoTeixeira1996/gbackup/targets"
)

const version = "4.0"

type backupResult struct {
	elapsed    utils.ElapsedTime
	targetSize utils.TargetSize
	err        error
}

var supportedTargets = []string{
	"gokr_perm_backup",
	"work_laptop",
	"pinute_backup",
}

// Handles POST to backup on demand
// TODO clean this code as well as logs
func backupHandle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != "POST" {
		http.Error(w, "NOT POST!", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	newBackup := struct {
		Op string `json:"operation"`
	}{}

	if err := decoder.Decode(&newBackup); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error while unmarshal json response:", err)
		fmt.Fprintf(w, "Please provide a valid POST body with the operation you want\n")
		return
	}

	log.Printf("Executing backup on demand with operation: %s\n", newBackup.Op)
	fmt.Fprintf(w, "Executing backup on demand with operation: %s\n", newBackup.Op)

	// Executes logic to backup
	// but returns always if there's an error or no
	if err := run(); err != nil {
		log.Printf(err.Error())
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Executed gbackup on demand! Check logs for more info"))
	}

}

// TODO clean this code as well as logs
func StartWebHook() {
	log.Println("started webhook ... ")
	http.HandleFunc("/backup", backupHandle)
	http.ListenAndServe(":8000", nil)
}

// Function that executes backup based on target type
// FIXME: Clean duplicated code
func getExecutionFunction(target string, cfg config.Config, el *utils.ElapsedTime, ts *utils.TargetSize) error {
	var err error

	switch target {
	case "work_laptop":
		ts.Before, err = utils.GetFolderSize(cfg.Targets[0].ExternalPath)
		if err != nil {
			log.Printf("[get execution error] could not get folder size for %s: %s\n", cfg.Targets[0].Name, err)
		}

		if err := targets.ExecuteWorkLaptopBackup(cfg, el); err != nil {
			log.Println(err)
		}

		ts.After, err = utils.GetFolderSize(cfg.Targets[0].ExternalPath)
		if err != nil {
			log.Printf("[get execution error] could not get folder size for %s on the second run: %s\n", cfg.Targets[0].Name, err)
		}

		ts.Name = cfg.Targets[0].Name

	case "gokr_perm_backup":
		ts.Before, err = utils.GetFolderSize(cfg.Targets[1].ExternalPath)
		if err != nil {
			log.Printf("[get execution error] could not get folder size for %s: %s\n", cfg.Targets[1].Name, err)
		}

		if err := targets.ExecuteGokrPermBackup(cfg, el); err != nil {
			log.Println(err)
		}

		ts.After, err = utils.GetFolderSize(cfg.Targets[1].ExternalPath)
		if err != nil {
			log.Printf("[get execution error] could not get folder size for %s on the second run: %s\n", cfg.Targets[1].Name, err)
		}

		ts.Name = cfg.Targets[1].Name

	case "pinute_backup":
		ts.Before, err = utils.GetFolderSize(cfg.Targets[2].ExternalPath)
		if err != nil {
			log.Printf("[get execution error] could not get folder size for %s: %s\n", cfg.Targets[2].Name, err)
		}

		if err := targets.ExecutePinuteBackup(cfg, el); err != nil {
			log.Println(err)
		}

		ts.After, err = utils.GetFolderSize(cfg.Targets[2].ExternalPath)
		if err != nil {
			log.Printf("[get execution error] could not get folder size for %s on the second run: %s\n", cfg.Targets[2].Name, err)
		}

		ts.Name = cfg.Targets[2].Name
	}
	log.Printf("\n\n")

	return nil
}

func run() error {
	var (
		ctx            = context.Background()
		configPathFlag = flag.String("config", "", "location of toml config file")
		cfg            config.Config
		setupOK        bool
		wg             sync.WaitGroup
		times          []utils.ElapsedTime
		targetsSize    []utils.TargetSize
		results        = make([]backupResult, len(supportedTargets)) // Slice to store backup results in order

	)

	flag.Parse()

	log.Printf("[setup backup info] validating setup\n")
	if cfg, setupOK = setup.IsEverythingConfigured(*configPathFlag); !setupOK {
		return fmt.Errorf("[run error] please configure the setup properly")
	}
	utils.Body("[SETUP] OK")

	log.Printf("[run info] verifying nas (%s) status\n", cfg.NAS.Name)
	if err := nas.Wakeup(cfg.NAS, ctx); err != nil {
		return fmt.Errorf("[run error] could not wake up nas (%s): %s", cfg.NAS.Name, err)
	}
	log.Printf("[run info] nas (%s) status OK\n", cfg.NAS.Name)
	utils.Body("[NAS] OK")

	// Launch each backup in its own goroutine
	for i, t := range supportedTargets {
		wg.Add(1)
		go func(i int, t string) {
			defer wg.Done()
			el := &utils.ElapsedTime{}
			ts := &utils.TargetSize{}
			log.Printf("[run info] starting backup %s\n", t)
			err := getExecutionFunction(t, cfg, el, ts)

			// Store the result for this target
			results[i] = backupResult{elapsed: *el, targetSize: *ts, err: err}
		}(i, t)
	}
	// Wait for all backups to complete
	wg.Wait()

	// Log results in order
	for i, res := range results {
		t := supportedTargets[i]
		if res.err != nil {
			log.Printf("[run error] backup %s failed: %s", t, res.err)
		} else {
			log.Printf("[run info] backup %s completed. Elapsed: %v, Size Before: %v, Size After: %v",
				t, res.elapsed, res.targetSize.Before, res.targetSize.After)
		}
		// Collect times and sizes for further processing
		times = append(times, res.elapsed)
		targetsSize = append(targetsSize, res.targetSize)
	}
	utils.Body("[BACKUP TARGETS] FINISHED")

	log.Printf("[run info] backup targets finished ... proceeding with external backup to NAS\n")
	if err := targets.ExecuteExternalToNASBackup(cfg); err != nil {
		log.Printf("[run error] backtup from external to NAS was NOT OK: %s\n", err)
	} else {
		utils.Body("[EXTERNAL BACKUP] OK")
	}

	// finalResult := &email.EmailTemplate{
	// 	Timestamp:          time.Now().String(),
	// 	Totalbackups:       len(supportedTargets),
	// 	Totalbackupsuccess: success,
	// 	PiTemp:             pi.GetPiTemp(),
	// 	ElapsedTimes:       times,
	// 	TotalElapsedTime:   utils.CalculateTotalElaspedTime(times),
	// 	TargetsSize:        targetsSize,
	// }

	//log.Printf("[run info] preparing email fields\n")
	// if err := email.SendEmail(finalResult); err != nil {
	// 	log.Printf("[run error] could not send email: %s", err)
	// }

	// check PBS backup, if err is nil, that means we can turn off NAS
	log.Printf("[run info] checking PBS backup status\n")
	if err := proxmox.CheckPBSBackupStatus(); err != nil {
		return fmt.Errorf("[run error] could not check PBS backup status ... ignoring turning off NAS: %s\n", err)
	}
	utils.Body("[PBS] Backup OK")

	log.Printf("[run info] shutting down nas (%s)\n", cfg.NAS.Name)
	if err := nas.Shutdown(cfg.NAS); err != nil {
		return fmt.Errorf("[run error] could not shut down nas (%s): %s", cfg.NAS.Name, err)
	}
	log.Printf("[run info] nas (%s) off\n", cfg.NAS.Name)
	utils.Body("[NAS] Shutdown OK")

	return nil
}

func main() {
	utils.Header(version)
	// Uncoment this if want to debug and run on command
	// if err := run(); err != nil {
	// 	log.Fatalf("[main error] could not proceed with gbackup: %s\n", err.Error())
	// }

	// // used by the on demand backup
	go StartWebHook()

	runCh := make(chan struct{})
	go func() {
		// Run forever, trigger a run at 13:00 every Friday.
		for {
			now := time.Now()
			runTodayHour := now.Hour() < 13
			runTodayDay := now.Weekday().String() == "Friday"
			today := now.Day()
			log.Printf("now = %v, runTodayDay = %v", now, runTodayDay)
			for {
				if time.Now().Day() != today {
					// Day changed, re-evaluate whether to run today.
					break
				}
				// If today is not Friday, sleep until next day and re-evaluate
				if !runTodayDay {
					nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
					hoursLeft := nextDay.Sub(now)
					log.Printf("Sleeping until next day ... %v hours to go", hoursLeft)
					time.Sleep(time.Until(nextDay))
					break
				}

				// Today is Friday, so wait until 13:00
				nextHour := time.Now().Truncate(time.Hour).Add(1 * time.Hour)
				log.Printf("today = %d, todayIsFriday = %v, todayHour = %v next hour: %v", today, runTodayDay, runTodayHour, nextHour)
				time.Sleep(time.Until(nextHour))

				if time.Now().Hour() >= 13 && runTodayHour && now.Weekday().String() == "Friday" {
					runTodayHour = false
					runTodayDay = false
					runCh <- struct{}{}
				}
			}
		}
	}()

	for range runCh {
		if err := run(); err != nil {
			log.Fatalf("[main error] could not proceed with gbackup: %s\n", err.Error())
		}
	}
	utils.Footer(version)
}
