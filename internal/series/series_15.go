package series

type series15 struct{ defaultStrategy }

func Series15() Strategy { return series15{} }

func (series15) MaxExecProgBuf() int { return 64 }

func (series15) ProgramSource(r ProgramReader) (string, error) {
	return r.ReadExecProg(64)
}
