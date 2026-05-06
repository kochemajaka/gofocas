# fwlib32 Setup

`fwlib32` is FANUC's proprietary Ethernet library. It is **not** distributed with this repository. You must obtain it from FANUC (typically included with the FOCAS2 Ethernet function option for your CNC) or from a FANUC partner.

## Files you need

| Platform | Files |
|----------|-------|
| Linux x64 | `libfwlib32-linux-x64.so.1.0.5` (or equivalent), `fwlib32.h` |
| Linux arm64 | `libfwlib32-linux-armv7.so.1.0.5`, `fwlib32.h` |
| Windows | `Fwlib32.dll`, `Fwlib32.lib`, `fwlib32.h` |

## Placement

Place the files inside `internal/fwlib32/`:

```
internal/fwlib32/
├── fwlib32.h          ← copy here
├── libfwlib32*.so     ← Linux
├── Fwlib32.dll        ← Windows
├── Fwlib32.lib        ← Windows
├── fwlib32.go         (tracked in git)
├── fwlib32_stub.go    (tracked in git)
├── helpers.h          (tracked in git)
└── types.go           (tracked in git)
```

These files are listed in `.gitignore` — they will never be committed.

## Linux build

```sh
export CGO_LDFLAGS="-L$(pwd)/internal/fwlib32 -lfwlib32 -Wl,-rpath,$(pwd)/internal/fwlib32"
go build -tags=focas_cgo ./...
```

At runtime the shared library must be discoverable. Either keep it in the same directory as the binary or add its path to `LD_LIBRARY_PATH`.

## Windows build (MinGW / CGo)

```cmd
set CGO_CFLAGS=-Iinternal\fwlib32
set CGO_LDFLAGS=-Linternal\fwlib32 -lFwlib32
go build -tags=focas_cgo ./...
```

`Fwlib32.dll` must be on `PATH` at runtime.

## Verifying the setup

Run the minimal connection example (requires a real CNC on the network):

```sh
FOCAS_ADDR=192.168.1.1 go run -tags=focas_cgo ./examples/01-connect/
```

## Troubleshooting

- **`EW_NODLL (-15)`** — the shared library is not on the runtime search path.
- **`EW_SOCKET (-16)`** — network unreachable; check IP address and that FOCAS is enabled on the CNC (parameter 904 on 30i).
- **`EW_HANDLE (-8)` on first call** — `cnc_startupprocess` failed or returned an error. Check the log file (default `fanuc.log`).
- **CGo disabled error at compile time** — ensure you are using a CGo-capable toolchain and have not set `CGO_ENABLED=0`.
