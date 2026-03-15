<!--
created_by:   jazicorn-tw
created_date: 2026-03-15
updated_by:   jazicorn-tw
updated_date: 2026-03-15
status:       active
tags:         [onboarding, go, build, docker, ci]
description:  "How hatch cross-compiles a CGO binary for linux/amd64 and linux/arm64 using tonistiigi/xx"
-->
# Cross-Compilation — Building for Multiple Platforms

Hatch ships as a single Docker image that runs on both `linux/amd64` (most cloud VMs)
and `linux/arm64` (Apple Silicon, AWS Graviton). Building a binary that targets a
different CPU architecture than the machine doing the building is called
**cross-compilation**.

Pure Go handles this trivially. CGO makes it hard. Hatch needs CGO (see [`CGO.md`](CGO.md)).
This document explains the problem and how hatch solves it.

---

## The three-way constraint

Hatch's Docker build has to satisfy three constraints simultaneously:

| Constraint                | Why                                                                                  |
| ------------------------- | ------------------------------------------------------------------------------------ |
| `CGO_ENABLED=1`           | `sqlite-vec-go-bindings/cgo` is a C extension; pure Go cannot compile it             |
| Static binary             | Final image is `distroless/static` — no C runtime, so the binary must carry its own  |
| `linux/amd64` + `arm64`   | Both architectures are first-class targets                                           |

Each constraint alone is straightforward. Together they create a problem:

- CGO requires a C compiler
- Cross-compiling C requires a **target-aware** C compiler (one that emits arm64 code when
  running on an amd64 host)
- Most base images don't ship cross-compilers by default

---

## Why pure-Go cross-compilation doesn't apply here

With no CGO, this works:

```bash
# Build a Linux arm64 binary from any machine
GOOS=linux GOARCH=arm64 go build ./cmd/hatch
```

Go's compiler handles the translation internally — no C toolchain needed.

As soon as `CGO_ENABLED=1`, Go delegates the C compilation to the host's `gcc` or
`clang`. Those compilers default to the **host** architecture. Running `gcc` on an amd64
machine produces amd64 code — not arm64.

To cross-compile C code for arm64, you need an arm64-targeting cross-compiler:
`aarch64-linux-gnu-gcc` (glibc) or `aarch64-linux-musl-gcc` (musl). Setting this up
manually per-platform is fragile and slow.

---

## The QEMU alternative (and why it's rejected)

One workaround: let Docker run the arm64 build stage **under QEMU emulation** on the
amd64 runner. The arm64 build sees itself as native and uses its own `gcc`.

```dockerfile
# Without --platform=$BUILDPLATFORM — arm64 stage runs under QEMU
FROM golang:1.26-bookworm AS builder
```

This works but is slow. QEMU-emulated builds run 5–10× slower than native because every
instruction is translated at runtime. On GitHub Actions, an arm64 CGO build under QEMU
can take 15–20 minutes vs 2–3 minutes natively.

**Hatch rejects QEMU** for this reason. See [ADR-015](../../../adr/ADR-015-cgo-cross-compilation.md).

---

## The solution: tonistiigi/xx

[`tonistiigi/xx`](https://github.com/tonistiigi/xx) is a small Docker image that
provides wrapper scripts (`xx-go`, `xx-apt-get`, `xx-verify`) that make CGO
cross-compilation straightforward.

The key insight: the **builder stage runs natively on the host platform** (`$BUILDPLATFORM`
= `linux/amd64` on GitHub Actions). `xx` handles setting the right compiler flags,
sysroot, and target headers for the **target platform** (`$TARGETPLATFORM` = `linux/arm64`
or `linux/amd64`).

```text
GitHub Actions runner (amd64)
│
├── builder stage runs natively on amd64      ← fast
│   ├── xx-go sets GOARCH=arm64, CC=aarch64-linux-gnu-gcc
│   ├── xx-apt-get installs arm64 sysroot + headers
│   └── compiles arm64 binary natively (no QEMU)
│
└── final stage: distroless/static (arm64)
    └── copies arm64 hatch binary in
```

---

## The Dockerfile explained

```dockerfile
# Stage 1: pull the xx helper scripts
FROM --platform=$BUILDPLATFORM tonistiigi/xx AS xx

# Stage 2: builder — runs natively on the host (amd64 on GHA)
FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS builder

# Copy xx scripts into the builder's PATH
COPY --from=xx / /

# TARGETPLATFORM tells xx which architecture to target (amd64 or arm64)
ARG TARGETPLATFORM

# Install host-side C toolchain (clang + lld run on amd64)
RUN apt-get update && apt-get install -y --no-install-recommends clang lld \
    && rm -rf /var/lib/apt/lists/*

# xx-apt-get installs the TARGET architecture's headers and cross-compiler
# For arm64: installs aarch64-linux-gnu-gcc + libsqlite3-dev for arm64
RUN xx-apt-get install -y --no-install-recommends gcc libsqlite3-dev

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

# xx-go wraps 'go build' — sets GOARCH, CC, and CGO flags for $TARGETPLATFORM
# -tags "netgo osusergo": use pure-Go DNS + user lookup (required for static glibc builds)
# -extldflags=-static: statically link all C dependencies into the binary
# xx-verify --static: hard gate — fails the build if the binary is not fully static
RUN CGO_ENABLED=1 xx-go build \
    -tags "netgo osusergo" \
    -ldflags="-s -w -extldflags=-static" -trimpath -o /hatch ./cmd/hatch/ && \
    xx-verify --static /hatch

# Stage 3: final image — no shell, no package manager, no C runtime
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /hatch /hatch
ENTRYPOINT ["/hatch"]
```

### What each xx wrapper does

| Command         | What it does                                                                   |
| --------------- | ------------------------------------------------------------------------------ |
| `xx-go`         | Sets `GOARCH`, `GOOS`, `CC`, `CGO_CFLAGS`, `CGO_LDFLAGS` for `$TARGETPLATFORM` |
| `xx-apt-get`    | Configures Debian multiarch and installs target-arch packages into a sysroot   |
| `xx-verify`     | Reads the binary's ELF headers and fails if any dynamic libraries are linked   |

---

## musl vs glibc — why Alpine was rejected

An earlier attempt used `golang:1.26-alpine` as the builder. Alpine uses **musl libc**
instead of glibc. `sqlite-vec.c` references `u_int8_t`, `u_int16_t`, and `u_int64_t` —
BSD compatibility types that glibc defines but musl does not.

```text
sqlite-vec.c:68:9: error: unknown type name 'u_int8_t'; did you mean 'uint8_t'?
```

The fix is to use a **Debian-based builder** (`golang:1.26-bookworm`) which ships glibc.
The final image remains `distroless/static` (Debian-based, no musl involved).

| Image                       | libc   | `u_int8_t` defined? | Builder size |
| --------------------------- | ------ | ------------------- | ------------ |
| `golang:1.26-alpine`        | musl   | No ❌               | ~300 MB      |
| `golang:1.26-bookworm`      | glibc  | Yes ✅              | ~900 MB      |
| `distroless/static-debian12`| none   | N/A (runtime only)  | ~2 MB        |

The larger builder image only affects CI layer cache — the final image is still ~2 MB.

---

## The `-tags "netgo osusergo"` build tags

When statically linking a CGO binary against **glibc**, two glibc components cause
problems in a no-libc runtime:

- **NSS (Name Service Switch)** — glibc's DNS resolver uses dynamic plugin loading
  (`libnss_*.so`). In a static binary targeting `distroless/static`, those plugins
  don't exist.
- **User/group lookup** — glibc's `getpwuid` and `getgrgid` also use NSS.

The build tags tell the Go standard library to use its own implementations instead:

| Build tag    | Replaces                                               |
| ------------ | ------------------------------------------------------ |
| `netgo`      | glibc DNS resolver → pure-Go DNS                       |
| `osusergo`   | glibc user/group lookup → pure-Go `/etc/passwd` reader |

Without these tags, `hatch` would compile and link cleanly but crash at runtime the
first time it resolves a hostname or looks up a user.

---

## xx-verify — the static linkage gate

`xx-verify --static /hatch` reads the compiled binary's ELF dynamic section. If any
shared library is referenced (`libc.so.6`, `libm.so`, etc.), the step fails the build.

This acts as a hard gate: a dynamically linked binary cannot be pushed to the registry.
It catches mistakes like:

- Forgetting `CGO_ENABLED=1` (produces a pure-Go binary that passes but misses sqlite-vec)
- Missing `-extldflags=-static` (compiles fine but crashes in distroless)
- A new CGO dependency that links dynamically by default

---

## Related

- [`CGO.md`](CGO.md) — what CGO is and why hatch needs it
- [`GO_BINARY.md`](GO_BINARY.md) — what's inside a Go binary
- [ADR-015](../../../adr/ADR-015-cgo-cross-compilation.md) — the architecture decision record for this approach

---

## Resources

- [tonistiigi/xx](https://github.com/tonistiigi/xx) — the cross-compilation helper used in hatch's Dockerfile
- [Docker multi-platform builds](https://docs.docker.com/build/building/multi-platform/) — official Docker docs on `--platform` and BuildKit
- [Go CGO documentation](https://pkg.go.dev/cmd/cgo) — how Go calls C code
- [Go build constraints](https://pkg.go.dev/cmd/go#hdr-Build_constraints) — `netgo`, `osusergo`, and how build tags work
- [distroless images](https://github.com/GoogleContainerTools/distroless) — why Google's minimal images have no C runtime
- [musl vs glibc compatibility](https://wiki.musl-libc.org/functional-differences-from-glibc.html) — what musl intentionally omits
- [ELF dynamic linking](https://man7.org/linux/man-pages/man8/ld.so.8.html) — what dynamic vs static linking means at the OS level
