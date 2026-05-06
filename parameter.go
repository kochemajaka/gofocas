package focas

import "time"

// ParamKind describes the data type of a FOCAS parameter or diagnosis value.
type ParamKind uint8

const (
	ParamBit  ParamKind = iota
	ParamByte
	ParamWord
	ParamLong
	ParamReal
)

func (k ParamKind) String() string {
	switch k {
	case ParamBit:
		return "BIT"
	case ParamByte:
		return "BYTE"
	case ParamWord:
		return "WORD"
	case ParamLong:
		return "LONG"
	case ParamReal:
		return "REAL"
	default:
		return "UNKNOWN"
	}
}

// Parameters holds the production counters and runtime totals read via cnc_rdparam.
// Parameter numbers follow standard FANUC assignments:
//   #6711 — parts count
//   #6750 — power-on time (minutes)
//   #6751 — operating time (minutes)
//   #6753 — cutting time (minutes)
//   #6757 — cycle time (milliseconds)
type Parameters struct {
	PartsCount int64
	PowerOn    time.Duration
	Operating  time.Duration
	Cutting    time.Duration
	Cycle      time.Duration
}

// ParamValue is the result of a generic cnc_rdparam call.
type ParamValue struct {
	No   int
	Axis int
	Kind ParamKind
	Int  int64
	Real float64
	Bits uint8
}

// DiagValue is the result of a generic cnc_diagnoss call.
type DiagValue struct {
	No   int
	Axis int
	Kind ParamKind
	Int  int64
	Real float64
	Bits uint8
}
