# TODO

- Create rsyncs and cronjobs
  - [x] Create postgresql
  - [x] Create gokr_backup_perm
  - [x] Create syncthing_backup
  - [x] Create backup for gokrazy config files that are present in my laptop
	- my laptop has a running cronjob that sends the gokrazy folder
	- and then proxmox backups that to storagepool
  - [] Create leak_backup
	- Fix leak_backup cronjob architecture
  - [] Create backup for grafana and prometheus conf files

- Fix paths to correct ones since right now I am using debug paths
- Add Alert Manager when all backups are created
- Document (properly) what every rsync and cronjobs do
