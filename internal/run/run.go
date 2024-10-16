package run

import (
	"context"
	"fmt"
	"log"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/email"
	"github.com/BrunoTeixeira1996/gbackup/internal/nas"
	"github.com/BrunoTeixeira1996/gbackup/internal/proxmox"
	"github.com/BrunoTeixeira1996/gbackup/internal/setup"
	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

type Args struct {
	Cfg            config.Config
	ConfigPathFlag string
	DebugFlag      bool
}

func Run(args Args) error {
	var (
		ctx        = context.Background()
		setupOK    bool
		tsExternal = &utils.TargetSize{}
	)

	log.Printf("[setup backup info] validating setup\n")
	if args.Cfg, setupOK = setup.IsEverythingConfigured(args.ConfigPathFlag, args.DebugFlag); !setupOK {
		return fmt.Errorf("[run error] please configure the setup properly")
	}
	utils.Body("[SETUP] OK")

	log.Printf("[run info] verifying nas (%s) status\n", args.Cfg.NAS.Name)
	if err := nas.Wakeup(args.Cfg.NAS, ctx); err != nil {
		return fmt.Errorf("[run error] could not wake up nas (%s): %s", args.Cfg.NAS.Name, err)
	}
	log.Printf("[run info] nas (%s) status OK\n", args.Cfg.NAS.Name)
	utils.Body("[NAS] OK")

	external := targets.InitExternal(args.Cfg)
	ts := targets.InitTargets(args.Cfg)

	// check external folder size before backup
	external.VerifyExternalSize("before", tsExternal)

	// execute the backup
	results := targets.ExecuteTargetsBackups(ts, args.Cfg)

	external.VerifyExternalSize("after", tsExternal)

	// add external target size before and after to gather results
	results = append(results, targets.BackupResult{TargetName: "external", ElapsedTime: utils.ElapsedTime{}, TargetSize: *tsExternal})
	targets.DisplayFinalResults(results)

	utils.Body("[BACKUP TARGETS] FINISHED")

	log.Printf("[run info] backup targets finished ... proceeding with external backup to NAS\n")
	if err := targets.ExecuteExternalToNASBackup(external, args.Cfg); err != nil {
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
	if !args.DebugFlag {
		log.Printf("[run info] shutting down nas (%s)\n", args.Cfg.NAS.Name)
		if err := nas.Shutdown(args.Cfg.NAS); err != nil {
			return fmt.Errorf("[run error] could not shut down nas (%s): %s", args.Cfg.NAS.Name, err)
		}
		log.Printf("[run info] nas (%s) off\n", args.Cfg.NAS.Name)
		utils.Body("[NAS] Shutdown OK")
	}

	e := email.EmailClient{}
	e.InitEmailClient()
	var logPathFile string

	if !args.DebugFlag {
		logPathFile = "/var/log/gbackup/gbackup.err.log"
	} else {
		logPathFile = "/home/brun0/Desktop/personal/gbackup/internal/email/testlog.txt"
	}

	if err := e.SendEmail(results, logPathFile); err != nil {
		log.Println(err)
	}

	return nil
}
