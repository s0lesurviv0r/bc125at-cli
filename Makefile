BIN     := bc125at
VERSION := $(shell grep 'version = ' cmd/root.go | awk -F'"' '{print $$2}')
BUILD   := build
DIST    := build/dist

.PHONY: all build test clean release

all: build

build:
	mkdir -p $(BUILD)
	go build -ldflags="-s -w" -o $(BUILD)/$(BIN) .

test:
	go test ./...

clean:
	rm -rf $(BUILD) $(DIST)

release: clean
	mkdir -p $(DIST)
	GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o $(DIST)/$(BIN)-$(VERSION)-linux-amd64   .
	GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w" -o $(DIST)/$(BIN)-$(VERSION)-linux-arm64   .
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o $(DIST)/$(BIN)-$(VERSION)-darwin-amd64  .
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o $(DIST)/$(BIN)-$(VERSION)-darwin-arm64  .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(DIST)/$(BIN)-$(VERSION)-windows-amd64.exe .
	@echo "Binaries written to $(DIST)/"
