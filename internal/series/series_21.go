package series

type series21 struct{ defaultStrategy }

func Series21() Strategy { return series21{} }

func (series21) MaxExecProgBuf() int { return 128 }

func (series21) ProgramSource(r ProgramReader) (string, error) {
	return r.ReadExecProg(128)
}
