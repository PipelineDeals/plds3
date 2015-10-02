## simple makefile to log workflow
.PHONY: all test clean build install

GOFLAGS ?= $(GOFLAGS:)
PLDS3_VERSION=0.1.0
DIST_DIR=dist
EXECUTABLE=plds3
RELEASE_VERSION_PATH=$(DIST_DIR)/plds3-$(PLDS3_VERSION)
LATEST_VERSION_PATH=$(DIST_DIR)/latest

all: install test

build: *.go
	mkdir -p $(DIST_DIR)
	rm -rf $(LATEST_VERSION_PATH)
	mkdir -p $(LATEST_VERSION_PATH)
	GOOS=linux GOARCH=amd64 go build -o $(LATEST_VERSION_PATH)/$(EXECUTABLE)
	mkdir -p $(RELEASE_VERSION_PATH)
	cp $(LATEST_VERSION_PATH)/$(EXECUTABLE) $(RELEASE_VERSION_PATH)/$(EXECUTABLE)
## EOF
