package focas

import "strconv"

// AlarmType classifies a FOCAS alarm by subsystem.
type AlarmType uint8

const (
	AlarmTypeUnknown AlarmType = iota
	AlarmTypeParam
	AlarmTypePulse
	AlarmTypeIO
	AlarmTypeMacro
	AlarmTypeSpindle
	AlarmTypeOvertravel
	AlarmTypeOverheat
	AlarmTypeServo
	AlarmTypeDataIO
	AlarmTypePMC
	AlarmTypeOT
	AlarmTypeExternal
)

func (t AlarmType) String() string {
	switch t {
	case AlarmTypeParam:
		return "PARAM"
	case AlarmTypePulse:
		return "PULSE"
	case AlarmTypeIO:
		return "IO"
	case AlarmTypeMacro:
		return "MACRO"
	case AlarmTypeSpindle:
		return "SPINDLE"
	case AlarmTypeOvertravel:
		return "OVERTRAVEL"
	case AlarmTypeOverheat:
		return "OVERHEAT"
	case AlarmTypeServo:
		return "SERVO"
	case AlarmTypeDataIO:
		return "DATA_IO"
	case AlarmTypePMC:
		return "PMC"
	case AlarmTypeOT:
		return "OT"
	case AlarmTypeExternal:
		return "EXTERNAL"
	default:
		return "UNKNOWN"
	}
}

// alarmTypeFromRaw maps the FOCAS alarm type nibble to AlarmType.
// Values follow FANUC FOCAS2 documentation (Table of alarm types).
func alarmTypeFromRaw(raw int) AlarmType {
	switch raw {
	case 0:
		return AlarmTypeParam
	case 1:
		return AlarmTypePulse
	case 2:
		return AlarmTypeIO
	case 3:
		return AlarmTypeMacro
	case 4:
		return AlarmTypeSpindle
	case 5:
		return AlarmTypeOvertravel
	case 6:
		return AlarmTypeOverheat
	case 7:
		return AlarmTypeServo
	case 8:
		return AlarmTypeDataIO
	case 9:
		return AlarmTypePMC
	case 10:
		return AlarmTypeOT
	case 11:
		return AlarmTypeExternal
	default:
		return AlarmTypeUnknown
	}
}

// Alarm represents a single active alarm from cnc_rdalmmsg.
type Alarm struct {
	Code    string    // composite "TYPE:NUMBER", e.g. "SERVO:401"
	Number  int
	Type    AlarmType
	Axis    int    // 0 if not axis-specific
	Message string
}

func makeAlarmCode(t AlarmType, number int) string {
	return t.String() + ":" + strconv.Itoa(number)
}
