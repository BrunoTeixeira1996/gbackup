package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal"
	"github.com/BrunoTeixeira1996/gbackup/targets"
)

const version = "3.0"

var supportedTargets = []string{
	// "leaks_backup",
	"postgresql_backup",
	"gokr_perm_backup",
	"gokr_config_backup",
	"syncthing_backup",
	"monitoring_backup",
	"work_laptop",
}

// Handles POST to backup on demand
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
	if err := logic(); err != nil {
		internal.Logger.Printf(err.Error())
	}
}

func StartWebHook() {
	log.Println("Started webhook ... ")
	http.HandleFunc("/backup", backupHandle)
	http.ListenAndServe(":8000", nil)
}

// Function that executes backup based on target type
// FIXME: Clean duplicated code
func getExecutionFunction(target string, cfg internal.Config, el *internal.ElapsedTime, ts *internal.TargetSize) error {
	var err error

	switch target {
	case "postgresql_backup":
		// GetFolderSize before backup
		ts.Before, err = internal.GetFolderSize(cfg.Targets[0].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[0].Name)
		}

		if err := targets.ExecutePostgreSQLBackup(cfg, el); err != nil {
			internal.Logger.Println(err)
		}

		// GetFolderSize after backup
		ts.After, err = internal.GetFolderSize(cfg.Targets[0].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[0].Name)
		}

		ts.Name = cfg.Targets[0].Name

	case "gokr_perm_backup":
		ts.Before, err = internal.GetFolderSize(cfg.Targets[1].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[1].Name)
		}

		if err := targets.ExecuteGokrPermBackup(cfg, el); err != nil {
			internal.Logger.Println(err)
		}

		ts.After, err = internal.GetFolderSize(cfg.Targets[1].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[1].Name)
		}

		ts.Name = cfg.Targets[1].Name

	case "gokr_config_backup":
		ts.Before, err = internal.GetFolderSize(cfg.Targets[5].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[5].Name)
		}

		if err := targets.ExecuteGokrConfBackup(cfg, el); err != nil {
			internal.Logger.Println(err)
		}

		ts.After, err = internal.GetFolderSize(cfg.Targets[5].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[5].Name)
		}

		ts.Name = cfg.Targets[5].Name

	case "syncthing_backup":
		ts.Before, err = internal.GetFolderSize(cfg.Targets[2].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[2].Name)
		}

		if err := targets.ExecuteSyncthingBackup(cfg, el); err != nil {
			internal.Logger.Println(err)
		}

		ts.After, err = internal.GetFolderSize(cfg.Targets[2].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[2].Name)
		}

		ts.Name = cfg.Targets[2].Name

	case "monitoring_backup":
		ts.Before, err = internal.GetFolderSize(cfg.Targets[4].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[4].Name)
		}

		if err := targets.ExecuteMonitoringBackup(cfg, el); err != nil {
			internal.Logger.Println(err)
		}

		ts.After, err = internal.GetFolderSize(cfg.Targets[4].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[4].Name)
		}

		ts.Name = cfg.Targets[4].Name

	case "leaks_backup":
		if err := targets.ExecuteLeaksBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}

	case "work_laptop":
		ts.Before, err = internal.GetFolderSize(cfg.Targets[6].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[6].Name)
		}

		if err := targets.ExecuteWorkLaptopBackup(cfg, el); err != nil {
			internal.Logger.Println(err)
		}

		ts.After, err = internal.GetFolderSize(cfg.Targets[6].ExternalPath)
		if err != nil {
			log.Printf("[ERROR] Could not get folder size for %s\n", cfg.Targets[6].Name)
		}

		ts.Name = cfg.Targets[6].Name

	}
	return nil
}

func logic() error {
	var (
		cfg         internal.Config
		err         error
		wg          sync.WaitGroup
		success     int
		times       []internal.ElapsedTime
		targetsSize []internal.TargetSize
	)

	if cfg, err = internal.ReadTomlFile(); err != nil {
		internal.Logger.Fatal(err)
	}

	for _, t := range supportedTargets {
		wg.Add(1)
		el := &internal.ElapsedTime{}
		ts := &internal.TargetSize{}
		go func(t string) {
			internal.Logger.Printf("Starting %s\n\n", t)
			if err := getExecutionFunction(t, cfg, el, ts); err != nil {
				internal.Logger.Println(err)
			} else {
				success += 1
			}
			wg.Done()
			// appends every run time of each target
			times = append(times, *el)

			// append every folder size change of each target
			targetsSize = append(targetsSize, *ts)
		}(t)
	}
	wg.Wait()

	finalResult := &internal.EmailTemplate{
		Timestamp:          time.Now().String(),
		Totalbackups:       len(supportedTargets),
		Totalbackupsuccess: success,
		PiTemp:             internal.GetPiTemp(),
		ElapsedTimes:       times,
		TotalElapsedTime:   internal.CalculateTotalElaspedTime(times),
		TargetsSize:        targetsSize,
	}

	if err := internal.SendEmail(finalResult); err != nil {
		internal.Logger.Printf(err.Error())
	}

	return nil
}

func isEverythingConfigured() bool {
	senderEmail := os.Getenv("SENDEREMAIL")
	senderPass := os.Getenv("SENDERPASS")
	if senderEmail != "" && senderPass != "" {
		log.Println("SENDEREMAIL && SENDERPASS are present, so lets continue ...")
		return true
	}

	log.Println("SENDEREMAIL && SENDERPASS not present, so quiting ...")
	return false
}

func main() {
	if !isEverythingConfigured() {
		os.Exit(1)
	}

	log.Println("Running version:", version)

	// used by the on demand backup
	go StartWebHook()

	runCh := make(chan struct{})
	go func() {
		// Run forever, trigger a run at 17:00 every Friday.
		for {
			now := time.Now()
			runTodayHour := now.Hour() < 17
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

				// Today is Friday, so wait until 17:00
				nextHour := time.Now().Truncate(time.Hour).Add(1 * time.Hour)
				log.Printf("today = %d, todayIsFriday = %v, todayHour = %v next hour: %v", today, runTodayDay, runTodayHour, nextHour)
				time.Sleep(time.Until(nextHour))

				if time.Now().Hour() >= 17 && runTodayHour && now.Weekday().String() == "Friday" {
					runTodayHour = false
					runTodayDay = false
					runCh <- struct{}{}
				}
			}
		}
	}()

	for range runCh {
		if err := logic(); err != nil {
			internal.Logger.Printf(err.Error())
		}
	}
}
