package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal"
	"github.com/BrunoTeixeira1996/gbackup/targets"
)

const version = "2.0"

var supportedTargets = []string{
	// "leaks_backup",
	"postgresql_backup",
	"gokr_perm_backup",
	"gokr_config_backup",
	"syncthing_backup",
	"monitoring_backup",
	"work_laptop",
}

// Function that executes backup based on target type
func getExecutionFunction(target string, cfg internal.Config) error {
	switch target {
	case "postgresql_backup":
		if err := targets.ExecutePostgreSQLBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}

	case "gokr_perm_backup":
		if err := targets.ExecuteGokrPermBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}

	case "gokr_config_backup":
		if err := targets.ExecuteGokrConfBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}

	case "syncthing_backup":
		if err := targets.ExecuteSyncthingBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}

	case "monitoring_backup":
		if err := targets.ExecuteMonitoringBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}
	case "leaks_backup":
		if err := targets.ExecuteLeaksBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}
	case "work_laptop":
		if err := targets.ExecuteWorkLaptopBackup(cfg); err != nil {
			internal.Logger.Println(err)
		}
	}

	return nil
}

func logic() error {
	var (
		cfg     internal.Config
		err     error
		wg      sync.WaitGroup
		success int
	)

	if cfg, err = internal.ReadTomlFile(); err != nil {
		internal.Logger.Fatal(err)
	}

	for _, t := range supportedTargets {
		wg.Add(1)
		go func(t string) {
			internal.Logger.Printf("Starting %s\n\n", t)
			if err := getExecutionFunction(t, cfg); err != nil {
				internal.Logger.Println(err)
			} else {
				success += 1
			}
			wg.Done()
		}(t)
	}
	wg.Wait()

	finalResult := &internal.EmailTemplate{
		Timestamp:          time.Now().String(),
		Totalbackups:       len(supportedTargets),
		Totalbackupsuccess: success,
		PiTemp:             internal.GetPiTemp(),
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
