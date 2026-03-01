# pub-sub

`pub-sub` is a small Go learning project that explores how to build a minimal publish/subscribe server with a NATS-like text protocol.

The codebase is focused on a simple actor-style design:
- a TCP server accepts client connections on `:8080`
- each client gets a reader loop and a writer loop
- a single broker goroutine owns routing state and session lifecycle
- subjects are matched through a registry that supports wildcards

This project is mainly for learning and experimentation with:
- Go concurrency and goroutine ownership
- protocol decoding and encoding
- fanout message routing
- heartbeat and slow-consumer handling

## Project Layout

- [`cmd/main.go`](/home/zero/Projects/golang/pub-sub/cmd/main.go): starts the TCP server, broker, and session controller
- [`internal/broker/broker.go`](/home/zero/Projects/golang/pub-sub/internal/broker/broker.go): central broker loop and heartbeat logic
- [`internal/sessioncontroller/session_controller.go`](/home/zero/Projects/golang/pub-sub/internal/sessioncontroller/session_controller.go): per-connection reader and writer loops
- [`internal/codec/codec.go`](/home/zero/Projects/golang/pub-sub/internal/codec/codec.go): wire protocol parsing and encoding
- [`internal/subjectregistry/subject_registry.go`](/home/zero/Projects/golang/pub-sub/internal/subjectregistry/subject_registry.go): subject routing and wildcard lookup
- [`DESIGN.md`](/home/zero/Projects/golang/pub-sub/DESIGN.md): design notes for the architecture

## Run

```bash
go run ./cmd
```

The server listens on `localhost:8080`.

## Test

```bash
go test ./...
```

This is not intended to be production-ready. 
It is a sandbox for learning how a pub-sub server can be structured in Go.
