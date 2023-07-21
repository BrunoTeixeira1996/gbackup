FILES = gbackup config.toml

deploy:
	CGO_ENABLED=0 go build .
	scp $(FILES) proxmox:/root/gbackup/
