package series_test

import (
	"testing"

	iseries "github.com/kochemajaka/gofocas/internal/series"
	pub "github.com/kochemajaka/gofocas/series"
)

func TestDefaultStrategy_InterpretMachineState(t *testing.T) {
	s := iseries.Default()
	raw := iseries.RawState{Mode: 8, Run: 3, Motion: 1, MSTB: 0, Emergency: 0, Alarm: 0, Edit: 0}
	got := s.InterpretMachineState(raw)
	if got.ModeRaw != 8 {
		t.Errorf("ModeRaw: want 8, got %d", got.ModeRaw)
	}
	if got.RunRaw != 3 {
		t.Errorf("RunRaw: want 3, got %d", got.RunRaw)
	}
}

func TestFactory(t *testing.T) {
	cases := []pub.Series{
		pub.S0i, pub.S15, pub.S15i, pub.S16, pub.S16i,
		pub.S18i, pub.S21, pub.S30i, pub.S31i, pub.S32i,
		pub.Unknown,
	}
	for _, s := range cases {
		st := iseries.For(s)
		if st == nil {
			t.Errorf("For(%v) returned nil", s)
		}
	}
}

func TestMaxExecProgBuf(t *testing.T) {
	if iseries.Series0i().MaxExecProgBuf() != 128 {
		t.Error("0i should have 128-byte exec prog buffer")
	}
	if iseries.Series30i().MaxExecProgBuf() != 256 {
		t.Error("30i should have 256-byte exec prog buffer")
	}
}

// fakeReader is a test implementation of ProgramReader.
type fakeReader struct{ data string }

func (f fakeReader) ReadExecProg(bufLen int) (string, error) {
	if len(f.data) > bufLen {
		return f.data[:bufLen], nil
	}
	return f.data, nil
}

func TestProgramSource(t *testing.T) {
	s := iseries.Default()
	r := fakeReader{data: "G01 X100 F200"}
	got, err := s.ProgramSource(r)
	if err != nil {
		t.Fatal(err)
	}
	if got != "G01 X100 F200" {
		t.Errorf("got %q", got)
	}
}
