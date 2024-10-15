# TODO

- Create rsyncs and cronjobs
  - [x] Create postgresql (needs testing in proxmox)
  - [x] Create gokr_backup_perm (needs testing in proxmox)
  - [x] Create syncthing_backup (needs testing in proxmox)
  - [x] Create backup for gokrazy config files that are present in my laptop
	- my laptop has a running cronjob that sends the gokrazy folder
	- and then proxmox backups that to storagepool
  - [x] Create leak_backup (needs testing in proxmox)
	- Fix leak_backup cronjob architecture
  - [x] Create backup for grafana and prometheus conf files
	- proxmox will download conf files from prometheus and grafana lxc
	- then it will save in the external hard drive and then copy to storagepool
  - [x] Add rule in AlertManager for cronjobs
  - [x] Setup crontab in proxmox
  - [x] Create Make file for uploading this to proxmox
  - [x] Implement goroutines for a faster backup
	- Every block of backup can run in a goroutine since the data that is shared is from rsync and cmd
	- [x] There's a bug in the log since it prints randomly
  - [x] Setup Email alert when script finishes (this is easier than seting up AlertManager to send alert when event is not failing)
  - [x] Format output as html template so its better to read
  - [x] Add more stuff when sending Email
	- Timestamp
	- Log
	- Number of backups and number of well executed backups
  - [x] Add temp control for the raspberry pi (node_hwmon_temp_celsius{instance="brun0-pi:9100"}) - create alert for temp high than 60
  - [x] Document (properly) what every rsync and cronjobs do

## v2

- New version of gbackup for a different setup

- [x] Setup new default branch
- [x] Check if `/mnt/external` is mounted
- [x] `external` backup should be the last backup since it will contain more data (i.e worklatop, gokrazy_perm ...)
- [x] Clean previous code
- [x] Implement logic to keep 2 backups and delete the older one for the external -> NAS since PBS does the same
- [x] Clean return code
  - Caller function only return fmt.Errorf if necessary, otherwise it only return err and the function that is being called returns fmt.Errorf
- [x] Fix Prometheus metrics
- [x] Find a way of validating if PBS finished the backup, only shutdown NAS when this is finished
  - https://forum.proxmox.com/threads/query-if-a-backup-task-is-running.112019/
  - https://www.reddit.com/r/Proxmox/comments/1cqi54w/backup_status_via_api/
- [x] Buffer the output from the backup go routine in order to have the correct output from the targets
- [x] Implement WoL (Wake on Lan)
- [x] Implement the shutdown of the NAS
- [x] Create new struct for targets and simplify main.go using methods
- [ ] Calculate full backup time
- [ ] Check if its end of month, if yes compress the data and send to new external HDD plugged in Pinute (Pi)
- [ ] Modify email template to be similar to PBS email
  - This is be the best approach as the first rsync will take sometime but then it will be fast since we are doing incremental backups
  - Also get the apt and pipx packages
  - Test this with a VM by having a backup of my work laptop and rsync that backup to a fresh clean VM
