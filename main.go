package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/proxmox"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
	"github.com/BrunoTeixeira1996/gbackup/targets"
)

const version = "4.0"

var supportedTargets = []string{
	"gokr_perm_backup",
	// "work_laptop",
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

	// Executes logic to backup
	// but returns always if there's an error or no
	if err := run(); err != nil {
		log.Printf(err.Error())
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Executed gbackup on demand! Check logs for more info"))
	}

}

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
			log.Printf("[get execution error] could not get folder size for %s\n", cfg.Targets[0].Name)
		}

		if err := targets.ExecuteWorkLaptopBackup(cfg, el); err != nil {
			log.Println(err)
		}

		ts.After, err = utils.GetFolderSize(cfg.Targets[0].ExternalPath)
		if err != nil {
			log.Printf("[get execution error] could not get folder size for %s on the second run\n", cfg.Targets[0].Name)
		}

		ts.Name = cfg.Targets[0].Name

	case "gokr_perm_backup":
		ts.Before, err = utils.GetFolderSize(cfg.Targets[1].ExternalPath)
		if err != nil {
			log.Printf("[get execution error] could not get folder size for %s\n", cfg.Targets[1].Name)
		}

		if err := targets.ExecuteGokrPermBackup(cfg, el); err != nil {
			log.Println(err)
		}

		ts.After, err = utils.GetFolderSize(cfg.Targets[1].ExternalPath)
		if err != nil {
			log.Printf("[get execution error] could not get folder size for %s on the second run\n", cfg.Targets[1].Name)
		}

		ts.Name = cfg.Targets[1].Name
	}
	fmt.Printf("\n\n")

	return nil
}

func run() error {
	var (
		//ctx            = context.Background()
		configPathFlag = flag.String("config", "", "location of toml config file")
		cfg            config.Config
		err            error
		wg             sync.WaitGroup
		success        int
		times          []utils.ElapsedTime
		targetsSize    []utils.TargetSize
	)

	flag.Parse()

	log.Printf("[run info] validating setup\n")
	if !isEverythingConfigured(*configPathFlag) {
		return fmt.Errorf("[run error] please configure the setup properly")
	}
	log.Printf("[run info] setup is OK\n")

	log.Printf("[run info] reading toml file\n")
	if cfg, err = config.ReadTomlFile(*configPathFlag); err != nil {
		return fmt.Errorf("[run error] could not read toml file properly: %s", err)
	}
	log.Printf("[run info] toml file is OK\n\n")

	/*DEBUG FOR NOW*/
	// log.Printf("[run info] verifying nas (%s) status\n", cfg.NAS.Name)
	// if err := nas.Wakeup(cfg.NAS, ctx); err != nil {
	// 	return fmt.Errorf("[run error] could not wake up nas (%s): %s", cfg.NAS.Name, err)
	// }
	// log.Printf("[run info] nas (%s) status OK\n", cfg.NAS.Name)
	/*DEBUG FOR NOW*/

	for _, t := range supportedTargets {
		wg.Add(1)
		el := &utils.ElapsedTime{}
		ts := &utils.TargetSize{}
		go func(t string) {
			log.Printf("[run info] starting backup %s\n", t)
			if err := getExecutionFunction(t, cfg, el, ts); err != nil {
				log.Println(err)
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

	/*	log.Printf("[run info] backup targets finished ... proceeding with external backup to NAS\n")
		if err := targets.ExecuteExternalToNASBackup(cfg); err != nil {
			log.Println(err)
		}*/

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

	/*DEBUG FOR NOW*/
	// log.Printf("[run info] shutting down nas (%s)\n", cfg.NAS.Name)
	// if err := internal.Shutdown(cfg.NAS); err != nil {
	// 	return fmt.Errorf("[run error] could not shut down nas (%s): %s", cfg.NAS.Name, err)
	// }
	// log.Printf("[run info] nas (%s) off\n", cfg.NAS.Name)

	/*DEBUG FOR NOW*/

	return nil
}

func isEverythingConfigured(configPathFlag string) bool {
	log.Printf("[setup info] validating config flag\n")
	if configPathFlag == "" {
		log.Printf("[setup error] please provide the path for the config file\n")
		return false
	}
	log.Printf("[setup info] config flag OK\n")

	log.Printf("[setup info] validating env vars \n")
	senderEmail := os.Getenv("SENDEREMAIL")
	senderPass := os.Getenv("SENDERPASS")
	if senderEmail == "" || senderPass == "" {
		log.Printf("[setup error] SENDEREMAIL or SENDERPASS are not present\n")
		return false
	}
	log.Printf("[setup info] env vars are OK\n")

	log.Printf("[setup info] validating mount point\n")
	if !utils.IsExternalMounted() {
		log.Printf("[setup error] mount point is not mounted in the system\n")
		return false
	}
	log.Printf("[setup info] mount point is OK\n")

	return true
}

func main() {
	proxmox.Test()
	/*	if err := run(); err != nil {
			log.Println(err.Error())
		}
	*/
	// log.Println("Running version:", version)

	// // used by the on demand backup
	// go StartWebHook()

	// runCh := make(chan struct{})
	// go func() {
	// 	// Run forever, trigger a run at 17:00 every Friday.
	// 	for {
	// 		now := time.Now()
	// 		runTodayHour := now.Hour() < 17
	// 		runTodayDay := now.Weekday().String() == "Friday"
	// 		today := now.Day()
	// 		log.Printf("now = %v, runTodayDay = %v", now, runTodayDay)
	// 		for {
	// 			if time.Now().Day() != today {
	// 				// Day changed, re-evaluate whether to run today.
	// 				break
	// 			}
	// 			// If today is not Friday, sleep until next day and re-evaluate
	// 			if !runTodayDay {
	// 				nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	// 				hoursLeft := nextDay.Sub(now)
	// 				log.Printf("Sleeping until next day ... %v hours to go", hoursLeft)
	// 				time.Sleep(time.Until(nextDay))
	// 				break
	// 			}

	// 			// Today is Friday, so wait until 17:00
	// 			nextHour := time.Now().Truncate(time.Hour).Add(1 * time.Hour)
	// 			log.Printf("today = %d, todayIsFriday = %v, todayHour = %v next hour: %v", today, runTodayDay, runTodayHour, nextHour)
	// 			time.Sleep(time.Until(nextHour))

	// 			if time.Now().Hour() >= 17 && runTodayHour && now.Weekday().String() == "Friday" {
	// 				runTodayHour = false
	// 				runTodayDay = false
	// 				runCh <- struct{}{}
	// 			}
	// 		}
	// 	}
	// }()

	// for range runCh {
	// 	if err := logic(); err != nil {
	// 		log.Printf(err.Error())
	// 	}
	// }
}
