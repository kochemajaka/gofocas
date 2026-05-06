// Package series defines constants for FANUC CNC controller series.
package series

import "strings"

// Series identifies the FANUC controller family.
type Series uint8

const (
	Unknown Series = iota
	S0i
	S15
	S15i
	S16
	S16i
	S18i
	S21
	S30i
	S31i
	S32i
)

func (s Series) String() string {
	switch s {
	case S0i:
		return "0i"
	case S15:
		return "15"
	case S15i:
		return "15i"
	case S16:
		return "16"
	case S16i:
		return "16i"
	case S18i:
		return "18i"
	case S21:
		return "21"
	case S30i:
		return "30i"
	case S31i:
		return "31i"
	case S32i:
		return "32i"
	default:
		return "unknown"
	}
}

// Parse converts a string like "0i", "30I", "31", "31i" to Series.
// Returns Unknown for unrecognised input.
func Parse(s string) Series {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "0i", "0":
		return S0i
	case "15":
		return S15
	case "15i":
		return S15i
	case "16":
		return S16
	case "16i":
		return S16i
	case "18i", "18":
		return S18i
	case "21":
		return S21
	case "30i", "30":
		return S30i
	case "31i", "31", "g433": // G433 is how FANUC 31i identifies itself in ODBSYS.series
		return S31i
	case "32i", "32":
		return S32i
	default:
		return Unknown
	}
}
