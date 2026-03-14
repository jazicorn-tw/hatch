<!--
created_by:   jazicorn-tw
created_date: 2026-03-14
updated_by:   jazicorn-tw
updated_date: 2026-03-14
status:       active
tags:         [onboarding, go, concurrency]
description:  "What goroutines and channels are, and where hatch uses them"
-->
# Goroutines and Channels

Go has built-in concurrency through **goroutines** (lightweight threads) and
**channels** (typed message queues). They're central to how hatch handles concurrent
SSH sessions and progress reporting during ingestion.

---

## Goroutines

A goroutine is a function that runs concurrently with the rest of your program. You
start one with the `go` keyword:

```go
go doSomething()     // starts doSomething in a new goroutine
go func() {          // anonymous goroutine
    fmt.Println("running concurrently")
}()
```

Goroutines are cheap. You can have thousands of them — they're not OS threads. The Go
runtime schedules them across available CPU cores automatically.

The calling function doesn't wait for the goroutine to finish. If you need to wait, you
use a channel or `sync.WaitGroup`.

---

## Channels

A channel is a typed queue for passing values between goroutines:

```go
ch := make(chan int)       // unbuffered: sender blocks until receiver is ready
ch := make(chan int, 16)   // buffered: sender can send up to 16 values without blocking
```

Send a value: `ch <- 42`
Receive a value: `val := <-ch`
Close when done: `close(ch)` — receivers can range over a closed channel

---

## How hatch uses goroutines and channels

### 1. Progress bar during ingestion

When you run `hatch ingest --source hatch-docs`, the ingestion pipeline runs in the
main goroutine and a second goroutine drives the terminal progress bar concurrently.

From `cmd/hatch/ingest.go`:

```go
progressCh := make(chan pipeline.Progress, 16)   // buffered channel
barDone := drainProgressBar(sourceName, progressCh)

runErr := pipeline.Run(ctx, src, newDispatchChunker(), emb, st, progressCh)
close(progressCh)   // signal the progress goroutine to stop
<-barDone           // wait for it to finish before exiting
```

`drainProgressBar` starts the goroutine:

```go
func drainProgressBar(sourceName string, ch <-chan pipeline.Progress) <-chan struct{} {
    done := make(chan struct{})
    go func() {
        defer close(done)          // signal caller we're done
        defer bar.Finish()
        for p := range ch {        // range stops when ch is closed
            bar.ChangeMax(p.Total)
            _ = bar.Set(p.Done)
        }
    }()
    return done
}
```

The pattern:

1. Main goroutine creates `progressCh` (buffered — pipeline doesn't block on progress)
2. `drainProgressBar` spawns a goroutine that reads from `progressCh` and updates the bar
3. It returns a `done` channel (empty struct — used as a signal only, no data)
4. After pipeline finishes, main closes `progressCh` → the goroutine's `range` loop ends
5. The goroutine closes `done` via `defer close(done)`
6. Main blocks on `<-barDone` to wait for the goroutine to finish cleanly

### 2. Concurrent SSH sessions

When hatch's SSH server (Wish) accepts multiple connections, each session gets its own
goroutine running an independent Bubble Tea program. Sessions don't share state.

```text
goroutine 1: SSH session for alice — running quiz loop
goroutine 2: SSH session for bob   — running quiz loop
goroutine 3: main server loop      — listening for new connections
```

WAL mode on SQLite lets each session read the database concurrently without blocking
each other (see [`SQLITE_WAL.md`](../data/SQLITE_WAL.md)).

---

## Directional channel types

Go lets you restrict a channel to send-only or receive-only in function signatures:

```go
func producer(ch chan<- int)   // can only send to ch
func consumer(ch <-chan int)   // can only receive from ch
```

In `drainProgressBar`, the parameter is `ch <-chan pipeline.Progress` — the goroutine
can only receive, not accidentally send. In `pipeline.Run`, the parameter is
`progressCh chan<- Progress` — the pipeline can only send. This is enforced at compile
time.

---

## The empty struct signal pattern

`chan struct{}` is the idiomatic Go way to signal "done" without passing data:

```go
done := make(chan struct{})
go func() {
    defer close(done)   // closing is the signal
    // ... do work ...
}()
<-done   // block until the goroutine closes done
```

An empty struct takes zero bytes, so it's the cheapest possible channel value. You'll
see this pattern wherever hatch needs to wait for a background goroutine to finish.

---

## What to watch for

- **Goroutine leaks**: if a goroutine is waiting on a channel that never gets a value or
  never gets closed, it runs forever. Hatch avoids this by always closing `progressCh`
  before `<-barDone`.
- **Race conditions**: two goroutines writing the same memory simultaneously causes
  undefined behaviour. Use `go test -race` to detect these. Hatch's CI runs the race
  detector.
- **Context cancellation**: `pipeline.Run` checks `ctx.Err()` between documents — this
  lets the pipeline stop cleanly if the user hits Ctrl-C.

---

## Related

- [`BUBBLE_TEA.md`](../tui/BUBBLE_TEA.md) — how Bubble Tea runs one goroutine per SSH session
- [`SQLITE_WAL.md`](../data/SQLITE_WAL.md) — how WAL mode handles concurrent reads from many goroutines

## Resources

- [A Tour of Go: Goroutines](https://go.dev/tour/concurrency/1) — interactive intro
- [A Tour of Go: Channels](https://go.dev/tour/concurrency/2) — channel basics
- [Go blog: Share Memory by Communicating](https://go.dev/blog/codelab-share) — the philosophy behind channels
- [Go race detector](https://go.dev/doc/articles/race_detector) — how to run and interpret `-race`
