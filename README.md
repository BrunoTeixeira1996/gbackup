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

**It is important to note that I first use `rsync` to backup everything to an external hard drive plugged in my proxmox instance and then I perform a `cp` to the storagepool that is a HDD that resides inside the proxmox and is used only for backups.**


![image](https://github.com/BrunoTeixeira1996/gbackup/assets/12052283/ddc4a4d6-813a-4eab-a70a-2e4a460aad3e)

Inside the proxmox instance I run the following cronjob
- Note that I do have the necessary env vars so cronjob knows whats the email and password for sending emails

``` bash
0 17 * * FRI /root/gbackup/gbackup > /root/gbackup/logstdout 2> /root/gbackup/logstderr
```


Then I monitor the `rsync` and `cp` commands using Prometheus as shown bellow.

**This is a simple dashboard as I am new to Grafana but the main thing I use is the AlertManager in case something is wrong**

![image](https://github.com/BrunoTeixeira1996/gbackup/assets/12052283/210c3976-f776-42a8-a215-f691ff21af45)
