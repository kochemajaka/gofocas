package focas

import (
	"context"
	"strings"
	"time"

	"github.com/kochemajaka/gofocas/internal/fwlib32"
	iseries "github.com/kochemajaka/gofocas/internal/series"
	"github.com/kochemajaka/gofocas/series"
)

// System reads static controller information via cnc_sysinfo.
func (c *Client) System(ctx context.Context) (System, error) {
	var out System
	err := c.call(ctx, "cnc_sysinfo", func(h fwlib32.Handle) error {
		si, err := c.binder.SysInfo(h)
		if err != nil {
			return err
		}
		rawSeries := strings.TrimRight(string(si.Series[:]), "\x00 ")
		out = System{
			Manufacturer: "FANUC",
			Model:        strings.TrimRight(string(si.MtType[:]), "\x00 "),
			Series:       series.Parse(rawSeries),
			Version:      strings.TrimRight(string(si.Version[:]), "\x00 "),
			MaxAxes:      int(si.MaxAxis),
			Axes:         int(si.Axes),
		}
		return nil
	})
	return out, err
}

// Status reads the current operational state via cnc_statinfo.
func (c *Client) Status(ctx context.Context) (Status, error) {
	var out Status
	err := c.call(ctx, "cnc_statinfo", func(h fwlib32.Handle) error {
		st, err := c.binder.StatInfo(h)
		if err != nil {
			return err
		}
		raw := iseries.RawState{
			Mode:      uint8(st.AutoMode),
			Run:       uint8(st.RunState),
			Motion:    uint8(st.Axis),
			MSTB:      uint8(st.MSTB),
			Emergency: uint8(st.Emer),
			Alarm:     uint8(st.Alarm),
			Edit:      uint8(st.Edit),
		}
		interpreted := c.strategy.InterpretMachineState(raw)
		out = Status{
			Mode:      modeFromRaw(interpreted.ModeRaw),
			Run:       runStateFromRaw(interpreted.RunRaw),
			Motion:    motionFromRaw(interpreted.MotionRaw),
			MSTB:      interpreted.MSTBRaw != 0,
			Emergency: emergencyFromRaw(interpreted.EmergencyRaw),
			Alarm:     alarmStateFromRaw(interpreted.AlarmRaw),
			Edit:      interpreted.EditRaw != 0,
		}
		return nil
	})
	return out, err
}

// Axes reads position, load, and diagnostic data for all controlled axes.
func (c *Client) Axes(ctx context.Context) ([]Axis, error) {
	var out []Axis
	err := c.call(ctx, "cnc_rdposition", func(h fwlib32.Handle) error {
		n := c.maxAxes
		if n == 0 {
			n = 8
		}
		positions, err := c.binder.Positions(h, n)
		if err != nil {
			return err
		}
		out = make([]Axis, len(positions))
		for i, p := range positions {
			unit := UnitMillimeter
			if p.Unit == 1 {
				unit = UnitInch
			}
			// Positions are already scaled by fwlib32.Positions using POSELM.dec.
			out[i] = Axis{
				Index: i + 1,
				Name:  string([]byte{p.Name}),
				Position: Position{
					Absolute: p.Absolute,
					Machine:  p.Machine,
					Relative: p.Relative,
					Distance: p.Distance,
					Unit:     unit,
				},
			}
		}

		// Read diagnostic #301 (servo error pulse) for each axis.
		for i := range out {
			d, err := c.binder.Diag(h, 301, i+1)
			if err == nil {
				out[i].Diag301 = float64(d.Int)
			}
		}
		return nil
	})
	return out, err
}

// Spindles reads speed, load, and power for all spindles.
func (c *Client) Spindles(ctx context.Context) ([]Spindle, error) {
	var out []Spindle
	err := c.call(ctx, "cnc_rdspmeter", func(h fwlib32.Handle) error {
		n := c.cfg.maxSpindles
		meters, err := c.binder.SpindleMeters(h, n)
		if err != nil {
			return err
		}
		loads, _ := c.binder.SpindleLoads(h, n) // best-effort

		out = make([]Spindle, len(meters))
		for i, m := range meters {
			sp := Spindle{
				Index:    i + 1,
				SpeedRPM: int(m.SpeedRPM),
				Load:     float64(m.Load),
			}
			if i < len(loads) {
				sp.PowerW = int(loads[i].PowerW)
			}
			// Read diag #411 (spindle motor current) per spindle.
			if d, err := c.binder.Diag(h, 411, i+1); err == nil {
				sp.Diag411 = int(d.Int)
			}
			out[i] = sp
		}
		return nil
	})
	return out, err
}

// Feed reads the actual feed rate and override.
func (c *Client) Feed(ctx context.Context) (Feed, error) {
	var out Feed
	err := c.call(ctx, "cnc_rdspeed", func(h fwlib32.Handle) error {
		spd, err := c.binder.Speed(h)
		if err != nil {
			return err
		}
		out.ActualMMPerMin = int(spd.ActualFeed)

		ov, err := c.binder.FeedOverride(h)
		if err == nil {
			out.OverridePercent = ov
		}
		return nil
	})
	return out, err
}

// ContourFeedRate returns the current contour feed rate from cnc_actf (×1000).
func (c *Client) ContourFeedRate(ctx context.Context) (int, error) {
	var out int
	err := c.call(ctx, "cnc_actf", func(h fwlib32.Handle) error {
		v, err := c.binder.ActualFeed(h)
		if err != nil {
			return err
		}
		out = int(v)
		return nil
	})
	return out, err
}

// FeedOverride returns the feed override percentage.
func (c *Client) FeedOverride(ctx context.Context) (int, error) {
	var out int
	err := c.call(ctx, "cnc_rdtofs/feed", func(h fwlib32.Handle) error {
		v, err := c.binder.FeedOverride(h)
		if err != nil {
			return err
		}
		out = v
		return nil
	})
	return out, err
}

// JogOverride returns the jog override percentage.
func (c *Client) JogOverride(ctx context.Context) (int, error) {
	var out int
	err := c.call(ctx, "cnc_rdtofs/jog", func(h fwlib32.Handle) error {
		v, err := c.binder.JogOverride(h)
		if err != nil {
			return err
		}
		out = v
		return nil
	})
	return out, err
}

// ExecutingProgram returns the currently executing NC program name and number.
func (c *Client) ExecutingProgram(ctx context.Context) (Program, error) {
	var out Program
	err := c.call(ctx, "cnc_exeprgname", func(h fwlib32.Handle) error {
		ep, err := c.binder.ExecProgName(h)
		if err != nil {
			return err
		}
		out.Name = ep.Name
		out.Number = int64(ep.Number)

		// Read the currently executing G-code block via the series strategy.
		block, err := c.strategy.ProgramSource(progReader{binder: c.binder, handle: h})
		if err == nil {
			out.GCodeLine = block
		}
		return nil
	})
	return out, err
}

// progReader bridges the internal ProgramReader interface to our binder.
type progReader struct {
	binder fwlib32.Binder
	handle fwlib32.Handle
}

func (r progReader) ReadExecProg(bufLen int) (string, error) {
	block, err := r.binder.ReadExecProg(r.handle, bufLen)
	if err != nil {
		return "", err
	}
	return block.Block, nil
}

// ProgramSource returns the raw G-code block being executed.
func (c *Client) ProgramSource(ctx context.Context) (string, error) {
	var out string
	err := c.call(ctx, "cnc_rdexecprog", func(h fwlib32.Handle) error {
		s, err := c.strategy.ProgramSource(progReader{binder: c.binder, handle: h})
		if err != nil {
			return err
		}
		out = s
		return nil
	})
	return out, err
}

// AlarmStatus returns the bit-mask of active alarm categories from cnc_alarm2.
// On 30i/31i/32i, bit 30 indicates an active emergency stop.
func (c *Client) AlarmStatus(ctx context.Context) (uint32, error) {
	var out uint32
	err := c.call(ctx, "cnc_alarm2", func(h fwlib32.Handle) error {
		v, err := c.binder.AlarmStatus(h)
		if err != nil {
			return err
		}
		out = v
		return nil
	})
	return out, err
}

// EmergencyStop reports whether the controller is currently in emergency-stop
// state. It checks the dedicated bit in cnc_alarm2's status mask, falling back
// to the statinfo emergency field if cnc_alarm2 is unavailable.
func (c *Client) EmergencyStop(ctx context.Context) (bool, error) {
	mask, err := c.AlarmStatus(ctx)
	if err == nil {
		// Bit 30 = emergency stop on 30i/31i/32i controllers.
		return mask&(1<<30) != 0, nil
	}
	st, err2 := c.Status(ctx)
	if err2 != nil {
		return false, err
	}
	return st.Emergency == EmergencyTriggered, nil
}

// Alarms returns all active alarms from cnc_rdalmmsg.
func (c *Client) Alarms(ctx context.Context) ([]Alarm, error) {
	var out []Alarm
	err := c.call(ctx, "cnc_rdalmmsg", func(h fwlib32.Handle) error {
		msgs, err := c.binder.AlarmMsgs(h, 32)
		if err != nil {
			return err
		}
		out = make([]Alarm, len(msgs))
		for i, m := range msgs {
			t := alarmTypeFromRaw(int(m.Type))
			out[i] = Alarm{
				Code:    makeAlarmCode(t, int(m.AlmNo)),
				Number:  int(m.AlmNo),
				Type:    t,
				Axis:    int(m.Axis),
				Message: m.Msg,
			}
		}
		return nil
	})
	return out, err
}

// Parameters reads production counters and runtime durations.
func (c *Client) Parameters(ctx context.Context) (Parameters, error) {
	var out Parameters
	err := c.call(ctx, "cnc_rdparam/production", func(h fwlib32.Handle) error {
		pp, err := c.binder.ProductionParams(h)
		if err != nil {
			return err
		}
		out = Parameters{
			PartsCount: pp.PartsCount,
			PowerOn:    pp.PowerOn,
			Operating:  pp.Operating,
			Cutting:    pp.Cutting,
			Cycle:      pp.Cycle,
		}
		return nil
	})
	return out, err
}

// Parameter reads a single parameter by number and axis.
func (c *Client) Parameter(ctx context.Context, no, axis int) (ParamValue, error) {
	var out ParamValue
	err := c.call(ctx, "cnc_rdparam", func(h fwlib32.Handle) error {
		pv, err := c.binder.Param(h, no, axis)
		if err != nil {
			return err
		}
		out = ParamValue{
			No:   int(pv.No),
			Axis: int(pv.Axis),
			Kind: paramKindFromRaw(int(pv.Kind)),
			Int:  pv.Int,
			Real: pv.Real,
			Bits: pv.Bits,
		}
		return nil
	})
	return out, err
}

// Diagnosis reads a single diagnosis data item by number and axis.
func (c *Client) Diagnosis(ctx context.Context, no, axis int) (DiagValue, error) {
	var out DiagValue
	err := c.call(ctx, "cnc_diagnoss", func(h fwlib32.Handle) error {
		dv, err := c.binder.Diag(h, no, axis)
		if err != nil {
			return err
		}
		out = DiagValue{
			No:   int(dv.No),
			Axis: int(dv.Axis),
			Kind: paramKindFromRaw(int(dv.Kind)),
			Int:  dv.Int,
			Real: dv.Real,
			Bits: dv.Bits,
		}
		return nil
	})
	return out, err
}

// --- mapping helpers ---

func modeFromRaw(v uint8) Mode {
	// FOCAS statinfo autoact / tmmode values per FANUC FOCAS2 manual.
	switch v {
	case 1:
		return ModeMDI
	case 2:
		return ModeTeach
	case 3:
		return ModeTeach
	case 4:
		return ModeHandle
	case 5:
		return ModeJog
	case 6:
		return ModeIncFeed
	case 7:
		return ModeRef
	case 8:
		return ModeMemory
	case 9:
		return ModeRemote
	case 10:
		return ModeEdit
	default:
		return ModeUnknown
	}
}

func runStateFromRaw(v uint8) RunState {
	switch v {
	case 0:
		return RunReset
	case 1:
		return RunStop
	case 2:
		return RunHold
	case 3:
		return RunStart
	case 4:
		return RunMSTR
	default:
		return RunUnknown
	}
}

func motionFromRaw(v uint8) MotionState {
	switch v {
	case 1:
		return MotionMoving
	case 2:
		return MotionDwell
	default:
		return MotionIdle
	}
}

func emergencyFromRaw(v uint8) EmergencyState {
	switch v {
	case 1:
		return EmergencyTriggered
	case 2:
		return EmergencyReset
	default:
		return EmergencyReleased
	}
}

func alarmStateFromRaw(v uint8) AlarmState {
	switch v {
	case 1:
		return AlarmsActive
	case 2:
		return AlarmsBatteryLow
	case 3:
		return AlarmsFanFault
	default:
		return AlarmsNone
	}
}

func paramKindFromRaw(v int) ParamKind {
	switch v {
	case 0:
		return ParamBit
	case 1:
		return ParamByte
	case 2:
		return ParamWord
	case 3:
		return ParamLong
	case 4:
		return ParamReal
	default:
		return ParamLong
	}
}

// ensure time import is used (Parameters.PowerOn etc. use time.Duration).
var _ = time.Second
