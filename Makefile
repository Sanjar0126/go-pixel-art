GO := go
BINARY := pixelart
CMD := ./cmd
DIST_DIR := dist
GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)
EXT := $(if $(filter windows,$(GOOS)),.exe,)
OUTPUT := $(DIST_DIR)/$(BINARY)-$(GOOS)-$(GOARCH)$(EXT)

.PHONY: all build build-all build-linux-amd64 build-linux-arm64 \
	build-darwin-amd64 build-darwin-arm64 build-windows-amd64 test clean

all: build

build:
	mkdir -p $(DIST_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o $(OUTPUT) $(CMD)

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 \
	build-darwin-arm64 build-windows-amd64

build-linux-amd64:
	$(MAKE) build GOOS=linux GOARCH=amd64

build-linux-arm64:
	$(MAKE) build GOOS=linux GOARCH=arm64

build-darwin-amd64:
	$(MAKE) build GOOS=darwin GOARCH=amd64

build-darwin-arm64:
	$(MAKE) build GOOS=darwin GOARCH=arm64

build-windows-amd64:
	$(MAKE) build GOOS=windows GOARCH=amd64

test:
	$(GO) test ./...

clean:
	rm -rf $(DIST_DIR)