SHELL := /bin/bash
FILES = gbackup config.toml email_template.html
REMOTE_USER = brun0
REMOTE_HOST = pinute
REMOTE_PATH = /home/$(REMOTE_USER)/src/gbackup
BINARY_NAME = gbackup
TARGET_OS = linux
TARGET_ARCH = arm64

run:
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -o $(BINARY_NAME) .
	rsync -avz --update $(FILES) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)
	ssh $(REMOTE_USER)@$(REMOTE_HOST) 'source .bash_profile; cd $(REMOTE_PATH) && ./$(BINARY_NAME) -config $(REMOTE_PATH)/config.toml'

deploy:
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -o $(BINARY_NAME) .
	rsync -avz --update $(FILES) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)