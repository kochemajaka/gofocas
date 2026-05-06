package series

// defaultStrategy is used for series 30i/31i/32i and as a base for others.
type defaultStrategy struct{}

func Default() Strategy { return defaultStrategy{} }

func (defaultStrategy) InterpretMachineState(raw RawState) InterpretedState {
	return InterpretedState{
		ModeRaw:      raw.Mode,
		RunRaw:       raw.Run,
		MotionRaw:    raw.Motion,
		MSTBRaw:      raw.MSTB,
		EmergencyRaw: raw.Emergency,
		AlarmRaw:     raw.Alarm,
		EditRaw:      raw.Edit,
	}
}

func (defaultStrategy) ProgramSource(r ProgramReader) (string, error) {
	return r.ReadExecProg(defaultStrategy{}.MaxExecProgBuf())
}

func (defaultStrategy) MaxExecProgBuf() int { return 256 }
