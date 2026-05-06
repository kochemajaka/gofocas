# gofocas

[![Go Reference](https://pkg.go.dev/badge/github.com/kochemajaka/gofocas.svg)](https://pkg.go.dev/github.com/kochemajaka/gofocas)
[![CI](https://github.com/kochemajaka/gofocas/actions/workflows/ci.yml/badge.svg)](https://github.com/kochemajaka/gofocas/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Go client for FANUC CNC controllers via the FOCAS Ethernet library (`fwlib32`).
Read machine status, axis positions, spindle data, alarms, active programs, and production counters from any series 0i / 15 / 16i / 18i / 30i / 31i / 32i controller.

> **Status:** v0.x — API may change before v1.0.

---

## Features

- `Dial` / `Close` / `Ping` — connection lifecycle with configurable timeout
- `System`, `Status` — controller info and operating state
- `Axes` — positions (absolute / machine / relative / distance-to-go) + servo diagnostics
- `Spindles` — speed, load, power
- `Feed`, `ContourFeedRate`, `FeedOverride`, `JogOverride`
- `ExecutingProgram`, `ProgramSource` — currently running program + G-code block
- `Alarms` — active alarms with type taxonomy and axis association
- `Parameters` — production counters and runtime durations
- `Parameter`, `Diagnosis` — generic single-parameter access
- Automatic reconnect on transient errors (`EW_HANDLE`, `EW_SOCKET`)
- Series-aware strategy layer — quirks of 0i / 15 / 16i handled transparently
- Builds without the FANUC SDK (stub mode); real CGo binding behind `focas_cgo` tag

---

## Series support matrix

| Feature             | 0i | 15 | 15i | 16 | 16i | 18i | 21 | 30i | 31i | 32i |
|---------------------|:--:|:--:|:---:|:--:|:---:|:---:|:--:|:---:|:---:|:---:|
| System / Status     | v  | v  | v   | v  | v   | v   | v  | v   | v   | v   |
| Axes (position)     | v  | v  | v   | v  | v   | v   | v  | v   | v   | v   |
| Spindles            | v  | v  | v   | v  | v   | v   | v  | v   | v   | v   |
| Feed / Override     | v  | v  | v   | v  | v   | v   | v  | v   | v   | v   |
| Program / Source    | v  | v  | v   | v  | v   | v   | v  | v   | v   | v   |
| Alarms              | v  | v  | v   | v  | v   | v   | v  | v   | v   | v   |
| Production params   | v  | -  | v   | v  | v   | v   | v  | v   | v   | v   |

---

## Installation

```sh
go get github.com/kochemajaka/gofocas
```

Then install the FANUC `fwlib32` SDK from [`vendor/fanuc/`](vendor/fanuc/) — pre-built binaries for Linux (x64/x86/ARMv7) and Windows are included. Run `sudo sh vendor/fanuc/install.sh` on Linux or copy `Fwlib32.dll` on Windows. Full FOCAS API reference: [inventcom.net/fanuc-focas-library](https://www.inventcom.net/fanuc-focas-library).

Build with `-tags=focas_cgo` to enable the real binding; omit the tag for stub/test mode.

---

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kochemajaka/gofocas"
)

func main() {
    ctx := context.Background()

    client, err := focas.Dial(ctx, "192.168.1.1",
        focas.WithReconnect(focas.DefaultReconnectPolicy()),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    sys, _ := client.System(ctx)
    st, _  := client.Status(ctx)
    axes, _ := client.Axes(ctx)

    fmt.Println(sys.Series, st.Mode, st.Run)
    for _, a := range axes {
        fmt.Printf("%s: %.3f mm\n", a.Name, a.Position.Absolute)
    }
}
```

---

## Examples

| Example | Description |
|---------|-------------|
| [01-connect](examples/01-connect/) | Minimal dial + ping + system info |
| [02-read-position](examples/02-read-position/) | Axes & spindles on a 1s ticker |
| [03-read-alarms](examples/03-read-alarms/) | Print active alarms |
| [04-parallel-read](examples/04-parallel-read/) | Fan-out all readers with errgroup, print JSON |
| [05-stream-prometheus](examples/05-stream-prometheus/) | Prometheus exporter for feed/spindle metrics |
| [06-multi-machine](examples/06-multi-machine/) | Concurrent reads from multiple controllers |

---

## Architecture

```
+--------------------------------------------------+
|  package focas  (public API)                     |
|  Client . domain types . errors . options        |
|                  |                               |
|     +------------v------------+                  |
|     |  internal/series        |  per-series      |
|     |  Strategy interface     |  strategies      |
|     +------------+------------+                  |
|                  |                               |
|     +------------v------------+                  |
|     |  internal/fwlib32       |  CGo binding     |
|     |  Binder interface       |  (+ stub)        |
|     +-------------------------+                  |
+--------------------------------------------------+
```

Only `internal/fwlib32` contains CGo / C includes. Everything above is pure Go and fully testable without the FANUC SDK. See [docs/architecture.md](docs/architecture.md) for details.

---

## Reconnect / error handling

```go
// Check the FOCAS error code:
if code, ok := focas.CodeOf(err); ok {
    fmt.Println(code) // e.g. EW_SOCKET
}

// Check for any transient (reconnectable) error:
if focas.IsTransient(err) { /* redial logic */ }

// Errors wrap sentinel values:
if errors.Is(err, focas.ErrClosed) { ... }
```

Enable automatic reconnect:

```go
client, err := focas.Dial(ctx, addr,
    focas.WithReconnect(focas.ReconnectPolicy{
        Enabled:     true,
        MaxAttempts: 10,
        InitialWait: 200 * time.Millisecond,
        MaxWait:     5 * time.Second,
        Multiplier:  2.0,
    }),
)
```

---

## Parallel reads recipe

```go
g, gCtx := errgroup.WithContext(ctx)
var axes   []focas.Axis
var alarms []focas.Alarm
g.Go(func() error { var e error; axes, e = client.Axes(gCtx); return e })
g.Go(func() error { var e error; alarms, e = client.Alarms(gCtx); return e })
if err := g.Wait(); err != nil { ... }
```

---

## FAQ

**Why CGo?** `fwlib32` is a proprietary C library distributed as `.dll` / `.so`. There is no pure-Go Ethernet implementation for the FOCAS protocol.

**Can I run on macOS?** You can build and test in stub mode (`-tags='!focas_cgo'`). The real binding requires Linux or Windows because FANUC only ships `fwlib32` for those platforms.

**Why no write methods?** Write operations (`cnc_wrparam`, program transfer, cycle start) carry risk on production machines. They are tracked in `docs/roadmap.md` for a future release.

**What about the fwlib32 license?** `fwlib32` is proprietary FANUC software distributed by FANUC under their terms. Pre-built binaries and the C header are bundled in [`vendor/fanuc/`](vendor/fanuc/) for convenience — see [vendor/fanuc/README.md](vendor/fanuc/README.md) for install instructions. Full FOCAS API documentation is available at [inventcom.net/fanuc-focas-library](https://www.inventcom.net/fanuc-focas-library).

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
