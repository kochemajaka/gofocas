package series

type series15i struct{ defaultStrategy }

func Series15i() Strategy { return series15i{} }

func (series15i) MaxExecProgBuf() int { return 128 }

func (series15i) ProgramSource(r ProgramReader) (string, error) {
	return r.ReadExecProg(128)
}
