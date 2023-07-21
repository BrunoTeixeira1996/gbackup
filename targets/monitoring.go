package targets

import "github.com/BrunoTeixeira1996/gbackup/internal"

/*
   Backups up conf files from:
   - Prometheus
     - /etc/prometheus/backups.rules.yml
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
	rCmd := []string{"monitoring:{/etc/prometheus/backups.rules.yml,/etc/prometheus/prometheus.yml,/etc/systemd/system/prometheus.service,/etc/grafana/grafana.ini,/etc/systemd/system/alertmanager.service,/etc/alertmanager/alertmanager.yml,/etc/systemd/system/pushgateway.service}", "/mnt/pve/external/monitoring_backup"}
	if err := internal.ExecCmdToProm("rsync", rCmd, "rsync", cfg.Targets[4].Instance, cfg.Pushgateway.Host); err != nil {
		return err
	}

	return nil
}

// Function that copies previous backup from external to
// HDD present in proxmox instance
func backupMonitoringToHDD(cfg internal.Config) error {
	c := []string{"-r", "/mnt/pve/external/monitoring_backup", "/storagepool/backups"}

	err := internal.ExecCmdToProm("cp", c, "cmd", cfg.Targets[4].Instance, cfg.Pushgateway.Host)
	if err != nil {
		return err
	}

	return nil
}

// Function that handles the backup
func ExecuteMonitoringBackup(cfg internal.Config) error {
	if err := backupMonitoringToExternal(cfg); err != nil {
		return err
	}

	if err := backupMonitoringToHDD(cfg); err != nil {
		return err
	}

	return nil
}
