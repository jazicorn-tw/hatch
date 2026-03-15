<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-15
status:       active
tags:         [onboarding, go]
description:  "What CGO is, why hatch needs it for sqlite-vec, and what to do when it breaks"
-->
# CGO — Go Calling C Code

Most Go code is pure Go. But hatch uses **sqlite-vec**, a C extension for SQLite that
provides vector search. To call C code from Go, Go uses a bridge called **CGO**.

---

## Why hatch needs CGO

Hatch needs fast approximate nearest-neighbour (KNN) search over embedding vectors.
That capability comes from `sqlite-vec` — a C extension that plugs into SQLite.

The Go driver `mattn/go-sqlite3` is a wrapper around the actual SQLite C library. It
uses CGO to call into that C code. Because of this, building hatch requires a C compiler
(`gcc` or `clang`) to be present on your machine.

Without CGO, hatch would have to use the pure-Go SQLite driver (`modernc.org/sqlite`),
which can't load C extensions — so sqlite-vec wouldn't work.

---

## What CGO means in practice

**For development** — mostly invisible. Install Xcode command line tools on macOS and
`go build` handles the rest:

```bash
# macOS — install C compiler if you don't have it:
xcode-select --install

# Then build normally:
go build -o hatch ./cmd/hatch
```

**Common error if C compiler is missing:**

```text
cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in $PATH
```

Fix: install Xcode command line tools (macOS) or `build-essential` (Ubuntu/Debian):

```bash
# Ubuntu / Debian
sudo apt-get install build-essential
```

---

## The binary is still one file

Even though CGO compiles C code, the result is **statically compiled in** — the SQLite
C code ends up inside the `hatch` binary, not as a separate `.so` shared library that
you'd have to distribute alongside it.

```text
hatch    ← one file, contains Go code + SQLite C code + sqlite-vec C code
```

You do not need to install SQLite separately on any machine that runs `hatch`.

---

## The cross-compilation trade-off

Pure Go code can be cross-compiled with a single env var:

```bash
# Pure Go — build a Linux binary from macOS, no problem
GOOS=linux GOARCH=amd64 go build ./cmd/hatch
```

CGO breaks this because cross-compiling C code requires a C compiler that targets the
destination OS/architecture — which most machines don't have set up.

**Practical impact for hatch:** CI (GitHub Actions) builds the release binary on a Linux
runner, so this is handled automatically. You don't need to cross-compile locally.

---

## Related

- [`GO_BINARY.md`](GO_BINARY.md) — what a Go binary is and what's inside it
- [`GO_EMBED.md`](GO_EMBED.md) — how hatch bakes SQL migrations into the binary
- [`CROSS_COMPILATION.md`](CROSS_COMPILATION.md) — how hatch builds CGO binaries for amd64 and arm64 simultaneously

## Resources

- [CGO reference](https://pkg.go.dev/cmd/cgo) — official CGO documentation
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) — the CGO SQLite driver used by hatch
- [sqlite-vec](https://github.com/asg017/sqlite-vec) — the vector search C extension compiled into hatch
- [Go cross-compilation](https://go.dev/doc/install/source#environment) — `GOOS` and `GOARCH` for targeting other platforms
