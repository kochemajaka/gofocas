package series

type series16 struct{ defaultStrategy }

func Series16() Strategy { return series16{} }

func (series16) MaxExecProgBuf() int { return 128 }

func (series16) ProgramSource(r ProgramReader) (string, error) {
	return r.ReadExecProg(128)
}
