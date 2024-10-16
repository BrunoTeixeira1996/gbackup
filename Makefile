SHELL := /bin/bash
FILES = gbackup config.toml temp_config.toml internal/email/email.html
REMOTE_USER = brun0
REMOTE_HOST = pinute
REMOTE_PATH = /home/$(REMOTE_USER)/src/gbackup
BINARY_NAME = gbackup
TARGET_OS = linux
TARGET_ARCH = arm64

compile:
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -o $(BINARY_NAME) ./cmd/gbackup

debug:
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -o $(BINARY_NAME) ./cmd/gbackup
	rsync -avz --update $(FILES) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)
	ssh $(REMOTE_USER)@$(REMOTE_HOST) 'source .bash_profile; cd $(REMOTE_PATH) && ./$(BINARY_NAME) -config $(REMOTE_PATH)/temp_config.toml'

run:
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -o $(BINARY_NAME) ./cmd/gbackup
	rsync -avz --update $(FILES) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)
	ssh $(REMOTE_USER)@$(REMOTE_HOST) 'source .bash_profile; cd $(REMOTE_PATH) && ./$(BINARY_NAME) -config $(REMOTE_PATH)/config.toml'

deploy:
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -o $(BINARY_NAME) ./cmd/gbackup
	rsync -avz --update $(FILES) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)
