# Message Broker Design Document

## Overview

This project is a learning exercise focused on implementing a NATS-like message broker.

The design focuses on correctness and uses an actor-style model with a single broker.
This design manages concurrency without locks, but it limits maximum throughput because the broker is a single synchronous bottleneck.

The implementation supports:

1. a subset of the NATS protocol
2. client subscriptions and publishes
3. NATS wildcards `*` and `>`
4. fanout to multiple subscribers
5. at-most-once delivery
6. a NATS-like `PING` / `PONG` heartbeat
7. slow-client backpressure

## Goals and Non-Goals

## Architecture Invariants

## High-Level Message Flow

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
