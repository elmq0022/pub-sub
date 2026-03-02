# Message Broker Design Document

## Overview

This project is a learning exercise focused on implementing a NATS-like message broker.
For broader background on Core NATS, see the [official Core NATS concepts documentation](https://docs.nats.io/nats-concepts/core-nats). For protocol details, see the [official NATS client protocol reference](https://docs.nats.io/reference/reference-protocols/nats-protocol).

The design focuses on correctness and uses an actor-style model with a single broker.
This design manages concurrency without locks, but it limits maximum throughput because the broker is a single synchronous bottleneck.

The implementation supports:

1. a subset of the NATS protocol
2. client subscriptions and publishes
3. NATS wildcards `*` and `>`
4. fanout to multiple subscribers
5. at-most-once delivery
6. a NATS-like `PING` / `PONG` heartbeat
7. slow-client back pressure

## Goals and Non-Goals

### Goals

- At-most-once message delivery
- Support for a defined subset of the NATS protocol: CONNECT, SUB, PUB, UNSUB, PING, PONG, +OK, and -ERR
- Subject-based routing with `*` and `>` wildcards
- Client support for subscribe, publish, and unsubscribe operations
- Slow-connection handling that preserves system responsiveness

### Non-Goals

- Full Core NATS compatibility
- Production readiness
- Message persistence
- Event replay
- Clustering
- Authentication
- At-least-once or exactly-once delivery guarantees
- Replication or horizontal scaling
- Observability beyond logs

## Architecture Invariants

- The session controller creates and starts the reader and writer loops for each connection.
- Readers receive and decode wire messages from the client connection.
- Writers receive outbound commands from the broker, encode them, and write them to the client connection.
- All protocol, lifecycle, and heartbeat events flow through a single synchronous broker.
- The broker is the sole owner of mutable application state.
- Outbound broker-to-writer messaging must never block.
- Only the broker interacts directly with the subject registry.
- The broker processes events in channel receive order.
- When a connection closes, the broker removes that client from session state and subscriptions.
- Readers and writers never communicate directly; they coordinate only through the broker.
- CID values are monotonically increasing `int64` values and are unique for the lifetime of the process.
- SID reuse is allowed and may result in duplicate deliveries.
- The broker does not attempt to deduplicate subscriptions for a client.
- `UNSUB` semantics are only reliable for subscriptions the registry can still directly index.
- Slow connections are dropped when delivering to them would block the broker.
- The writer is the sole writer for a connection.
- The writer processes commands in the order they are received from its outbound channel.
- The reader is the sole reader to a connection.

## High-Level Message Flow

![High-level message flow](references/message-flow.svg)

A simplified view of how session, protocol, broker, and registry interactions move through the system.

## Components

### Wire Protocol Encoder / Decoder

### Session Controller

#### Session Lifecycle

#### Reader Loop

#### Writer Loop

### Broker

#### Event Types

#### Session State

#### Command Handling

#### Disconnect Policy

### Subject Registry

#### Data Model

#### Lookup Semantics

#### Subscription Removal and Pruning

## Runtime Flows

### Connection Startup

### Publish Delivery

### Protocol Error Handling

### Connection Shutdown

## Backpressure and Slow Consumers

## Heartbeats and Liveness

## Error Handling Strategy

## Known Limitations

## Future Improvements
