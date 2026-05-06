package focas

// Mode is the CNC operating mode.
type Mode uint8

const (
	ModeUnknown Mode = iota
	ModeMDI
	ModeMemory
	ModeEdit
	ModeHandle
	ModeJog
	ModeTeach
	ModeIncFeed
	ModeRef
	ModeRemote
)

func (m Mode) String() string {
	switch m {
	case ModeMDI:
		return "MDI"
	case ModeMemory:
		return "MEMORY"
	case ModeEdit:
		return "EDIT"
	case ModeHandle:
		return "HANDLE"
	case ModeJog:
		return "JOG"
	case ModeTeach:
		return "TEACH"
	case ModeIncFeed:
		return "INC_FEED"
	case ModeRef:
		return "REF"
	case ModeRemote:
		return "REMOTE"
	default:
		return "UNKNOWN"
	}
}

// ParseMode parses a mode string (case-insensitive).
func ParseMode(s string) Mode {
	switch s {
	case "MDI", "mdi":
		return ModeMDI
	case "MEMORY", "memory", "AUTO", "AUTOMATIC":
		return ModeMemory
	case "EDIT", "edit":
		return ModeEdit
	case "HANDLE", "handle", "MANUAL_HANDLE":
		return ModeHandle
	case "JOG", "jog":
		return ModeJog
	case "TEACH", "teach":
		return ModeTeach
	case "INC_FEED", "inc_feed", "INCREMENTAL":
		return ModeIncFeed
	case "REF", "ref", "REFERENCE":
		return ModeRef
	case "REMOTE", "remote":
		return ModeRemote
	default:
		return ModeUnknown
	}
}

// RunState is the CNC run state.
type RunState uint8

const (
	RunUnknown RunState = iota
	RunReset
	RunStop
	RunHold
	RunStart
	RunMSTR
)

func (r RunState) String() string {
	switch r {
	case RunReset:
		return "RESET"
	case RunStop:
		return "STOP"
	case RunHold:
		return "HOLD"
	case RunStart:
		return "START"
	case RunMSTR:
		return "MSTR"
	default:
		return "UNKNOWN"
	}
}

// MotionState indicates whether axes are moving.
type MotionState uint8

const (
	MotionIdle MotionState = iota
	MotionMoving
	MotionDwell
)

func (m MotionState) String() string {
	switch m {
	case MotionMoving:
		return "MOVING"
	case MotionDwell:
		return "DWELL"
	default:
		return "IDLE"
	}
}

// EmergencyState reflects the E-stop condition.
type EmergencyState uint8

const (
	EmergencyReleased EmergencyState = iota
	EmergencyTriggered
	EmergencyReset
)

func (e EmergencyState) String() string {
	switch e {
	case EmergencyTriggered:
		return "TRIGGERED"
	case EmergencyReset:
		return "RESET"
	default:
		return "RELEASED"
	}
}

// AlarmState indicates the alarm summary from statinfo.
type AlarmState uint8

const (
	AlarmsNone AlarmState = iota
	AlarmsActive
	AlarmsBatteryLow
	AlarmsFanFault
)

func (a AlarmState) String() string {
	switch a {
	case AlarmsActive:
		return "ACTIVE"
	case AlarmsBatteryLow:
		return "BATTERY_LOW"
	case AlarmsFanFault:
		return "FAN_FAULT"
	default:
		return "NONE"
	}
}

// Status is a point-in-time snapshot of the controller's operational state.
type Status struct {
	Mode      Mode
	Run       RunState
	Motion    MotionState
	MSTB      bool // M/S/T/B-function busy
	Emergency EmergencyState
	Alarm     AlarmState
	Edit      bool
}
