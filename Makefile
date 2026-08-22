# Define variables
CMD_DIR=./cmd/fldrecall

# Define standard installation paths
PREFIX ?= /usr/local
BINDIR  = $(PREFIX)/bin

# Define target binary file name
CPUARCH := $(shell go env GOARCH)
OSTYPE  := $(shell go env GOHOSTOS)
TARGET  := fldrecall

.PHONY: all build run test clean tidy

all: build

# build: Compile the binary
build:
	@./ci/linux/build.sh

# run: Build and execute the application
run: build
	@echo
	@./bin/$(TARGET) || true

# test: Run unit tests across all packages
test:
	@./ci/linux/test.sh

# clean: Remove compiled binaries
clean:
	@./ci/linux/clean.sh

# tidy: Format code and clean up dependencies
tidy:
	go fmt ./...
	go mod tidy

# install: Install the binary to the user's bin directory
install: build
	@echo
	mkdir -p $(DESTDIR)$(BINDIR)
	install -m 0755 $(TARGET) $(DESTDIR)$(BINDIR)/$(TARGET)
