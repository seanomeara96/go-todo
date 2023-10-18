# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

# Main binary name
BINARY_NAME=main
BIN_DIR=bin

all: build

build:
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME)/ cmd/main/main.go

clean:
	$(GOCLEAN)
	rm -f $(BIN_DIR)/$(BINARY_NAME)

run:
	$(GORUN) cmd/main/main.go

.PHONY: all build clean run

