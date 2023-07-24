# README

gbackup is a go tool to help me backup useful information from various systems I am running.
This could be something private but maybe this will help anyone or give an idea on how to backup stuff usyng go.

The current backup plan is shown below but basicaly I backup the following:
- gokrazy perm partion (resides in gokrazy)
- gokrazy data folder (resides in my personal laptop)
- postgresql databases (resides in a LXC in proxmox)
  - currently I am only doing the backup for the waiw and leak databases
- syncthing folder (resides in a LXC in proxmox)
- monitoring files (resides in a LXC in proxmox)
  - this is files from alertmanager, prometheus, grafana and pushgateway

**It is important to note that I first use `rsync` to backup everything to an external hard drive plugged in my proxmox instance and then I perform a `cp` to the storagepool that is a HDD that resides inside the proxmox and is used only for backups.**



![backup_nobackground](https://github.com/BrunoTeixeira1996/gbackup/assets/12052283/907964c6-ebb4-48be-8654-eb01dcdf130f)
