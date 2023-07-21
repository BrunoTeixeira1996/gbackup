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
  - [] Add cronjobs for functions in main (maybe I could just use the binary in the crontab)
  - [x] Add rule in AlertManager for cronjobs
  - [] Write to log file
  - [x] Create Make file for uploading this to proxmox
  - [] Create total written graph in grafana for rsync and cmd commands
  - []Document (properly) what every rsync and cronjobs do
