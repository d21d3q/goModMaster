---
title: 'goModMaster Web MVP'
slug: 'gomodmaster-web-mvp'
created: 'Fri Jan 23 10:14:54 CET 2026'
status: 'Implementation Complete'
stepsCompleted: [1, 2, 3, 4]
tech_stack: ['Go', 'React', 'Tailwind CSS', 'Cobra', 'Echo', 'Vite', 'Testify', 'github.com/simonvetter/modbus']
files_to_modify: ['cmd/gmm/**', 'internal/core/**', 'internal/transport/http/**', 'internal/transport/ws/**', 'web/**', 'embed/**', 'go.mod', 'Makefile or justfile']
code_patterns: ['Clean separation: core Modbus logic + transport adapters (web now, TUI later)', 'Single WS endpoint with typed messages', 'REST for config/actions; WS for live updates', 'Read-only MVP; in-memory logs and stats']
test_patterns: ['Go testing + Testify for unit tests', 'Frontend tests optional later']
---

# Tech-Spec: goModMaster Web MVP

**Created:** Fri Jan 23 10:14:54 CET 2026

## Overview

### Problem Statement

Build a single-binary Modbus master that launches a local web UI for dense, precise diagnostics/commissioning workflows, with a clean shared core that can later power a TUI.

### Solution

Implement `gmm web` to start a local React UI (embedded FS, hash routing, Tailwind styling) backed by a Go core using `github.com/simonvetter/modbus`. Use REST for configuration/actions and a single WebSocket endpoint for live updates/logs/stats. MVP is read-only with in-memory raw frame logs and user-selectable value interpretations.

### Scope

**In Scope:**
- Modbus RTU and TCP support via `simonvetter/modbus`
- Local web UI (React, embedded FS, hash routing, Tailwind CSS)
- Connection configuration via CLI and UI (serial/TCP params)
- Dense row-wise coil/register views
- Hex/dec address and value toggles (default 1-based addressing)
- Optional user-selected decoders: uint16/int16/uint32/int32/float32 with configurable byte/word order
- Manual read actions (no polling)
- Latency and error statistics
- In-memory raw frame log with live updates
- Display current invocation string for copy/reuse

**Out of Scope:**
- TUI implementation
- Modbus ASCII
- Write operations
- Automated polling/backoff
- File logging or persisted profiles
- Cross-arch packaging/binaries

## Context for Development

### Codebase Patterns

- Clean separation between core Modbus logic and transport/UI adapters to enable future TUI reuse.

### Files to Reference

| File | Purpose |
| ---- | ------- |
| (none) | Clean slate repo; establish initial structure |
### Technical Decisions

- CLI with Cobra
- HTTP server with Echo
- Frontend built with Vite + React + Tailwind
- Single WebSocket endpoint with typed messages (topics: data, logs, stats, errors)
- REST for config/actions; WS for live updates
- Hash routing for UI
- Embedded web assets in Go binary
- No persistence for MVP; CLI or UI config only
- Default to local-only access with a generated URL token; optional flag to disable token check

## Implementation Plan

### Tasks

- [x] Task 1: Bootstrap Go module, CLI, and config model
  - File: `go.mod`
  - Action: Initialize module and add dependencies (`cobra`, `echo`, `simonvetter/modbus`, `testify`)
  - Notes: Keep module path aligned with repo name
  - File: `cmd/gmm/main.go`
  - Action: Wire Cobra root command and subcommand `web`
  - Notes: `gmm web` should accept serial/TCP flags, generate a URL token by default, and launch server
  - File: `internal/config/config.go`
  - Action: Define config structs for serial/TCP parameters, address/value display settings, decoder options
  - Notes: Defaults: 1-based addressing, hex/dec toggles; no persistence; token enabled unless flag disables

- [x] Task 2: Implement core Modbus service layer (read-only)
  - File: `internal/core/client.go`
  - Action: Wrap `simonvetter/modbus` client setup for RTU/TCP and expose read methods
  - Notes: Keep interface transport-agnostic for future TUI
  - File: `internal/core/read.go`
  - Action: Implement reads for coils, discrete inputs, holding registers, input registers
  - Notes: Return raw values plus timing and error metadata
  - File: `internal/core/decoders.go`
  - Action: Implement optional value decoders (uint16/int16/uint32/int32/float32) with byte/word order options
  - Notes: Only apply when user selects decoder options
  - File: `internal/core/logs.go`
  - Action: In-memory raw frame log buffer and stats tracking (latency/error counts)
  - Notes: Ring buffer with max size; emit events to WS

- [x] Task 3: Add HTTP API (REST + WS)
  - File: `internal/transport/http/server.go`
  - Action: Start Echo server, enforce optional URL token, serve embedded web assets, and REST endpoints
  - Notes: Endpoints for connect/disconnect, read requests, config updates, stats snapshot; token printed to console
  - File: `internal/transport/ws/hub.go`
  - Action: Single WS endpoint with typed messages (data/logs/stats/errors)
  - Notes: Implement subscribe semantics client-side via message type
  - File: `internal/transport/http/routes.go`
  - Action: Register REST routes and WS route, bind to handlers
  - Notes: Keep handlers thin; call core service

- [x] Task 4: Build web UI shell and data views
  - File: `web/package.json`
  - Action: Vite + React + Tailwind setup
  - Notes: Hash routing enabled
  - File: `web/src/main.tsx`
  - Action: Bootstrap app, router, and WS client
  - Notes: Single WS client with typed message handling
  - File: `web/src/components/ReadPanels.tsx`
  - Action: Dense row-wise views for coils and registers with hex/dec toggles
  - Notes: Manual read actions only
  - File: `web/src/components/DecoderPanel.tsx`
  - Action: UI for selecting decoders and byte/word order
  - Notes: Decoded rows shown as additional rows
  - File: `web/src/components/RawLog.tsx`
  - Action: Streaming raw frame log viewer
  - Notes: In-memory only
  - File: `web/src/components/Stats.tsx`
  - Action: Latency/error stats display
  - Notes: Use WS updates with REST fallback snapshot
  - File: `web/src/components/Invocation.tsx`
  - Action: Display current invocation string for copy/reuse
  - Notes: Uses current config from server

- [x] Task 5: Embed web build and wire build tooling
  - File: `embed/web.go`
  - Action: Embed built web assets for serving via Echo
  - Notes: Use Go embed for `web/dist`
  - File: `justfile` or `Makefile`
  - Action: Add build targets for web build and Go binary
  - Notes: `just build` should run `vite build` then `go build`

- [x] Task 6: Add backend tests
  - File: `internal/core/decoders_test.go`
  - Action: Unit test decoder conversions and byte/word order
  - Notes: Use Testify assertions
  - File: `internal/core/logs_test.go`
  - Action: Unit test ring buffer behavior and stats aggregation
  - Notes: Use Testify and Go testing

### Acceptance Criteria

- [ ] AC 1: Given `gmm web` is run without flags, when the server starts, then the web UI is served locally and a default, editable config is shown.
- [ ] AC 1a: Given `gmm web` is run without flags, when the server starts, then a URL token is printed to the console and required as a URL query param to access the UI.
- [ ] AC 2: Given a TCP connection is configured, when the user triggers a read, then the UI shows the returned values and the WS stream updates data/logs/stats.
- [ ] AC 3: Given an RTU configuration is provided via CLI flags, when the app starts, then the UI reflects those settings and the invocation string matches the active config.
- [ ] AC 4: Given the user toggles hex/dec addressing, when values are rendered, then addresses and values display in the selected base without changing the underlying read.
- [ ] AC 5: Given a decoder selection (e.g., float32 with word order), when a read completes, then decoded rows appear as additional rows with correct interpretation.
- [ ] AC 6: Given errors occur during reads, when the error is emitted, then the UI shows the error and error stats increment while the app remains responsive.
- [ ] AC 7: Given reads occur, when raw frames are produced, then the in-memory log shows new frames and does not persist across restarts.
- [ ] AC 8: Given the user starts `gmm web --no-token` (or equivalent), when the server starts, then the UI is accessible without a token.


## Additional Context

### Dependencies

- https://github.com/simonvetter/modbus
- https://github.com/spf13/cobra
- https://github.com/labstack/echo
- https://github.com/stretchr/testify
- Vite + React + Tailwind CSS

### Testing Strategy

- Go `testing` + Testify for unit tests on decoders/logs; manual integration testing for Modbus reads; frontend tests deferred.

### Notes

- Future TUI will reuse core logic; keep UI adapters thin.
- Polling is out of scope but design core/read API so polling can be added without breaking UI/WS contract.
