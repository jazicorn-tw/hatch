# syntax=docker/dockerfile:1

# ── Cross-compilation helper ────────────────────────────────────────────────────
# xx (https://github.com/tonistiigi/xx) provides xx-go, xx-apk, and xx-verify.
# xx-go sets GOOS/GOARCH/CC automatically for the target platform so the builder
# stays on --platform=$BUILDPLATFORM (native amd64 on GHA) with no QEMU needed.
FROM --platform=$BUILDPLATFORM tonistiigi/xx@sha256:c64defb9ed5a91eacb37f96ccc3d4cd72521c4bd18d5442905b95e2226b0e707 AS xx

# ── Build ──────────────────────────────────────────────────────────────────────
# CGO_ENABLED=1 is required: mattn/go-sqlite3 and sqlite-vec use CGO.
# -extldflags="-static" produces a fully static binary for distroless/static.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

# Copy xx helpers into the builder.
# COPY --from=xx / / is the canonical pattern — xx is built FROM scratch and
# globbing /usr/local/bin/xx-* fails in BuildKit on scratch-based images.
COPY --from=xx / /

ARG TARGETPLATFORM

# clang + lld: cross-compiler that works for all target architectures.
# xx-apk installs the target-specific sysroot (musl headers, libc).
RUN apk add --no-cache clang lld
RUN xx-apk add --no-cache musl-dev gcc

WORKDIR /build

# Download dependencies as a separate layer — only reruns when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Copy only the Go source directories needed for the build.
# Explicit paths prevent sensitive files (env vars, keys, docs) from
# entering the build context even if .dockerignore is misconfigured.
# -ldflags="-s -w"              strips debug info and DWARF tables (smaller binary).
# -extldflags="-static"         statically links C deps (sqlite3, sqlite-vec).
# -trimpath                     removes local build paths from stack traces.
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=1 xx-go build \
    -ldflags="-s -w -extldflags=-static" -trimpath -o /hatch ./cmd/hatch/ && \
    xx-verify --static /hatch

# ── Runtime ────────────────────────────────────────────────────────────────────
# distroless/static-debian12:nonroot
#   - no shell or package manager (minimal attack surface)
#   - includes CA certificates (required for LLM API calls)
#   - runs as uid 65532 (nonroot) with /home/nonroot home directory
FROM gcr.io/distroless/static-debian12:nonroot

# Explicit USER directive — distroless:nonroot defaults to uid 65532 but
# declaring it here satisfies static analysis tools (e.g. SonarQube S6471).
USER nonroot

COPY --from=builder /hatch /hatch

# SSH server (Milestone 6) and HTTP dashboard (Milestone 8).
EXPOSE 2222 8080

ENTRYPOINT ["/hatch"]
