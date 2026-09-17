GO := go
BINARY := pixelart
CMD := ./cmd
BUILD_DIR := build
WEB_DIR := web
PALETTE_DIRS := minecraft-icons dota-icons
PALETTE_MANIFESTS := $(addprefix $(WEB_DIR)/,$(addsuffix .json,$(PALETTE_DIRS)))
GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)
EXT := $(if $(filter windows,$(GOOS)),.exe,)
OUTPUT := $(BUILD_DIR)/$(BINARY)-$(GOOS)-$(GOARCH)$(EXT)
BUILD_TARGETS := build-linux-amd64 build-linux-arm64 build-darwin-amd64 \
	build-darwin-arm64 build-windows-amd64

.PHONY: all build build-web web-palettes $(PALETTE_MANIFESTS) build-all build-linux-amd64 build-linux-arm64 \
	build-darwin-amd64 build-darwin-arm64 build-windows-amd64 test clean

all: build

build:
	mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o $(OUTPUT) $(CMD)

build-web: web-palettes
	mkdir -p $(WEB_DIR)
	GOOS=js GOARCH=wasm CGO_ENABLED=0 $(GO) build -o $(WEB_DIR)/app.wasm ./cmd/web
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" $(WEB_DIR)/wasm_exec.js

web-palettes: $(PALETTE_MANIFESTS)

$(WEB_DIR)/minecraft-icons.json:
	@find $(WEB_DIR)/minecraft-icons -maxdepth 1 -type f -name '*.png' -print 2>/dev/null | sort | sed 's#^$(WEB_DIR)/##' | awk 'BEGIN { printf "[" } { if (NR > 1) printf ","; printf "%c%s%c", 34, $$0, 34 } END { print "]" }' > $@

$(WEB_DIR)/dota-icons.json:
	@find $(WEB_DIR)/dota-icons -maxdepth 1 -type f -name '*.png' -print 2>/dev/null | sort | sed 's#^$(WEB_DIR)/##' | awk 'BEGIN { printf "[" } { if (NR > 1) printf ","; printf "%c%s%c", 34, $$0, 34 } END { print "]" }' > $@

build-all:
	@failed=0; \
	for target in $(BUILD_TARGETS); do \
		echo "==> $$target"; \
		if ! $(MAKE) $$target; then \
			echo "WARNING: $$target failed; continuing"; \
			failed=1; \
		fi; \
	done; \
	if [ $$failed -ne 0 ]; then \
		echo "Some builds failed."; \
		exit 1; \
	fi

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
	rm -rf $(BUILD_DIR) $(WEB_DIR)/app.wasm $(WEB_DIR)/wasm_exec.js $(PALETTE_MANIFESTS)