package series

// series0i handles differences specific to the FANUC 0i family.
// The mode byte encoding on 0i differs from the 30i/31i default.
type series0i struct{ defaultStrategy }

func Series0i() Strategy { return series0i{} }

// 0i uses a narrower exec-prog buffer due to older Ethernet stack limitations.
func (series0i) MaxExecProgBuf() int { return 128 }

func (series0i) ProgramSource(r ProgramReader) (string, error) {
	return r.ReadExecProg(128)
}
