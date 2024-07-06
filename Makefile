# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

# Main binary name
BINARY_NAME=server
BIN_DIR=bin

all: build

build:
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/server

clean:
	$(GOCLEAN)
	rm -f $(BIN_DIR)/$(BINARY_NAME)

run:
	$(GORUN) ./cmd/server

.PHONY: all build clean run

