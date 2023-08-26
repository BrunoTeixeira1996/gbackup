FILES = gbackup config.toml email_template.html

deploy:
	CGO_ENABLED=0 go build .
	scp $(FILES) proxmox:/root/gbackup/
