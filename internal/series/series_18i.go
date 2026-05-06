package series

type series18i struct{ defaultStrategy }

func Series18i() Strategy { return series18i{} }

func (series18i) MaxExecProgBuf() int { return 128 }

func (series18i) ProgramSource(r ProgramReader) (string, error) {
	return r.ReadExecProg(128)
}
