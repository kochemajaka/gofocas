# Series Support

## Feature matrix

| Feature | 0i | 15 | 15i | 16 | 16i | 18i | 21 | 30i | 31i | 32i |
|---------|:--:|:--:|:---:|:--:|:---:|:---:|:--:|:---:|:---:|:---:|
| `System` | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `Status` | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `Axes` | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `Spindles` | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `Feed` | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `ContourFeedRate` | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `FeedOverride` | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `JogOverride` | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `ExecutingProgram` | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `ProgramSource` | Y[1] | Y[1] | Y[1] | Y[1] | Y[1] | Y[1] | Y[1] | Y | Y | Y |
| `Alarms` | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `Parameters` (prod counters) | Y | — | Y | Y | Y | Y | Y | Y | Y | Y |
| `Parameter` (generic) | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `Diagnosis` (generic) | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |

**[1]** Buffer limited to 128 bytes on older series (vs 256 on 30i/31i/32i).

## Known quirks

### 0i (S0i)
- `cnc_rdexecprog` buffer must not exceed 128 bytes.
- Some 0i-D/F variants report `AutoMode` in a different byte than 0i-C; the strategy normalises this.

### 15 (S15)
- No production counter parameters (#6711 / #6750-#6757) — `Parameters()` returns `ErrUnsupported` on series 15.
- `cnc_rdexecprog` buffer limited to 64 bytes.

### 16 / 16i / 18i / 21 (S16, S16i, S18i, S21)
- `cnc_rdspmeter` spindle count is capped at 4.
- Contour feed rate from `cnc_actf` is scaled ×1000 on these series (same as 30i).

### 30i / 31i / 32i (S30i, S31i, S32i)
- Full feature set; default strategy applies.
- `cnc_rdexecprog` supports up to 256-byte blocks.
