##############################################################################
# Variables
##############################################################################
APP_NAME = datasafed
VERSION ?= 0.1.0
GITHUB_PROXY ?=
GIT_COMMIT  = $(shell git rev-list -1 HEAD)
GIT_VERSION = $(shell git describe --always --abbrev=0 --tag)

TAG_LATEST ?= false
BUILDX_ENABLED ?= false
BUILD_DIR ?= build

# Go setup
export GO111MODULE = auto
export GOSUMDB = sum.golang.org
GO ?= go
GOFMT ?= gofmt
GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell $(GO) env GOBIN))
GOBIN=$(shell $(GO) env GOPATH)/bin
else
GOBIN=$(shell $(GO) env GOBIN)
endif
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://proxy.golang.org
## use following GOPROXY settings for Chinese mainland developers.
#GOPROXY := https://goproxy.cn
endif
export GOPROXY

BUILD_TAGS ?= ""
LD_FLAGS = "-s -w \
	-X github.com/apecloud/datasafed/version.BuildDate=`date -u +'%Y-%m-%dT%H:%M:%SZ'` \
	-X github.com/apecloud/datasafed/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/apecloud/datasafed/version.GitVersion=$(GIT_VERSION) \
	-X github.com/apecloud/datasafed/version.Version=$(VERSION)"

##############################################################################
# Targets
##############################################################################
.DEFAULT_GOAL := help
.PHONY: default
default: help

.PHONY: build
build: datasafed ## Build binaries.

.PHONY: datasafed
datasafed: ## Build datasafed.
	mkdir -p $(BUILD_DIR)
	$(GO) build -v -o $(BUILD_DIR)/datasafed -tags $(BUILD_TAGS) -ldflags $(LD_FLAGS) .

.PHONY: docgen
docgen: ## Generate documents.
	$(GO) generate ./hack/docgen/...

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php
# https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##############################################################################
# Includes
##############################################################################
include docker.mk
include lint.mk
