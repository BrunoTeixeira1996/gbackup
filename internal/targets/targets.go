package targets

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

type BackupResult struct {
	TargetName  string
	ElapsedTime utils.ElapsedTime
	TargetSize  utils.TargetSize
	Err         error
}

type Target struct {
	Name          string                `toml:"name"`
	IP            string                `toml:"ip"`
	Keypath       string                `toml:"keypath,omitempty"`
	Instance      string                `toml:"instance"`
	MAC           string                `toml:"mac"`
	ExternalPath  string                `toml:"external_path"`
	RsyncCommands []config.RsyncCommand `toml:"rsync_commands"`
}

// InitTargets initializes the targets from the config package.
func InitTargets(cfg config.Config) []Target {
	var targets []Target
	for _, t := range cfg.Targets {
		targets = append(targets, Target{
			Name:          t.Name,
			IP:            t.IP,
			Keypath:       t.Keypath,
			Instance:      t.Instance,
			MAC:           t.MAC,
			ExternalPath:  t.ExternalPath,
			RsyncCommands: t.RsyncCommands, // This now references the updated structure
		})
	}
	return targets
}

func DisplayFinalResults(results []BackupResult) {
	for _, r := range results {
		log.Printf("TargetName: %s - ElapsedTime: %.3f - TargetSize Before: %.3f, TargetSize After: %.3f - Error: %v", r.TargetName, r.ElapsedTime.Value, r.TargetSize.Before, r.TargetSize.After, r.Err)
	}
}

func (t *Target) getAssociatedIPFromMAC() (string, error) {
	// FIXME: retry if error, sometimes this can happen for some reason
	command := fmt.Sprintf("ip neighbor | grep '%s'", t.MAC)
	out, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		return "", fmt.Errorf("[get associated IP from MAC error] could not grep that mac address: %s\n", err)
	}
	return strings.Split(string(out), " ")[0], nil
}

func (t *Target) isAlive() (bool, error) {
	targetIP, err := t.getAssociatedIPFromMAC()
	if err != nil {
		return false, err
	}
	out, err := exec.Command("ping", targetIP, "-c 2").Output()
	if err != nil {
		return false, fmt.Errorf("[is alive error] could not ping that IP: %s\n", err)
	}

	if strings.Contains(string(out), "Destination Host Unreachable") {
		return false, nil
	} else {
		return true, nil
	}
}

func (t *Target) executeBackup(cfg config.Config, el *utils.ElapsedTime, ts *utils.TargetSize) error {
	var (
		e   error
		err error
	)

	ts.Before, err = utils.GetFolderSize(t.ExternalPath)
	if err != nil {
		log.Printf("[executeBackup error] could not get folder size for %s on the first validation: %s\n", t.Name, err)
	}

	start := time.Now()

	log.Printf("[executeBackup info] starting job: %s\n", t.Name)
	for _, rsyncCommand := range t.RsyncCommands {
		if err := commands.RsyncCommand(rsyncCommand.Command, "toExternal", rsyncCommand.Name, cfg.Pushgateway.Url); err != nil {
			log.Printf("[executeBackup error] could not perform RsyncCommand in %s: %s\n", t.Name, err)
			e = err
		}
	}

	ts.After, err = utils.GetFolderSize(t.ExternalPath)
	if err != nil {
		log.Printf("[executeBackup error] could not get folder size for %s on the second validation: %s\n", t.Name, err)
	}

	end := time.Now()
	el.Target = t.Name
	el.Value = end.Sub(start).Seconds()

	return e
}

func ExecuteTargetsBackups(targets []Target, cfg config.Config) []BackupResult {
	var err error
	results := make([]BackupResult, len(targets)) // Slice to store backup results in order

	for i, target := range targets {
		el := &utils.ElapsedTime{}
		ts := &utils.TargetSize{}
		if target.MAC != "" {
			log.Printf("[execute backups info] target %s contains mac (%s) - checking if it is alive\n", target.Name, target.MAC)
			isAlive, err := target.isAlive()
			if err != nil {
				log.Println(err)
			}
			if !isAlive {
				log.Printf("[execute backups info] target %s is not alive skipping backup\n", target.Name)
				continue
			}
			log.Printf("[execute backups info] target %s is alive\n", target.Name)

		}

		if err = target.executeBackup(cfg, el, ts); err != nil {
			log.Println(err)
		}
		results[i] = BackupResult{TargetName: target.Name, ElapsedTime: *el, TargetSize: *ts, Err: err}
	}

	return results
}

func ValidateBackupResultErrors(backupResults []BackupResult) {
	for _, b := range backupResults {
		if b.Err != nil {
			log.Printf("[errors in backup] found error in %s: %v\n", b.TargetName, b.Err)
		}
	}
}
