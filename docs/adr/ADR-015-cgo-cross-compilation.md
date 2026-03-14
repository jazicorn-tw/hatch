<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [adr, build, ci, deploy]
description:  "ADR-015: CGO cross-compilation strategy for Docker multi-platform builds using tonistiigi/xx"
-->
# ADR-015: CGO cross-compilation for Docker multi-platform builds

- **Status:** Accepted
- **Date:** 2026-03-14
- **Deciders:** Project maintainers
- **Scope:** `Dockerfile`, `.github/workflows/` image publish job

---

## Context

M2 replaced `modernc.org/sqlite` (pure Go, no CGO) with `mattn/go-sqlite3` (CGO) to
gain access to the `sqlite-vec` extension for KNN vector search. This broke the Docker
image publish pipeline: the Dockerfile was building with `CGO_ENABLED=0`, which excluded
all Go files in `sqlite-vec-go-bindings/cgo` and produced a build error.

Three additional constraints apply:

1. **Static linking required** — the final image uses `gcr.io/distroless/static-debian12`
   which has no C runtime. CGO-linked binaries must be statically linked
   (`-extldflags="-static"`) or they will crash at startup.

2. **Multi-platform builds** — the publish job builds `linux/amd64` and `linux/arm64`.
   The original Dockerfile used `--platform=$BUILDPLATFORM` on the builder stage to avoid
   QEMU emulation, but CGO cross-compilation requires a target-aware C toolchain — not
   just the host's `gcc`.

3. **musl incompatibility** — an initial Alpine-based builder failed to compile
   `sqlite-vec.c` because it references `u_int8_t`, `u_int16_t`, and `u_int64_t` — BSD
   compatibility types defined by glibc but absent from musl libc.

---

## Decision

Use **[tonistiigi/xx](https://github.com/tonistiigi/xx)** as a cross-compilation helper
in the Dockerfile builder stage, with **`golang:1.26-bookworm`** (Debian) as the builder
base image to avoid musl incompatibilities.

```dockerfile
FROM --platform=$BUILDPLATFORM tonistiigi/xx AS xx
FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS builder

COPY --from=xx / /

ARG TARGETPLATFORM
RUN apt-get update && apt-get install -y --no-install-recommends clang lld \
    && rm -rf /var/lib/apt/lists/*
RUN xx-apt-get install -y --no-install-recommends gcc libsqlite3-dev

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=1 xx-go build \
    -tags "netgo osusergo" \
    -ldflags="-s -w -extldflags=-static" -trimpath -o /hatch ./cmd/hatch/ && \
    xx-verify --static /hatch
```

`xx-go` automatically sets `GOOS`, `GOARCH`, and `CC` for `$TARGETPLATFORM`.
`xx-apt-get` configures Debian multiarch and installs the target-architecture
cross-toolchain and `libsqlite3-dev` headers. The builder stage itself runs natively on
`$BUILDPLATFORM` — no QEMU emulation required.

`-tags "netgo osusergo"` replaces glibc's DNS resolver and user/group lookups with
pure-Go implementations. This is required when statically linking against glibc — without
it, NSS (Name Service Switch) symbols cause runtime failures in a no-libc environment.

`xx-verify --static /hatch` fails the build if the binary is not fully statically
linked, catching linkage errors before the image is pushed.

---

## Alternatives Considered

### 1. `CGO_ENABLED=0` — keep the original approach

Would work only if `sqlite-vec-go-bindings/cgo` were replaced with a pure-Go alternative.
No production-ready pure-Go alternative to sqlite-vec exists today.

**Rejected** — requires replacing a core dependency.

### 2. `CGO_ENABLED=1` without xx, QEMU for arm64

Remove `--platform=$BUILDPLATFORM`, install `gcc musl-dev`, and let the arm64 build run
under QEMU emulation on the amd64 GitHub Actions runner.

**Rejected** — arm64 builds under QEMU take 5–10× longer than native. Acceptable short-
term but poor developer experience at scale.

### 3. Separate `linux/amd64`-only image

Drop `linux/arm64` from the publish matrix.

**Rejected** — arm64 (Apple Silicon, AWS Graviton) is a first-class target. Dropping it
would block deployment on common VPS configurations.

### 4. Alpine builder with `CGO_CFLAGS` workaround

Keep `golang:1.26-alpine` and define the missing BSD types via compiler flags:
`CGO_CFLAGS="-Du_int8_t=uint8_t -Du_int16_t=uint16_t -Du_int64_t=uint64_t"`.

**Rejected** — a preprocessor patch treats the symptom, not the cause. Other C libraries
used by sqlite-vec or future CGO dependencies may surface additional musl gaps. Debian
removes the entire class of musl compatibility issues at the source.

### 5. Replace `distroless/static` with `alpine` runtime image

Use `alpine:latest` as the final stage — libc is present, no need for static linking.

**Rejected** — `distroless/static` has a significantly smaller attack surface (no shell,
no package manager). The security properties are worth the static-linking constraint.

---

## Consequences

### Positive

- Multi-platform builds (`linux/amd64`, `linux/arm64`) remain fast — no QEMU
- `xx-verify --static` provides a hard gate: a non-static binary cannot be pushed
- Builder stage stays on native platform — `go mod download` and compile are fast
- Pattern is reusable for any future CGO dependency

### Negative

- `tonistiigi/xx` is an additional image dependency in the build pipeline — pinning
  to a digest is recommended for reproducibility and supply-chain security
- `golang:1.26-bookworm` is larger than `golang:1.26-alpine`; initial pull and layer
  cache are heavier (mitigated by GHA cache)
- `clang lld` are heavier than `gcc` alone; build layer is slightly larger
- Cross-compilation with CGO is more complex to debug than pure-Go builds

### Follow-up

`tonistiigi/xx` is pinned to digest `sha256:c64defb9ed5a91eacb37f96ccc3d4cd72521c4bd18d5442905b95e2226b0e707`
(latest as of 2026-03-14). Update the digest when upgrading xx.
