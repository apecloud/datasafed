# Build the manager binary
ARG DIST_IMG=gcr.io/distroless/static:nonroot

ARG GO_VERSION=1.21

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION} as builder

## docker buildx build injected build-args:
#BUILDPLATFORM — matches the current machine. (e.g. linux/amd64)
#BUILDOS — os component of BUILDPLATFORM, e.g. linux
#BUILDARCH — e.g. amd64, arm64, riscv64
#BUILDVARIANT — used to set ARM variant, e.g. v7
#TARGETPLATFORM — The value set with --platform flag on build
#TARGETOS - OS component from --platform, e.g. linux
#TARGETARCH - Architecture from --platform, e.g. arm64
#TARGETVARIANT

ARG TARGETOS
ARG TARGETARCH

ARG GOPROXY
#ARG GOPROXY=https://goproxy.cn
ARG LD_FLAGS="-s -w"

ENV GOPROXY=${GOPROXY}

WORKDIR /src

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go env && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} BUILD_DIR=/out make datasafed

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ${DIST_IMG} as dist

WORKDIR /
COPY --from=builder /out/datasafed .
USER 65532:65532

ENTRYPOINT ["/datasafed"]
