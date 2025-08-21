BINARY=bin/version-watcher-bot
LDFLAGS := -s -w 

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd