// Package series provides per-model strategy implementations for interpreting
// raw FOCAS data that varies across controller families.
package series

// RawState is the raw output from cnc_statinfo, passed to the strategy.
type RawState struct {
	Mode      uint8
	Run       uint8
	Motion    uint8
	MSTB      uint8
	Emergency uint8
	Alarm     uint8
	Edit      uint8
}

// ProgramReader abstracts the cnc_rdexecprog call so strategies can vary the
// buffer handling without depending on CGo types.
type ProgramReader interface {
	ReadExecProg(bufLen int) (string, error)
}

// InterpretedState is the result of applying a strategy to a RawState.
type InterpretedState struct {
	ModeRaw      uint8
	RunRaw       uint8
	MotionRaw    uint8
	MSTBRaw      uint8
	EmergencyRaw uint8
	AlarmRaw     uint8
	EditRaw      uint8
}

// Strategy is the interface that per-series implementations satisfy.
type Strategy interface {
	// InterpretMachineState maps raw cnc_statinfo bytes to our typed enums.
	// Returns the same RawState enriched/normalised for the series.
	InterpretMachineState(raw RawState) InterpretedState

	// ProgramSource reads the currently executing program source via r.
	// Some older series need a smaller buffer or a different call sequence.
	ProgramSource(r ProgramReader) (string, error)

	// MaxExecProgBuf is the maximum buffer size for cnc_rdexecprog on this series.
	MaxExecProgBuf() int
}
