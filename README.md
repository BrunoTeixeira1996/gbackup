gbackup is a go tool to help me backup useful information from various systems I am running.
This could be something private but maybe this will help someone or give an idea on how to backup stuff using go.

The current backup plan is shown below but I backup the following:
- gokrazy perm partion (resides in gokrazy)
- gokrazy data folder (resides in a LXC in proxmox)
- postgresql databases (resides in a LXC in proxmox)
  - currently I am only doing the backup for the waiw and leak databases
- syncthing folder (resides in a LXC in proxmox)
- monitoring files (resides in a LXC in proxmox)
  - this is files from alertmanager, prometheus, grafana and pushgateway
- work laptop

At first I was using `rsync` to backup everything to an external hard drive plugged in my proxmox and then use `cp` to backup everything to a different location (storagepool - HDD that resides inside the proxmox). However now I am using `rsync` in both backups.

![image](https://github.com/BrunoTeixeira1996/gbackup/assets/12052283/82b2a47d-4998-410d-b9ff-583c338d846f)

I was using crontab inside proxmox but now I've managed to create a similiar cron using golang so its easier to apply this to any unix based OS without caring if crontab is installed or no.

Then I monitor the `rsync` commands using Prometheus as shown bellow.

![image](https://github.com/BrunoTeixeira1996/gbackup/assets/12052283/953d3c75-29ce-401d-9973-3f773533e664)
