package targets

import "github.com/BrunoTeixeira1996/gbackup/internal"

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
	locations := [][]string{
		{"-av", "--delete", "-e", "ssh", "monitoring:/etc/prometheus/rsync.rules.yml", "/mnt/pve/external/monitoring_backup"},
		{"-av", "--delete", "-e", "ssh", "monitoring:/etc/prometheus/cmd.rules.yml", "/mnt/pve/external/monitoring_backup"},
		{"-av", "--delete", "-e", "ssh", "monitoring:/etc/prometheus/prometheus.yml", "/mnt/pve/external/monitoring_backup"},
		{"-av", "--delete", "-e", "ssh", "monitoring:/etc/systemd/system/prometheus.service", "/mnt/pve/external/monitoring_backup"},
		{"-av", "--delete", "-e", "ssh", "monitoring:/etc/grafana/grafana.ini", "/mnt/pve/external/monitoring_backup"},
		{"-av", "--delete", "-e", "ssh", "monitoring:/etc/systemd/system/alertmanager.service", "/mnt/pve/external/monitoring_backup"},
		{"-av", "--delete", "-e", "ssh", "monitoring:/etc/alertmanager/alertmanager.yml", "/mnt/pve/external/monitoring_backup"},
		{"-av", "--delete", "-e", "ssh", "monitoring:/etc/systemd/system/pushgateway.service", "/mnt/pve/external/monitoring_backup"},
	}

	for _, v := range locations {
		// FIXME: This is a workaround for the issue https://github.com/stapelberg/rsyncprom/issues/1
		//rCmd := []string{"-av", "--delete", "-e", "ssh", "monitoring:{/etc/prometheus/rsync.rules.yml,/etc/prometheus/cmd.rules.yml,/etc/prometheus/prometheus.yml,/etc/systemd/system/prometheus.service,/etc/grafana/grafana.ini,/etc/systemd/system/alertmanager.service,/etc/alertmanager/alertmanager.yml,/etc/systemd/system/pushgateway.service}", "/mnt/pve/external/monitoring_backup"}
		if err := internal.ExecCmdToProm("rsync", v, "rsync", cfg.Targets[4].Instance, cfg.Pushgateway.Host); err != nil {
			return err
		}
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
