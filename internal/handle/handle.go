package handle

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/BrunoTeixeira1996/gbackup/internal/run"
)

type Demand struct {
	Args run.Args
}

// handles POST request when performing backup on demand
// meaning that I can always use something like
// curl -X POST http://pinute.lan:8000/backup -d '{"operation": ""}' -v
// to perform a backup whenever I want
// this also works with telegram bot
func (d *Demand) BackupHandle(w http.ResponseWriter, r *http.Request) {
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

	// Executes logic to backup
	if err := run.Run(d.Args); err != nil {
		log.Printf(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Executed gbackup on demand! Check logs for more info"))
	}
}

// var so it can be mocked in tests
var listenAndServe = http.ListenAndServe

// StartWebHook starts the webhook server
func StartWebHook(args run.Args) {
	demand := Demand{
		Args: args,
	}

	log.Println("started webhook ... ")
	mux := http.NewServeMux()
	mux.HandleFunc("/backup", demand.BackupHandle)
	if err := listenAndServe(":8000", mux); err != nil {
		log.Printf("[handle error] webhook server stopped: %v", err)
	}
}
