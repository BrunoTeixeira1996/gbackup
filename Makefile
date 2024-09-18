FILES = gbackup config.toml email_template.html
REMOTE_USER = brun0
REMOTE_HOST = pinute
REMOTE_PATH = /home/$(REMOTE_USER)/src/gbackup
BINARY_NAME = gbackup
TARGET_OS = linux
TARGET_ARCH = arm64

SENDEREMAIL="1"
SENDERPASS="2"

deploy:
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -o $(BINARY_NAME) .
	rsync -avz --update $(FILES) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)
	ssh $(REMOTE_USER)@$(REMOTE_HOST) 'cd $(REMOTE_PATH) && export SENDEREMAIL=$(SENDEREMAIL) SENDERPASS=$(SENDERPASS) && ./$(BINARY_NAME) -config $(REMOTE_PATH)/config.toml'
