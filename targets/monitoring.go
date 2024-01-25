package targets

import (
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal"
)

/*
   Backups up conf files from:
   - Prometheus
     - /etc/prometheus/rsync.rules.yml
     - /etc/prometheus/cmd.rules.yml
     - /etc/prometheus/prometheus.yml
     - /etc/systemd/system/prometheus.service

   - Grafana
     - /etc/grafana/grafana.ini

   - AlertManager
     - /etc/systemd/system/alertmanager.service
     - /etc/alertmanager/alertmanager.yml

   - Pushgateway
     - /etc/systemd/system/pushgateway.service
*/

// Function that backups all files from prometheus,
// grafana, pushgateway and alertmanager
func backupMonitoringToExternal(cfg internal.Config) error {
	rCmd := []string{"-av", "--delete", "-e", "ssh", "monitoring:/etc/prometheus/rsync.rules.yml", "monitoring:/etc/prometheus/cmd.rules.yml", "monitoring:/etc/prometheus/prometheus.yml", "monitoring:/etc/systemd/system/prometheus.service", "monitoring:/etc/grafana/grafana.ini", "monitoring:/etc/systemd/system/alertmanager.service", "monitoring:/etc/alertmanager/alertmanager.yml", "monitoring:/etc/systemd/system/pushgateway.service", "/mnt/pve/external/monitoring_backup"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "toExternal", cfg.Targets[4].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that copies previous backup from external to
// HDD present in proxmox instance
func backupMonitoringToHDD(cfg internal.Config) error {
	c := []string{"-av", "--delete", "/mnt/pve/external/monitoring_backup", "/storagepool/backups"}

	err := internal.ExecCmdToProm("rsync", c, "toStoragePool", cfg.Targets[4].Instance, cfg.Pushgateway.Host)
	if err != nil {
		return err
	}

	return nil
}

// Function that handles the backup
func ExecuteMonitoringBackup(cfg internal.Config, el *internal.ElapsedTime) error {
	start := time.Now()
	if err := backupMonitoringToExternal(cfg); err != nil {
		return err
	}

	if err := backupMonitoringToHDD(cfg); err != nil {
		return err
	}

	// Calculate run time
	end := time.Now()
	el.Target = cfg.Targets[4].Name
	el.Elapsed = end.Sub(start).Seconds()

	return nil
}
