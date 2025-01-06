#!/usr/bin/make -f

BUILD_DIR ?= $(CURDIR)/build
RUBIX_NEXUS_DIR ?= $(CURDIR)
BUILD_NAME = rubix-nexus

# Detect OS for binary naming
ifeq ($(OS),Windows_NT)
    BINARY_NAME = $(BUILD_NAME).exe
else
    BINARY_NAME = $(BUILD_NAME)
endif

.PHONY: build install

build:
	@echo "Building Rubix Nexus..."
	go build -mod=readonly -o $(BUILD_DIR)/$(BINARY_NAME) $(RUBIX_NEXUS_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

install:
	@echo "Installing Rubix Nexus..."
	@echo $(RUBIX_NEXUS_DIR)
	go install -mod=readonly $(RUBIX_NEXUS_DIR)
	@echo "$(BINARY_NAME) installed to $(shell go env GOPATH)/bin"
