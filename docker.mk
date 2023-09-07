
# To use buildx: https://github.com/docker/buildx#docker-ce
export DOCKER_CLI_EXPERIMENTAL=enabled

# Docker image build and push setting
DOCKER:=DOCKER_BUILDKIT=1 docker

# BUILDX_PLATFORMS ?= $(subst -,/,$(ARCH))
BUILDX_PLATFORMS ?= linux/amd64,linux/arm64

BUILDX_BUILDER ?= "$(APP_NAME)-builder"

# Image URL to use all building/pushing image targets
IMG ?= docker.io/apecloud/$(APP_NAME)

IMAGE_TAG ?=
ifeq ($(IMAGE_TAG),)
  ifeq ($(TAG_LATEST), true)
    IMAGE_TAG = latest
  else
    IMAGE_TAG = $(VERSION)
  endif
endif

DOCKERFILE_DIR = .
GO_BUILD_ARGS ?= --build-arg GITHUB_PROXY=$(GITHUB_PROXY) --build-arg GOPROXY=$(GOPROXY)
BUILD_ARGS ?=
DOCKER_BUILD_ARGS ?=
DOCKER_BUILD_ARGS += $(GO_BUILD_ARGS) $(BUILD_ARGS)

##@ Docker containers

.PHONY: build-docker-image
build-docker-image: install-docker-buildx ## Build container image.
ifneq ($(BUILDX_ENABLED), true)
	$(DOCKER) build . $(DOCKER_BUILD_ARGS) --file $(DOCKERFILE_DIR)/Dockerfile --tag $(IMG):$(IMAGE_TAG)
else
	$(DOCKER) buildx build . $(DOCKER_BUILD_ARGS) --file $(DOCKERFILE_DIR)/Dockerfile --platform $(BUILDX_PLATFORMS) --tag $(IMG):$(IMAGE_TAG)
endif


.PHONY: push-docker-image
push-docker-image: install-docker-buildx ## Push container image.
ifneq ($(BUILDX_ENABLED), true)
	$(DOCKER) push $(IMG):$(IMAGE_TAG)
else
	$(DOCKER) buildx build . $(DOCKER_BUILD_ARGS) --file $(DOCKERFILE_DIR)/Dockerfile --platform $(BUILDX_PLATFORMS) --tag $(IMG):$(IMAGE_TAG) --push
endif


.PHONY: install-docker-buildx
install-docker-buildx: ## Create `docker buildx` builder.
	@if ! docker buildx inspect $(BUILDX_BUILDER) > /dev/null; then \
		echo "Buildx builder $(BUILDX_BUILDER) does not exist, creating..."; \
		docker buildx create --name=$(BUILDX_BUILDER) --use --driver=docker-container --platform linux/amd64,linux/arm64; \
	else \
		echo "Buildx builder $(BUILDX_BUILDER) already exists"; \
	fi
