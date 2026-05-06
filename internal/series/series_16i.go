package series

type series16i struct{ defaultStrategy }

func Series16i() Strategy { return series16i{} }

func (series16i) MaxExecProgBuf() int { return 128 }

func (series16i) ProgramSource(r ProgramReader) (string, error) {
	return r.ReadExecProg(128)
}
