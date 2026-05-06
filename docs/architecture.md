# Architecture

## Layer overview

```
┌─────────────────────────────────────────────────────────┐
│  github.com/kochemajaka/gofocas  (package focas)        │
│                                                         │
│  Dial()  Client.{System,Status,Axes,Spindles,...}       │
│  domain types  errors  options  reconnect               │
│                        │                                │
│        ┌───────────────▼──────────────┐                 │
│        │  internal/series             │                 │
│        │  Strategy interface          │                 │
│        │  per-series implementations  │                 │
│        └───────────────┬──────────────┘                 │
│                        │                                │
│        ┌───────────────▼──────────────┐                 │
│        │  internal/fwlib32            │                 │
│        │  Binder interface            │                 │
│        │  fwlib32.go  (CGo, focas_cgo tag)              │
│        │  fwlib32_stub.go  (!focas_cgo tag)             │
│        └──────────────────────────────┘                 │
└─────────────────────────────────────────────────────────┘
```

## Packages

### `focas` (root)

The only import consumers need. Contains:

- `Client` struct and its methods
- All domain types (`Axis`, `Status`, `Alarm`, …)
- `Error` type, sentinels, helper functions
- Options (`WithPort`, `WithReconnect`, …)
- Reconnect loop (`reconnect.go`)

The public package never imports `internal/fwlib32` types directly; it holds a `fwlib32.Binder` interface value. This keeps the CGo dependency fully contained.

### `series`

Public package for series constants (`S0i`, `S30i`, …) and `Parse`. Consumers import this alongside `focas` when they want to pass `WithSeries(series.S30i)`.

### `internal/series`

Implements the `Strategy` interface for each controller family. Strategies are pure Go and easily unit-tested without a CNC. The factory function `For(series.Series)` maps the public constant to the right strategy.

### `internal/fwlib32`

The only CGo layer. Two build-tag variants:

| File | Tag | Purpose |
|------|-----|---------|
| `fwlib32.go` | `focas_cgo` | Real binding wrapping `cnc_*` calls |
| `fwlib32_stub.go` | `!focas_cgo` | Returns `ErrUnsupported` for every call |

Both satisfy the `Binder` interface defined in `types.go`.

The package-level `sync.Mutex` (`mu`) serialises every FOCAS call. FOCAS is not reentrant across concurrent goroutines even on separate handles.

## Reconnect flow

```
call(ctx, op, fn)
  └─ fn() → transient error?
       └─ reconnectDo(ctx, fn)
            ├─ policy.Enabled && IsTransient(err)?
            ├─ wait (exponential back-off)
            ├─ redial() → close old handle, Alloc new handle
            └─ fn() again (up to MaxAttempts)
```

## Mutex rationale

`fwlib32` makes no thread-safety guarantees for concurrent calls on the same handle. A single process-wide mutex is the safest option. The trade-off is that parallel `errgroup` fan-outs serialize at the binding layer, but latency is dominated by network round-trips anyway, so the mutex is not a bottleneck in practice.
