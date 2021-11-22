# ------------------------------------------------------------
# Copyright (c) Microsoft Corporation and TKEEL Contributors.
# Licensed under the MIT License.
# ------------------------------------------------------------

################################################################################
# Variables                                                                    #
################################################################################

export GO111MODULE ?= on
export GOPROXY ?= https://proxy.golang.org
export GOSUMDB ?= sum.golang.org

GOPATH:=$(shell go env GOPATH)

# Proto files.
INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
API_PROTO_FILES=$(shell find api -name *.proto)

GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_BRANCH=$(shell git name-rev --name-only HEAD)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty)
# By default, disable CGO_ENABLED. See the details on https://golang.org/cmd/cgo
CGO         ?= 0
BINARIES    ?= rudder keel
HA_MODE     ?= false

# Add latest tag if LATEST_RELEASE is true
LATEST_RELEASE ?=

PROTOC ?=protoc

ifdef REL_VERSION
	TKEEL_VERSION := $(REL_VERSION)
else
	TKEEL_VERSION := edge
endif

LOCAL_ARCH := $(shell uname -m)
ifeq ($(LOCAL_ARCH),x86_64)
	TARGET_ARCH_LOCAL=amd64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 5),armv8)
	TARGET_ARCH_LOCAL=arm64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 4),armv)
	TARGET_ARCH_LOCAL=arm
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 5),arm64)
	TARGET_ARCH_LOCAL=arm64
else
	TARGET_ARCH_LOCAL=amd64
endif
export GOARCH ?= $(TARGET_ARCH_LOCAL)

ifeq ($(GOARCH),amd64)
	LATEST_TAG=latest
else
	LATEST_TAG=latest-$(GOARCH)
endif

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
   TARGET_OS_LOCAL = linux
else ifeq ($(LOCAL_OS),Darwin)
   TARGET_OS_LOCAL = darwin
else
   TARGET_OS_LOCAL ?= windows
endif
export GOOS ?= $(TARGET_OS_LOCAL)

# Default docker container and e2e test targst.
TARGET_OS ?= linux
TARGET_ARCH ?= amd64
TEST_OUTPUT_FILE_PREFIX ?= ./test_report

ifeq ($(GOOS),windows)
BINARY_EXT_LOCAL:=.exe
GOLANGCI_LINT:=golangci-lint.exe
export ARCHIVE_EXT = .zip
else
BINARY_EXT_LOCAL:=
GOLANGCI_LINT:=golangci-lint
export ARCHIVE_EXT = .tar.gz
endif

export BINARY_EXT ?= $(BINARY_EXT_LOCAL)

OUT_DIR := ./dist


################################################################################
# Go build details                                                             #
################################################################################
BASE_PACKAGE_NAME := github.com/tkeel-io/tkeel

DEFAULT_LDFLAGS:=-X $(BASE_PACKAGE_NAME)/pkg/version.GitCommit=$(GIT_COMMIT) \
  -X $(BASE_PACKAGE_NAME)/pkg/version.Version=$(GIT_VERSION) \
  -X $(BASE_PACKAGE_NAME)/pkg/version.GitVersion=$(GIT_VERSION) \
  -X $(BASE_PACKAGE_NAME)/pkg/version.BuildDate=$(BUILD_DATE) \
  -X $(BASE_PACKAGE_NAME)/pkg/version.Version=$(TKEEL_VERSION)

ifeq ($(origin DEBUG), undefined)
  BUILDTYPE_DIR:=release
  LDFLAGS:="$(DEFAULT_LDFLAGS) -s -w"
else ifeq ($(DEBUG),0)
  BUILDTYPE_DIR:=release
  LDFLAGS:="$(DEFAULT_LDFLAGS) -s -w"
else
  BUILDTYPE_DIR:=debug
  GCFLAGS:=-gcflags="all=-N -l"
  LDFLAGS:="$(DEFAULT_LDFLAGS)"
  $(info Build with debugger information)
endif

TKEEL_OUT_DIR := $(OUT_DIR)/$(GOOS)_$(GOARCH)/$(BUILDTYPE_DIR)
TKEEL_LINUX_OUT_DIR := $(OUT_DIR)/linux_$(GOARCH)/$(BUILDTYPE_DIR)

################################################################################
# Target: build                                                                #
################################################################################
.PHONY: build
TKEEL_BINS:=$(foreach ITEM,$(BINARIES),$(TKEEL_OUT_DIR)/$(ITEM)$(BINARY_EXT))
build: $(TKEEL_BINS)

# Generate builds for TKEEL binaries for the target
# Params:
# $(1): the binary name for the target
# $(2): the binary main directory
# $(3): the target os
# $(4): the target arch
# $(5): the output directory
define genBinariesForTarget
.PHONY: $(5)/$(1)
$(5)/$(1):
	CGO_ENABLED=$(CGO) GOOS=$(3) GOARCH=$(4) go build $(GCFLAGS) -ldflags=$(LDFLAGS) \
	-o $(5)/$(1) $(2)/;
endef

# Generate binary targets
$(foreach ITEM,$(BINARIES),$(eval $(call genBinariesForTarget,$(ITEM)$(BINARY_EXT),./cmd/$(ITEM),$(GOOS),$(GOARCH),$(TKEEL_OUT_DIR))))

################################################################################
# Target: build-linux                                                          #
################################################################################
BUILD_LINUX_BINS:=$(foreach ITEM,$(BINARIES),$(TKEEL_LINUX_OUT_DIR)/$(ITEM))
build-linux: $(BUILD_LINUX_BINS)

# Generate linux binaries targets to build linux docker image
ifneq ($(GOOS), linux)
$(foreach ITEM,$(BINARIES),$(eval $(call genBinariesForTarget,$(ITEM),./cmd/$(ITEM),linux,$(GOARCH),$(TKEEL_LINUX_OUT_DIR))))
endif

################################################################################
# Target: archive                                                              #
################################################################################
ARCHIVE_OUT_DIR ?= $(TKEEL_OUT_DIR)
ARCHIVE_FILE_EXTS:=$(foreach ITEM,$(BINARIES),archive-$(ITEM)$(ARCHIVE_EXT))

archive: $(ARCHIVE_FILE_EXTS)

# Generate archive files for each binary
# $(1): the binary name to be archived
# $(2): the archived file output directory
define genArchiveBinary
ifeq ($(GOOS),windows)
archive-$(1).zip:
	7z.exe a -tzip "$(2)\\$(1)_$(GOOS)_$(GOARCH)$(ARCHIVE_EXT)" "$(TKEEL_OUT_DIR)\\$(1)$(BINARY_EXT)"
else
archive-$(1).tar.gz:
	tar czf "$(2)/$(1)_$(GOOS)_$(GOARCH)$(ARCHIVE_EXT)" -C "$(TKEEL_OUT_DIR)" "$(1)$(BINARY_EXT)"
endif
endef

# Generate archive-*.[zip|tar.gz] targets
$(foreach ITEM,$(BINARIES),$(eval $(call genArchiveBinary,$(ITEM),$(ARCHIVE_OUT_DIR))))

################################################################################
# Target: archive                                                              #
################################################################################
release: build archive

################################################################################
# Target: test                                                                 #
################################################################################
.PHONY: test
test: test-deps
	gotestsum --jsonfile $(TEST_OUTPUT_FILE_PREFIX)_unit.json --format standard-quiet -- -gcflags=-l ./pkg/... ./utils/... ./cmd/... $(COVERAGE_OPTS)
	go test ./tests/...

################################################################################
# Target: lint                                                                 #
################################################################################
# Due to https://github.com/golangci/golangci-lint/issues/580, we need to add --fix for windows
.PHONY: lint
lint:
	$(GOLANGCI_LINT) run --timeout=20m

################################################################################
# Target: modtidy                                                              #
################################################################################
.PHONY: modtidy
modtidy:
	go mod tidy

################################################################################
# Target: check-diff                                                           #
################################################################################
.PHONY: check-diff
check-diff:
	git diff --exit-code ./go.mod # check no changes
	git diff --exit-code ./go.sum # check no changes

################################################################################
# Target: init-proto                                                           #
################################################################################
.PHONY: init-proto
init-proto:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go get -u github.com/tkeel-io/kit/cmd/protoc-gen-go-http
	go get -u github.com/tkeel-io/tkeel-interface/protoc-gen-go-errors
	go get -u github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go install github.com/tkeel-io/kit/cmd/protoc-gen-go-http
	go install github.com/tkeel-io/tkeel-interface/protoc-gen-go-errors
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
################################################################################
# Target: gen-api-proto                                                        #
################################################################################
.PHONY: gen-api-proto
gen-api-proto:
	protoc --proto_path=. \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:. \
 	       --go-http_out=paths=source_relative:. \
 	       --go-grpc_out=paths=source_relative:. \
		   --go-errors_out=paths=source_relative:. \
 	       --openapiv2_out . \
 	       --openapiv2_opt logtostderr=true \
 	       --openapiv2_opt json_names_for_fields=false \
	       $(API_PROTO_FILES)