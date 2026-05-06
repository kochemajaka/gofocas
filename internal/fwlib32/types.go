// Package fwlib32 is the only place in gofocas that knows about FOCAS C types.
// All exported symbols are plain Go; no C types escape this package.
package fwlib32

import (
	"fmt"
	"time"
)

// FocasError carries the raw EW_* return code from an fwlib32 call.
type FocasError struct {
	Op   string
	Code int
}

func (e *FocasError) Error() string {
	return fmt.Sprintf("focas: %s: EW_%d", e.Op, e.Code)
}

// Handle is an opaque FOCAS library handle (uint16 on the wire).
type Handle uint16

// SysInfo mirrors the relevant fields from ODBSYS returned by cnc_sysinfo.
type SysInfo struct {
	AddInfo  uint16
	MaxAxis  int16
	CNCType  [2]byte
	MtType   [2]byte
	Series   [4]byte
	Version  [4]byte
	Axes     int16
}

// StatInfo mirrors the relevant fields from ODBST returned by cnc_statinfo.
type StatInfo struct {
	Dummy    [2]int16
	TMMode   int16
	AutoMode int16
	RunState int16
	Axis     int16
	Edit     int16
	MSTB     int16
	Emer     int16
	Alarm    int16
	Prgram   int16
}

// Position mirrors cnc_rdposition output for a single axis.
type AxisPosition struct {
	Absolute float64
	Machine  float64
	Relative float64
	Distance float64
	Unit     int16 // 0=mm 1=inch
	Name     byte
	Suf      byte
}

// SpindleMeter mirrors cnc_rdspmeter for one spindle.
type SpindleMeter struct {
	SpeedRPM int32
	Load     int32 // %
}

// SpindleLoad is the load percentage from cnc_rdspload.
type SpindleLoad struct {
	Load    int32
	PowerW  int32
}

// AlarmMsg mirrors one entry from cnc_rdalmmsg output.
type AlarmMsg struct {
	AlmNo  int32
	Type   int16
	Axis   int16
	MsgLen int16
	Msg    string
}

// ExecProg is the result of cnc_exeprgname.
type ExecProg struct {
	Name   string
	Number int32
}

// ExecProgBlock is the result of cnc_rdexecprog (one G-code block).
type ExecProgBlock struct {
	Block string
}

// Speed mirrors cnc_rdspeed.
type Speed struct {
	ActualFeed int32
	JogFeed    int32
}

// DiagValue is the result of cnc_diagnoss for one entry.
type DiagValue struct {
	No   int16
	Axis int16
	Kind int16 // 0=bit 1=byte 2=word 3=long 4=real
	Int  int64
	Real float64
	Bits uint8
}

// ParamValue is the result of cnc_rdparam for one entry.
type ParamValue struct {
	No   int16
	Axis int16
	Kind int16
	Int  int64
	Real float64
	Bits uint8
}

// ProductionParams is the result of reading params #6711/6750/6751/6753/6757.
type ProductionParams struct {
	PartsCount int64
	PowerOn    time.Duration
	Operating  time.Duration
	Cutting    time.Duration
	Cycle      time.Duration
}

// Binder is the interface that both the real CGo binding and the stub satisfy.
// The Client uses this interface exclusively — it never calls fwlib32 directly.
type Binder interface {
	// Startup must be called once before any handle is opened.
	Startup(logPath string) error

	// Alloc opens a FOCAS handle to the given host:port within timeout ms.
	Alloc(host string, port uint16, timeoutMs uint32) (Handle, error)

	// Free closes a handle.
	Free(h Handle) error

	// SysInfo returns cnc_sysinfo data.
	SysInfo(h Handle) (SysInfo, error)

	// StatInfo returns cnc_statinfo data.
	StatInfo(h Handle) (StatInfo, error)

	// Positions returns cnc_rdposition for axes 0..n-1.
	Positions(h Handle, n int) ([]AxisPosition, error)

	// SpindleMeters returns cnc_rdspmeter for spindles 0..n-1.
	SpindleMeters(h Handle, n int) ([]SpindleMeter, error)

	// SpindleLoads returns cnc_rdspload for spindles 0..n-1.
	SpindleLoads(h Handle, n int) ([]SpindleLoad, error)

	// ActualFeed returns cnc_actf (contour feed rate * 1000).
	ActualFeed(h Handle) (int32, error)

	// Speed returns cnc_rdspeed.
	Speed(h Handle) (Speed, error)

	// FeedOverride returns the feed override percent from cnc_rdtofs.
	FeedOverride(h Handle) (int, error)

	// JogOverride returns the jog override percent from cnc_rdtofs.
	JogOverride(h Handle) (int, error)

	// ExecProgName returns cnc_exeprgname.
	ExecProgName(h Handle) (ExecProg, error)

	// ReadExecProg returns cnc_rdexecprog with bufLen capacity.
	ReadExecProg(h Handle, bufLen int) (ExecProgBlock, error)

	// AlarmMsgs returns cnc_rdalmmsg, up to max alarms.
	AlarmMsgs(h Handle, max int) ([]AlarmMsg, error)

	// AlarmStatus returns the bit-mask of active alarm categories from
	// cnc_alarm2. Bit 30 = emergency stop on 30i/31i/32i.
	AlarmStatus(h Handle) (uint32, error)

	// Diag returns cnc_diagnoss for a single no/axis pair.
	Diag(h Handle, no, axis int) (DiagValue, error)

	// Param returns cnc_rdparam for a single no/axis pair.
	Param(h Handle, no, axis int) (ParamValue, error)

	// ProductionParams reads params #6711, #6750, #6751, #6753, #6757.
	ProductionParams(h Handle) (ProductionParams, error)
}
