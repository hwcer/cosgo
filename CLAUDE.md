# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

cosgo is a Go application scaffold and utility toolkit (`github.com/hwcer/cosgo`). It provides module lifecycle management, routing, serialization, session/storage, and assorted utilities. Go 1.24+.

## Common Commands

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./registry/...
go test ./schema/...

# Run a single test by name
go test ./zset/ -run TestZSetInsert

# Run benchmarks for a package
go test ./registry/ -bench=. -benchmem
go test ./storage/ -bench=. -benchmem

# Race detector
go test -race ./...

# Vet
go vet ./...
```

No Makefile, no linter config, no CI pipeline. Standard `go test` and `go vet` are the development tools.

## Code Conventions

- **Receiver name**: Always `this`, not single-letter. This is intentional and must be preserved.
- **Language**: Comments and documentation are in Chinese. Follow existing convention.
- **No modern Go idioms on purpose**: The codebase deliberately avoids Go 1.22+ features like `range over int`, `slices.Contains`, etc. Do not introduce them.
- **No `//nolint` or linter-driven changes**: Do not refactor to satisfy linters if it changes the existing style.
- **Registration pattern**: Many subsystems (binder, registry) use `Register()` at startup, then lock-free reads at runtime. Respect this init-time-only mutation pattern.

## Architecture

### Module Lifecycle (root package)

The root `cosgo` package orchestrates application startup via the `Module` interface (`Id`, `Init`, `Start`, `Close`). Modules are registered with `Use()` and driven by `Start()`:

```
Start() → Config.Init() → writePidFile → emit(EventTypBegin)
  → pprofStart → Module.Init() (all, in order) → emit(EventTypLoaded)
  → Options.Process() hook → Module.Start() (all) → emit(EventTypStarted)
  → WaitForSystemExit() [signal loop]
  → stop() → emit(EventTypClosing) → Module.Close() (reverse order)
  → scc.Wait() → deletePidFile → emit(EventTypStopped)
```

Configuration priority (high to low): runtime `Set()` > CLI flags (pflag) > env vars > config file (TOML) > defaults.

### Key Packages

- **registry/**: URL routing with radix tree. Static paths use hash lookup (~16ns), dynamic paths (`:id`, `*wildcard`) use tree traversal. `Service` groups handlers under a prefix; `Node` wraps reflected function values.

- **schema/**: Struct metadata cache. `Parse()` builds field accessors from struct tags, cached in `sync.Map`. Concurrent-safe with channel-based init coordination (`initDone chan struct{}`). Hot path is zero-alloc.

- **storage/**: Bucket-based object pool with O(1) get/set via 28-char hex tokens (2-byte bucket + 4-byte slot + 8-byte random). LIFO dirty index for slot recycling. `unsafe.Pointer` for values.

- **session/**: HTTP session with memory and Redis backends. Uses Copy-on-Write + `atomic.Pointer` for event listeners. Constant-time token comparison.

- **scc/**: Goroutine lifecycle manager. `GO()`, `CGO()` (with context), `SGO()` (with panic recovery). Global `Default` singleton. `Cancel()` + `Wait(timeout)` for graceful shutdown.

- **zset/**: Skip list + dictionary sorted set for leaderboards. Supports descending/ascending, gatekeeper (top-N soft limit), FIFO tie-breaking.

- **binder/**: Multi-format serializer registry (JSON, XML, YAML, Form, Protobuf, Msgpack, Bytes). All registration at init time; reads are lock-free.

- **values/**: Generic `Attach[K]` container with `*Message` (Code + Data) for error propagation. Recently refactored to use generics.

- **safety/**: IP whitelist/blacklist with zero-alloc IPv4 parsing. Copy-on-Write rules + `atomic.Pointer`.

### Concurrency Patterns

The codebase consistently uses these patterns:
- **Copy-on-Write + atomic.Pointer** for lock-free reads on rarely-mutated state (safety rules, session events)
- **sync.Map** for schema caching (concurrent reads, infrequent writes)
- **sync.RWMutex** for storage buckets and zset (concurrent readers, serialized writers)
- **Channel-based init coordination** in schema (`initDone chan struct{}`) to avoid duplicate work

### Dependencies

External: `github.com/hwcer/logger` (logging), `spf13/viper` + `spf13/pflag` (config/flags), `go-redis/redis/v8`, `mongo-driver/v2`, `gopsutil` (PID checks), `protobuf`, `yaml.v2`.

Packages are mostly independent and can be imported standalone (e.g., just `cosgo/zset` or `cosgo/slice`).
