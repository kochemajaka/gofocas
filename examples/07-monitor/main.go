package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kochemajaka/gofocas/series"

	"github.com/kochemajaka/gofocas"
)

const (
	altScreenOn  = "\033[?1049h" // switch to alternate screen buffer
	altScreenOff = "\033[?1049l" // restore primary screen
	cursorHide   = "\033[?25l"
	cursorShow   = "\033[?25h"
	cursorHome   = "\033[H"
	clearScreen  = "\033[2J" // clear entire screen
	clearLine    = "\033[K"  // clear from cursor to end of line
	bold         = "\033[1m"
	dim          = "\033[2m"
	red          = "\033[31m"
	green        = "\033[32m"
	yellow       = "\033[33m"
	cyan         = "\033[36m"
	reset        = "\033[0m"
)

type snapshot struct {
	system    focas.System
	status    focas.Status
	axes      []focas.Axis
	feed      focas.Feed
	program   focas.Program
	alarms    []focas.Alarm
	params    focas.Parameters
	emergency bool
	at        time.Time
	errs      []string
}

func main() {
	addr := os.Getenv("FOCAS_ADDR")
	if addr == "" {
		addr = "192.168.1.1"
	}
	interval := 250 * time.Millisecond
	if v := os.Getenv("FOCAS_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client, err := focas.Dial(ctx, addr,
		focas.WithDialTimeout(5*time.Second),
		focas.WithCallTimeout(8*time.Second),
		focas.WithReconnect(focas.DefaultReconnectPolicy()),
		focas.WithLogPath("/tmp/fanuc.log"),
	)
	if err != nil {
		fmt.Printf("%sdial failed: %v%s\n", red, err, reset)
		os.Exit(1)
	}
	defer client.Close()

	sys, _ := client.System(ctx)

	// Enter alternate screen so the terminal returns to its previous state on exit.
	fmt.Print(altScreenOn + cursorHide)
	defer fmt.Print(cursorShow + altScreenOff)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		s := poll(ctx, client)
		s.system = sys
		render(addr, s)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func poll(ctx context.Context, c *focas.Client) snapshot {
	rCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	var s snapshot
	s.at = time.Now()

	track := func(op string, err error) {
		if err != nil {
			s.errs = append(s.errs, fmt.Sprintf("%s: %v", op, err))
		}
	}

	var err error
	s.status, err = c.Status(rCtx)
	track("status", err)
	s.axes, err = c.Axes(rCtx)
	track("axes", err)
	s.feed, err = c.Feed(rCtx)
	track("feed", err)
	s.program, err = c.ExecutingProgram(rCtx)
	track("program", err)
	s.alarms, err = c.Alarms(rCtx)
	track("alarms", err)
	s.emergency, err = c.EmergencyStop(rCtx)
	track("emergency", err)
	s.params, err = c.Parameters(rCtx)
	track("parameters", err)

	return s
}

// render builds the entire frame in a buffer, then writes it once. Each line is
// terminated with \033[K (clear-to-end-of-line) so previous frames with longer
// content don't leave artifacts.
// frameHeight is the fixed number of lines the frame occupies. Render pads up
// to this many rows so a shorter frame never leaves artifacts from a taller one.
const frameHeight = 40

func render(addr string, s snapshot) {
	var b bytes.Buffer
	// Wipe the whole screen each frame, then redraw from the top-left. This
	// avoids any leftover characters from previous frames or stray output that
	// the FOCAS library may have written outside our buffer.
	b.WriteString(clearScreen + cursorHome)

	lines := 0
	w := func(format string, args ...any) {
		fmt.Fprintf(&b, format, args...)
		b.WriteString(clearLine + "\n")
		lines++
	}

	seriesName := s.system.Series.String()
	if s.system.Series == series.Unknown {
		seriesName = "?"
	}

	w("%s┌─ FANUC %s ─ %s ─ %s%s",
		cyan, seriesName, addr, s.at.Format("15:04:05"), reset)
	w("%s│%s %sversion%s %s   %saxes%s %d   %sport%s 8193",
		cyan, reset, dim, reset, s.system.Version, dim, reset, s.system.Axes, dim, reset)
	w("%s└─%s", cyan, reset)
	w("")

	writeStatus(w, s.status, s.program, s.alarms, s.emergency)
	w("")
	writeAxes(w, s.axes)
	w("")
	writeFeed(w, s.feed)
	w("")
	writeProduction(w, s.params)
	w("")
	writeAlarms(w, s.alarms)

	w("")
	if len(s.errs) > 0 {
		w("%spoll errors:%s", red, reset)
		for _, e := range s.errs {
			w("  %s", e)
		}
	} else {
		w("%spoll: ok%s", dim, reset)
	}

	// Pad to a fixed frame height so taller frames (with errors) don't push
	// content around when they shrink back.
	for lines < frameHeight {
		b.WriteString(clearLine + "\n")
		lines++
	}

	os.Stdout.Write(b.Bytes())
}

type writer func(format string, args ...any)

func writeStatus(w writer, st focas.Status, p focas.Program, alarms []focas.Alarm, emergency bool) {
	emerStr := "RELEASED"
	emerColor := green
	if emergency || hasEmergencyAlarm(alarms) {
		emerStr = "TRIGGERED"
		emerColor = red
	}

	w("%sSTATUS%s", bold, reset)
	w("  mode:      %s%-9s%s", colorMode(st.Mode), st.Mode, reset)
	w("  run:       %s%-9s%s", colorRun(st.Run), st.Run, reset)
	w("  motion:    %s%-9s%s", colorMotion(st.Motion), st.Motion, reset)
	w("  emergency: %s%-9s%s", emerColor, emerStr, reset)
	w("  alarms:    %s%-9s%s", colorAlarmState(st.Alarm), st.Alarm, reset)

	progName := p.Name
	if progName == "" {
		progName = "-"
	}
	gcode := p.GCodeLine
	if gcode == "" {
		gcode = "-"
	}
	w("  program:   %-20s  (O%d)", progName, p.Number)
	w("  gcode:     %s", gcode)
}

func writeAxes(w writer, axes []focas.Axis) {
	w("%sAXES%s", bold, reset)
	if len(axes) == 0 {
		w("  (no data)")
		return
	}
	w("  %-4s %12s", "name", "absolute")
	for _, a := range axes {
		w("  %-4s %12.3f", a.Name, a.Position.Absolute)
	}
}

func writeFeed(w writer, f focas.Feed) {
	w("%sFEED%s", bold, reset)
	w("  override:  %d%%", f.OverridePercent)
}

func writeProduction(w writer, p focas.Parameters) {
	w("%sPRODUCTION%s", bold, reset)
	w("  parts:     %d", p.PartsCount)
	w("  run time:  %s", fmtDuration(p.Operating))
	w("  cycle:     %s", fmtDuration(p.Cycle))
}

func fmtDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	h := int(d / time.Hour)
	m := int((d % time.Hour) / time.Minute)
	s := int((d % time.Minute) / time.Second)
	if h > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func hasEmergencyAlarm(alarms []focas.Alarm) bool {
	for _, a := range alarms {
		if strings.Contains(strings.ToUpper(a.Message), "EMERGENCY") ||
			strings.Contains(strings.ToUpper(a.Message), "EMG") {
			return true
		}
	}
	return false
}

func writeAlarms(w writer, alarms []focas.Alarm) {
	w("%sALARMS%s", bold, reset)
	if len(alarms) == 0 {
		w("  %sno active alarms%s", green, reset)
		return
	}
	for _, a := range alarms {
		axis := ""
		if a.Axis > 0 {
			axis = fmt.Sprintf(" axis=%d", a.Axis)
		}
		msg := strings.TrimSpace(a.Message)
		w("  %s[%s]%s%s  %s", red, a.Code, reset, axis, msg)
	}
}

func colorMode(m focas.Mode) string {
	switch m {
	case focas.ModeMemory, focas.ModeMDI:
		return green
	case focas.ModeEdit:
		return yellow
	}
	return ""
}

func colorRun(r focas.RunState) string {
	switch r {
	case focas.RunStart:
		return green
	case focas.RunHold, focas.RunStop:
		return yellow
	case focas.RunReset:
		return dim
	}
	return ""
}

func colorMotion(m focas.MotionState) string {
	if m == focas.MotionMoving {
		return green
	}
	return dim
}

func colorAlarmState(a focas.AlarmState) string {
	if a == focas.AlarmsNone {
		return green
	}
	return red
}
