OS ?= $(shell uname)
ARCH ?= $(shell uname -m)

GOOS ?= $(shell echo "$(OS)" | tr '[:upper:]' '[:lower:]')
GOARCH_x86_64 = amd64
GOARCH_aarch64 = arm64
GOARCH_arm64 = arm64
GOARCH ?= $(shell echo "$(GOARCH_$(ARCH))")

VERSION := $(shell git describe --tags --always)
REVISION := $(shell git rev-parse HEAD)
PACKAGE := github.com/day0ops/netctl/config
VERSION_VARIABLES := -X $(PACKAGE).appVersion=$(VERSION) -X $(PACKAGE).revision=$(REVISION)

OUTPUT_DIR := _output/binaries
OUTPUT_BIN := netctl-$(GOOS)-$(ARCH)
BIN_NAME := netctl

LDFLAGS := $(VERSION_VARIABLES)

.PHONY: all
all: build

.PHONY: clean
clean:
	rm -rf _output _build

.PHONY: fmt
fmt:
	go fmt ./...
	goimports -w .

.PHONY: build
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(OUTPUT_BIN) ./cmd
ifeq ($(GOOS),darwin)
	codesign -s - $(OUTPUT_DIR)/$(OUTPUT_BIN)
endif
	cd $(OUTPUT_DIR) && openssl sha256 -r -out $(OUTPUT_BIN).sha256sum $(OUTPUT_BIN)

