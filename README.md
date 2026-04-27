# go104

IEC 60870-5-104 master server with embedded web UI.

Single self-contained binary: Go backend + Vue 3 frontend compiled in.

---

## Features

- **Full IEC104 master** — I/S/U frames, T1/T2/T3 timers, K/W windows, GI lifecycle
- **Multi-line** — each communication line is an independent goroutine with its own TCP connection and sequence counters
- **Signal mapping** — map IOA addresses to named analog/digital signals per line; scale/offset support for analog values
- **Live monitor** — WebSocket-pushed real-time table; stale detection on disconnect; GI-in-progress indicator
- **Commands** — single-point, double-point and setpoint commands (C_SC/C_DC/C_SE)
- **Single binary** — Vue dist is embedded via `go:embed`; no separate web server needed
- **Pure Go** — uses `modernc.org/sqlite` (no CGo); cross-compiles cleanly

---

## Quick start

```bash
# Development (two terminals)
make dev-backend      # Go server with live reload
make dev-frontend     # Vite dev server at :5173 with HMR

# Production build + run
make run              # build everything, serve at :8080
make run PORT=9000    # custom port
```

Default port: **8080**. Set `HTTP_PORT` env var or use `PORT=` on the make target.

---

## Building

### Prerequisites

| Tool | Version |
|------|---------|
| Go   | 1.22+   |
| Node | 20+     |
| pnpm | 9.x     |

```bash
make build            # host architecture (development)
make release          # all release targets (see below)
```

### Release targets

| Target | Binary | Platform |
|--------|--------|----------|
| `make release-amd64`   | `bin/go104-linux-amd64`   | Red Hat / x86\_64 (RHEL 7+, Rocky, AlmaLinux) |
| `make release-ppc64le` | `bin/go104-linux-ppc64le` | IBM POWER LE (RHEL for POWER / OpenPOWER)     |
| `make release`         | both                      | builds all targets                             |

Cross-compilation requires no additional toolchain — the SQLite driver is pure Go.

```bash
make release
ls -lh bin/
# go104-linux-amd64
# go104-linux-ppc64le
```

Copy the binary to the target host and run:

```bash
HTTP_PORT=8080 ./go104-linux-amd64
```

---

## Configuration

All configuration is done through the web UI at `http://<host>:<port>`.

### Communication line parameters

| Field | Default | Description |
|-------|---------|-------------|
| Host / Port | — | RTU/gateway address |
| Common Address | 1 | ASDU common address (CA) |
| K | 12 | Max unacknowledged I-frames |
| W | 8 | ACK after W received I-frames |
| T1 (ms) | 15000 | Unacknowledged frame / response timeout |
| T2 (ms) | 10000 | Delayed S-frame ACK timeout |
| T3 (ms) | 20000 | Idle TESTFR heartbeat interval |
| GI Interval (s) | 0 | Periodic GI; 0 = disabled |

### Signal kinds

| Kind | TypeIDs |
|------|---------|
| Digital | M_SP_NA_1 (1), M_DP_NA_1 (3), M_SP_TB_1 (30), M_DP_TB_1 (31) |
| Analog  | M_ME_NA_1 (9), M_ME_NB_1 (11), M_ME_NC_1 (13), M_ME_TF_1 (36), … |

Analog value delivered to the UI: `raw_value × scale + offset`.

---

## Architecture

```
cmd/server/
  main.go              HTTP server, router, SPA handler

internal/
  iec104/
    apci.go            Frame encode/decode (I/S/U), SeqDiff
    asdu.go            ASDU parse + build (GI, commands, CP56Time2a)
    typeid.go          TypeID / COT constants
    line_worker.go     Per-line goroutine: TCP, timers, GI lifecycle
    master.go          Worker registry (start/stop/restart lines)
  api/
    handlers.go        HTTP handler wiring
    lines.go           Line CRUD + start/stop/GI endpoints
    signals.go         Signal CRUD
    monitor.go         Snapshot endpoint
    commands.go        Command send endpoint
  hub/
    hub.go             WebSocket broadcast hub
  models/
    models.go          Shared structs (Line, Signal, Datapoint, Command)
  store/
    store.go           SQLite store (WAL mode, migrations)
  ui/
    embed.go           go:embed all:dist

web/                   Vue 3 + Pinia + Vue Router + Tailwind CSS
  src/
    api/client.ts      Typed HTTP + WS client
    stores/
      lines.ts         Line list + live status
      signals.ts       Signal list per line
      monitor.ts       WS datapoints, stale tracking, GI-in-progress
    views/
      Lines.vue        Line management
      Signals.vue      Signal mapping
      Monitor.vue      Live signal table
      Commands.vue     Command panel
```

---

## WebSocket events

| Event | Direction | Payload |
|-------|-----------|---------|
| `datapoint`    | server → client | Full `Datapoint` struct + `type` field |
| `line_status`  | server → client | `{line_id, state, rx_count, tx_count, addr}` |
| `gi_start`     | server → client | `{line_id}` — GI sent to slave |
| `gi_complete`  | server → client | `{line_id}` — ActTerm received |
| `signal_stale` | server → client | `{signal_id, line_id}` — line disconnected |

---

## License

MIT
