# README

## gokr_backup_perm

### Rsync

``` console
rsync -av -e ssh rsync://waiw-backup/waiw/ /mnt/pve/external/gokrazy_backup/waiw_backup/
```

## postgresql_backup

### Rsync

``` console
ssh bot 'pg_dump waiw > waiw.sql && pg_dump leaks > leaks.sql'
rsync -av bot:/root/*.sql /mnt/pve/external/postgresql_backup
```

### Cronjob

``` console
cp /mnt/pve/external/postgresql_backup/*.sql /storagepool/backups/postgresql_backup/
```

## leak_backup

### Cronjob

``` console
DT=`date +%Y%m%d+%H+%M`
tar zcf /mnt/pve/external/leaks_backup/leak-$DT.tar.gz --absolute-names /mnt/pve/external/leaks/ #compress folder
find /mnt/pve/external/leaks_backup/ -name '*.tar.gz' -mtime +15 -exec rm {} \; # remove 15 days old file
cp -r /mnt/pve/external/leaks_backup/leak-$DT.tar.gz /storagepool/backups/leaks_backup # external hard drive -> hdd in proxmox
```

## syncthing_backup

``` console
rsync -Pav -e syncthing:/root/config/Sync /mnt/pve/external/syncthing_backup
```

### Cronjob

``` console
cp -r /mnt/pve/external/syncthing_backup/Sync /storagepool/backups/syncthing_backup
```
