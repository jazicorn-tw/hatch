# syntax=docker/dockerfile:1

# ── Build ──────────────────────────────────────────────────────────────────────
# Build on the host platform, cross-compile for the target.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /build

# Download dependencies as a separate layer — only reruns when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build.
# modernc.org/sqlite is pure Go — CGO_ENABLED=0 works without a C toolchain.
# -ldflags="-s -w"  strips debug info and DWARF tables (smaller binary).
# -trimpath          removes local build paths from stack traces.
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -trimpath -o /hatch ./cmd/hatch/

# ── Runtime ────────────────────────────────────────────────────────────────────
# distroless/static-debian12:nonroot
#   - no shell or package manager (minimal attack surface)
#   - includes CA certificates (required for LLM API calls)
#   - runs as uid 65532 (nonroot) with /home/nonroot home directory
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /hatch /hatch

# SSH server (Milestone 6) and HTTP dashboard (Milestone 8).
EXPOSE 2222 8080

ENTRYPOINT ["/hatch"]
