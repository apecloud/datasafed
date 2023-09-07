
##@ Linting

.PHONY: lint
lint: staticcheck vet golangci-lint ## Run all lint job against code.

.PHONY: golangci-lint
golangci-lint: golangci ## Run golangci-lint against code.
	$(GOLANGCILINT) run ./...

.PHONY: staticcheck
staticcheck: staticchecktool ## Run staticcheck against code.
	$(STATICCHECK) -tags $(BUILD_TAGS) ./...

.PHONY: vet
vet: ## Run go vet against code.
	GOOS=$(GOOS) $(GO) vet -tags $(BUILD_TAGS) -mod=mod ./...

.PHONY: golangci
golangci: GOLANGCILINT_VERSION = v1.54.2
golangci: ## Download golangci-lint locally if necessary.
ifneq ($(shell which golangci-lint),)
	@echo golangci-lint is already installed
GOLANGCILINT=$(shell which golangci-lint)
else ifeq (, $(shell which $(GOBIN)/golangci-lint))
	@{ \
	set -e ;\
	echo 'installing golangci-lint-$(GOLANGCILINT_VERSION)' ;\
	curl -sSfL $(GITHUB_PROXY)https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCILINT_VERSION) ;\
	echo 'Successfully installed' ;\
	}
GOLANGCILINT=$(GOBIN)/golangci-lint
else
	@echo golangci-lint is already installed
GOLANGCILINT=$(GOBIN)/golangci-lint
endif

.PHONY: staticchecktool
staticchecktool: ## Download staticcheck locally if necessary.
ifeq (, $(shell which staticcheck))
	@{ \
	set -e ;\
	echo 'installing honnef.co/go/tools/cmd/staticcheck' ;\
	go install honnef.co/go/tools/cmd/staticcheck@latest;\
	}
STATICCHECK=$(GOBIN)/staticcheck
else
STATICCHECK=$(shell which staticcheck)
endif
