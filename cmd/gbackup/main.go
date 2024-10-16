package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/forward"
	"github.com/BrunoTeixeira1996/gbackup/internal/handle"
	"github.com/BrunoTeixeira1996/gbackup/internal/run"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

func main() {
	utils.Header()
	var (
		cfg            config.Config
		configPathFlag = flag.String("config", "", "location of toml config file")
		debugFlag      = flag.Bool("debug", false, "use this for debug so it does not wait for the cronjob")
	)
	flag.Parse()

	args := run.Args{
		Cfg:            cfg,
		ConfigPathFlag: *configPathFlag,
		DebugFlag:      *debugFlag,
	}

	if args.DebugFlag {
		if err := run.Run(args); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	} else {
		// used by the on demand backup
		go handle.StartWebHook(args)

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
			if err := run.Run(args); err != nil {
				forward.ForwardMessageToTelegram("FINISHED BACKUP [NOT OK]", forward.Message{Content: "Got an error in the backup", Err: err})
				log.Fatalf("[main error] could not proceed with gbackup: %s\n", err)
			}
		}
	}

	utils.Footer()
}
