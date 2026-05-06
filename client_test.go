package focas

import (
	"context"
	"testing"
	"time"

	"github.com/kochemajaka/gofocas/internal/fwlib32"
	iseries "github.com/kochemajaka/gofocas/internal/series"
	"github.com/kochemajaka/gofocas/series"
)

// fakeBinder is a fully-controllable test double for fwlib32.Binder.
type fakeBinder struct {
	sysInfo        fwlib32.SysInfo
	statInfo       fwlib32.StatInfo
	positions      []fwlib32.AxisPosition
	spindleMeters  []fwlib32.SpindleMeter
	spindleLoads   []fwlib32.SpindleLoad
	actualFeed     int32
	speed          fwlib32.Speed
	feedOverride   int
	jogOverride    int
	execProg       fwlib32.ExecProg
	execProgBlock  fwlib32.ExecProgBlock
	alarmMsgs      []fwlib32.AlarmMsg
	diag           fwlib32.DiagValue
	param          fwlib32.ParamValue
	productionParams fwlib32.ProductionParams
	err            error
}

func (f *fakeBinder) Startup(_ string) error                                     { return f.err }
func (f *fakeBinder) Alloc(_ string, _ uint16, _ uint32) (fwlib32.Handle, error) { return 1, f.err }
func (f *fakeBinder) Free(_ fwlib32.Handle) error                                { return f.err }
func (f *fakeBinder) SysInfo(_ fwlib32.Handle) (fwlib32.SysInfo, error)          { return f.sysInfo, f.err }
func (f *fakeBinder) StatInfo(_ fwlib32.Handle) (fwlib32.StatInfo, error)        { return f.statInfo, f.err }
func (f *fakeBinder) Positions(_ fwlib32.Handle, _ int) ([]fwlib32.AxisPosition, error) {
	return f.positions, f.err
}
func (f *fakeBinder) SpindleMeters(_ fwlib32.Handle, _ int) ([]fwlib32.SpindleMeter, error) {
	return f.spindleMeters, f.err
}
func (f *fakeBinder) SpindleLoads(_ fwlib32.Handle, _ int) ([]fwlib32.SpindleLoad, error) {
	return f.spindleLoads, f.err
}
func (f *fakeBinder) ActualFeed(_ fwlib32.Handle) (int32, error)        { return f.actualFeed, f.err }
func (f *fakeBinder) Speed(_ fwlib32.Handle) (fwlib32.Speed, error)     { return f.speed, f.err }
func (f *fakeBinder) FeedOverride(_ fwlib32.Handle) (int, error)        { return f.feedOverride, f.err }
func (f *fakeBinder) JogOverride(_ fwlib32.Handle) (int, error)         { return f.jogOverride, f.err }
func (f *fakeBinder) ExecProgName(_ fwlib32.Handle) (fwlib32.ExecProg, error) {
	return f.execProg, f.err
}
func (f *fakeBinder) ReadExecProg(_ fwlib32.Handle, _ int) (fwlib32.ExecProgBlock, error) {
	return f.execProgBlock, f.err
}
func (f *fakeBinder) AlarmMsgs(_ fwlib32.Handle, _ int) ([]fwlib32.AlarmMsg, error) {
	return f.alarmMsgs, f.err
}
func (f *fakeBinder) Diag(_ fwlib32.Handle, _, _ int) (fwlib32.DiagValue, error) {
	return f.diag, f.err
}
func (f *fakeBinder) Param(_ fwlib32.Handle, _, _ int) (fwlib32.ParamValue, error) {
	return f.param, f.err
}
func (f *fakeBinder) ProductionParams(_ fwlib32.Handle) (fwlib32.ProductionParams, error) {
	return f.productionParams, f.err
}

// newTestClient builds a Client bypassing Dial (no network needed).
func newTestClient(b fwlib32.Binder) *Client {
	return &Client{
		handle:   1,
		addr:     "192.168.1.1:8193",
		host:     "192.168.1.1",
		cfg:      defaultConfig(),
		binder:   b,
		strategy: iseries.Default(),
		series:   series.S30i,
		maxAxes:  3,
	}
}

func TestStatus(t *testing.T) {
	b := &fakeBinder{
		statInfo: fwlib32.StatInfo{
			AutoMode: 8, // memory
			RunState: 3, // start
			Axis:     1, // moving
		},
	}
	c := newTestClient(b)
	st, err := c.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode != ModeMemory {
		t.Errorf("Mode: got %v, want ModeMemory", st.Mode)
	}
	if st.Run != RunStart {
		t.Errorf("Run: got %v, want RunStart", st.Run)
	}
	if st.Motion != MotionMoving {
		t.Errorf("Motion: got %v, want MotionMoving", st.Motion)
	}
}

func TestAxes(t *testing.T) {
	b := &fakeBinder{
		positions: []fwlib32.AxisPosition{
			{Absolute: 100000, Machine: 90000, Relative: 10000, Distance: 0, Unit: 0, Name: 'X'},
			{Absolute: 50000, Machine: 45000, Relative: 5000, Distance: 1000, Unit: 0, Name: 'Y'},
		},
	}
	c := newTestClient(b)
	c.maxAxes = 2

	axes, err := c.Axes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(axes) != 2 {
		t.Fatalf("want 2 axes, got %d", len(axes))
	}
	if axes[0].Name != "X" {
		t.Errorf("axis[0].Name: got %q, want X", axes[0].Name)
	}
	// 100000 * 0.001 = 100.0
	if axes[0].Position.Absolute != 100.0 {
		t.Errorf("axis[0].Position.Absolute: got %v, want 100.0", axes[0].Position.Absolute)
	}
}

func TestFeed(t *testing.T) {
	b := &fakeBinder{
		speed:        fwlib32.Speed{ActualFeed: 500},
		feedOverride: 80,
	}
	c := newTestClient(b)
	f, err := c.Feed(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if f.ActualMMPerMin != 500 {
		t.Errorf("ActualMMPerMin: got %d, want 500", f.ActualMMPerMin)
	}
	if f.OverridePercent != 80 {
		t.Errorf("OverridePercent: got %d, want 80", f.OverridePercent)
	}
}

func TestAlarms(t *testing.T) {
	b := &fakeBinder{
		alarmMsgs: []fwlib32.AlarmMsg{
			{AlmNo: 401, Type: 7, Axis: 1, Msg: "SERVO ALARM"},
		},
	}
	c := newTestClient(b)
	alarms, err := c.Alarms(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(alarms) != 1 {
		t.Fatalf("want 1 alarm, got %d", len(alarms))
	}
	if alarms[0].Type != AlarmTypeServo {
		t.Errorf("Type: got %v, want AlarmTypeServo", alarms[0].Type)
	}
	if alarms[0].Message != "SERVO ALARM" {
		t.Errorf("Message: got %q", alarms[0].Message)
	}
}

func TestParameters(t *testing.T) {
	b := &fakeBinder{
		productionParams: fwlib32.ProductionParams{
			PartsCount: 42,
			PowerOn:    2 * time.Hour,
			Operating:  1 * time.Hour,
			Cutting:    30 * time.Minute,
			Cycle:      45 * time.Second,
		},
	}
	c := newTestClient(b)
	p, err := c.Parameters(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p.PartsCount != 42 {
		t.Errorf("PartsCount: got %d, want 42", p.PartsCount)
	}
	if p.PowerOn != 2*time.Hour {
		t.Errorf("PowerOn: got %v, want 2h", p.PowerOn)
	}
}

func TestClosedClient(t *testing.T) {
	c := newTestClient(&fakeBinder{})
	c.closed = true
	_, err := c.Status(context.Background())
	if err != ErrClosed {
		t.Errorf("want ErrClosed, got %v", err)
	}
}

func TestSplitAddr(t *testing.T) {
	cases := []struct {
		in       string
		defPort  uint16
		wantHost string
		wantPort uint16
	}{
		{"192.168.1.1", 8193, "192.168.1.1", 8193},
		{"192.168.1.1:9000", 8193, "192.168.1.1", 9000},
		{"cnc.local", 8193, "cnc.local", 8193},
	}
	for _, tc := range cases {
		h, p, err := splitAddr(tc.in, tc.defPort)
		if err != nil {
			t.Errorf("splitAddr(%q): unexpected error %v", tc.in, err)
			continue
		}
		if h != tc.wantHost || p != tc.wantPort {
			t.Errorf("splitAddr(%q) = (%q, %d), want (%q, %d)", tc.in, h, p, tc.wantHost, tc.wantPort)
		}
	}
}
