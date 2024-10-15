package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/nas"
	"github.com/BrunoTeixeira1996/gbackup/internal/proxmox"
	"github.com/BrunoTeixeira1996/gbackup/internal/setup"
	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

// FIXME: move this to a different package (handle)
// Handles POST to backup on demand
func (d *Args) backupHandle(w http.ResponseWriter, r *http.Request) {
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
	if err := run(*d); err != nil {
		log.Printf(err.Error())
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Executed gbackup on demand! Check logs for more info"))
	}

}

// FIXME: move this to a different package (handle)
func StartWebHook(args Args) {
	log.Println("started webhook ... ")
	http.HandleFunc("/backup", args.backupHandle)
	http.ListenAndServe(":8000", nil)
}

type Args struct {
	cfg            config.Config
	configPathFlag string
	debugFlag      bool
}

func run(args Args) error {
	var (
		ctx        = context.Background()
		setupOK    bool
		tsExternal = &utils.TargetSize{}
	)

	log.Printf("[setup backup info] validating setup\n")
	if args.cfg, setupOK = setup.IsEverythingConfigured(args.configPathFlag, args.debugFlag); !setupOK {
		return fmt.Errorf("[run error] please configure the setup properly")
	}
	utils.Body("[SETUP] OK")

	log.Printf("[run info] verifying nas (%s) status\n", args.cfg.NAS.Name)
	if err := nas.Wakeup(args.cfg.NAS, ctx); err != nil {
		return fmt.Errorf("[run error] could not wake up nas (%s): %s", args.cfg.NAS.Name, err)
	}
	log.Printf("[run info] nas (%s) status OK\n", args.cfg.NAS.Name)
	utils.Body("[NAS] OK")

	external := targets.InitExternal(args.cfg)
	ts := targets.InitTargets(args.cfg)

	// check external folder size before backup
	external.VerifyExternalSize("before", tsExternal)

	// execute the backup
	results := targets.ExecuteTargetsBackups(ts, args.cfg)

	external.VerifyExternalSize("after", tsExternal)

	// add external target size before and after to gather results
	results = append(results, targets.BackupResult{TargetName: "external", ElapsedTime: utils.ElapsedTime{}, TargetSize: *tsExternal})
	targets.DisplayFinalResults(results)

	utils.Body("[BACKUP TARGETS] FINISHED")

	log.Printf("[run info] backup targets finished ... proceeding with external backup to NAS\n")
	if err := targets.ExecuteExternalToNASBackup(external, args.cfg); err != nil {
		log.Printf("[run error] backup from external to NAS was NOT OK: %s\n", err)
	} else {
		utils.Body("[EXTERNAL BACKUP] OK")
	}

	// check PBS backup, if err is nil, that means we can turn off NAS
	log.Printf("[run info] checking PBS backup status\n")
	if err := proxmox.CheckPBSBackupStatus(); err != nil {
		return fmt.Errorf("[run error] could not check PBS backup status ... ignoring turning off NAS: %s\n", err)
	}
	utils.Body("[PBS] Backup OK")

	// we dont want to keep shuting dow NAS while debuging
	if !args.debugFlag {
		log.Printf("[run info] shutting down nas (%s)\n", args.cfg.NAS.Name)
		if err := nas.Shutdown(args.cfg.NAS); err != nil {
			return fmt.Errorf("[run error] could not shut down nas (%s): %s", args.cfg.NAS.Name, err)
		}
		log.Printf("[run info] nas (%s) off\n", args.cfg.NAS.Name)
		utils.Body("[NAS] Shutdown OK")
	}

	return nil
}

func main() {
	utils.Header()
	var (
		cfg            config.Config
		configPathFlag = flag.String("config", "", "location of toml config file")
		debugFlag      = flag.Bool("debug", false, "use this for debug so it does not wait for the cronjob")
	)
	flag.Parse()

	args := Args{
		cfg:            cfg,
		configPathFlag: *configPathFlag,
		debugFlag:      *debugFlag,
	}

	if args.debugFlag {
		if err := run(args); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	} else {
		// used by the on demand backup
		go StartWebHook(args)

		runCh := make(chan struct{})
		go func() {
			for {
				now := time.Now()
				log.Printf("now = %v\n", now)
				if now.Weekday() == time.Friday && now.Hour() >= 13 {
					runCh <- struct{}{}
					time.Sleep(24 * time.Hour) // Sleep for a day to avoid multiple triggers
				} else {
					nextRun := utils.NextFridayAt13(now)
					log.Printf("next run = %v", nextRun)
					time.Sleep(time.Until(nextRun))
				}
			}
		}()

		for range runCh {
			if err := run(args); err != nil {
				log.Fatalf("[main error] could not proceed with gbackup: %s\n", err)
			}
		}
	}

	utils.Footer()
}
