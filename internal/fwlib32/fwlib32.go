//go:build focas_cgo

package fwlib32

/*
#cgo CFLAGS: -I.
#cgo linux LDFLAGS: -L. -lfwlib32 -lm
#cgo windows LDFLAGS: -L. -lFwlib32
#include "helpers.h"
#include <stdlib.h>
#include <string.h>
#ifdef _WIN32
#include <sync.h>
#endif
*/
import "C"

import (
	"math"
	"sync"
	"time"
	"unsafe"
)

// once guards cnc_startupprocess so it runs exactly once per process.
var once sync.Once

// mu serialises all FOCAS calls — fwlib32 is not thread-safe for concurrent
// calls on the same handle, and even different handles share process-wide state
// on some platforms.
var mu sync.Mutex

type cgoBinding struct{}

// New returns the real CGo-backed Binder.
func New() Binder { return cgoBinding{} }

func (cgoBinding) Startup(logPath string) error {
	var retErr error
	once.Do(func() {
		cPath := C.CString(logPath)
		defer C.free(unsafe.Pointer(cPath))
		ret := C.cnc_startupprocess(0, cPath)
		if ret != C.EW_OK {
			retErr = focasErr("cnc_startupprocess", int(ret))
		}
	})
	return retErr
}

func (cgoBinding) Alloc(host string, port uint16, timeoutMs uint32) (Handle, error) {
	mu.Lock()
	defer mu.Unlock()

	cHost := C.CString(host)
	defer C.free(unsafe.Pointer(cHost))

	var h C.ushort
	ret := C.cnc_allclibhndl3(cHost, C.ushort(port), C.long(timeoutMs/1000), &h)
	if ret != C.EW_OK {
		return 0, focasErr("cnc_allclibhndl3", int(ret))
	}
	return Handle(h), nil
}

func (cgoBinding) Free(h Handle) error {
	mu.Lock()
	defer mu.Unlock()
	ret := C.cnc_freelibhndl(C.ushort(h))
	if ret != C.EW_OK {
		return focasErr("cnc_freelibhndl", int(ret))
	}
	return nil
}

func (cgoBinding) SysInfo(h Handle) (SysInfo, error) {
	mu.Lock()
	defer mu.Unlock()

	var odbsys C.ODBSYS
	ret := C.cnc_sysinfo(C.ushort(h), &odbsys)
	if ret != C.EW_OK {
		return SysInfo{}, focasErr("cnc_sysinfo", int(ret))
	}

	var si SysInfo
	si.AddInfo = uint16(C.focas_sys_addinfo(&odbsys))
	si.MaxAxis = int16(C.focas_sys_maxaxis(&odbsys))
	si.CNCType[0] = byte(C.focas_sys_cnc0(&odbsys))
	si.CNCType[1] = byte(C.focas_sys_cnc1(&odbsys))
	si.MtType[0] = byte(C.focas_sys_mt0(&odbsys))
	si.MtType[1] = byte(C.focas_sys_mt1(&odbsys))
	si.Axes = int16(C.focas_sys_axes(&odbsys))

	// Series and Version are fixed-length char arrays in ODBSYS.
	for i := 0; i < 4; i++ {
		si.Series[i] = byte(odbsys.series[i])
		si.Version[i] = byte(odbsys.version[i])
	}

	return si, nil
}

func (cgoBinding) StatInfo(h Handle) (StatInfo, error) {
	mu.Lock()
	defer mu.Unlock()

	var odbst C.ODBST
	ret := C.cnc_statinfo(C.ushort(h), &odbst)
	if ret != C.EW_OK {
		return StatInfo{}, focasErr("cnc_statinfo", int(ret))
	}

	return StatInfo{
		TMMode:   int16(C.focas_st_tmmode(&odbst)),
		AutoMode: int16(C.focas_st_autoact(&odbst)),
		RunState: int16(C.focas_st_runstate(&odbst)),
		Axis:     int16(C.focas_st_axis(&odbst)),
		Edit:     int16(C.focas_st_edit(&odbst)),
		MSTB:     int16(C.focas_st_mstb(&odbst)),
		Emer:     int16(C.focas_st_emer(&odbst)),
		Alarm:    int16(C.focas_st_alarm(&odbst)),
	}, nil
}

func (cgoBinding) Positions(h Handle, n int) ([]AxisPosition, error) {
	mu.Lock()
	defer mu.Unlock()

	// cnc_rdposition fills an array of ODBPOS (each 48 bytes for this SDK version).
	// We read raw bytes to avoid CGo union/struct alignment issues, matching
	// the layout: POSELM abs, mach, rel, dist — each 12 bytes:
	//   [0..3]  data  long
	//   [4..5]  dec   short
	//   [6..7]  unit  short
	//   [8..9]  disp  short
	//   [10]    name  char
	//   [11]    suff  char
	const poselm = 12
	const odbposSize = poselm * 4 // abs + mach + rel + dist
	buf := make([]byte, odbposSize*96)
	nn := C.short(n)
	ret := C.cnc_rdposition(C.ushort(h), -1, &nn, (*C.ODBPOS)(unsafe.Pointer(&buf[0])))
	if ret != C.EW_OK {
		return nil, focasErr("cnc_rdposition", int(ret))
	}
	n = int(nn)

	axes := make([]AxisPosition, n)
	for i := 0; i < n; i++ {
		base := i * odbposSize
		readLong := func(off int) int32 {
			return int32(buf[base+off]) | int32(buf[base+off+1])<<8 |
				int32(buf[base+off+2])<<16 | int32(buf[base+off+3])<<24
		}
		readShort := func(off int) int16 {
			return int16(buf[base+off]) | int16(buf[base+off+1])<<8
		}
		dec := int(readShort(4))
		if dec < 0 || dec > 9 {
			dec = 3
		}
		scale := math.Pow(10, float64(-dec))
		axes[i] = AxisPosition{
			Absolute: float64(readLong(0)) * scale,
			Machine:  float64(readLong(poselm)) * scale,
			Relative: float64(readLong(poselm*2)) * scale,
			Distance: float64(readLong(poselm*3)) * scale,
			Unit:     int16(readShort(6)),
			Name:     buf[base+10],
			Suf:      buf[base+11],
		}
	}
	return axes, nil
}

func (cgoBinding) SpindleMeters(h Handle, n int) ([]SpindleMeter, error) {
	mu.Lock()
	defer mu.Unlock()

	var buf [8]C.ODBSPLOAD
	cn := C.short(n)
	ret := C.cnc_rdspmeter(C.ushort(h), 0, &cn, &buf[0])
	if ret != C.EW_OK {
		return nil, focasErr("cnc_rdspmeter", int(ret))
	}
	n = int(cn)

	out := make([]SpindleMeter, n)
	for i := 0; i < n; i++ {
		out[i] = SpindleMeter{
			SpeedRPM: int32(buf[i].spspeed.data),
			Load:     int32(buf[i].spload.data),
		}
	}
	return out, nil
}

func (cgoBinding) SpindleLoads(h Handle, n int) ([]SpindleLoad, error) {
	mu.Lock()
	defer mu.Unlock()

	var buf [8]C.ODBSPLOAD
	cn := C.short(n)
	ret := C.cnc_rdspmeter(C.ushort(h), 1, &cn, &buf[0])
	if ret != C.EW_OK {
		return nil, focasErr("cnc_rdspmeter(load)", int(ret))
	}
	n = int(cn)

	out := make([]SpindleLoad, n)
	for i := 0; i < n; i++ {
		out[i] = SpindleLoad{
			Load:   int32(buf[i].spload.data),
			PowerW: int32(buf[i].spspeed.data),
		}
	}
	return out, nil
}

func (cgoBinding) ActualFeed(h Handle) (int32, error) {
	mu.Lock()
	defer mu.Unlock()

	var val C.ODBACT
	ret := C.cnc_actf(C.ushort(h), &val)
	if ret != C.EW_OK {
		return 0, focasErr("cnc_actf", int(ret))
	}
	return int32(val.data), nil
}

func (cgoBinding) Speed(h Handle) (Speed, error) {
	mu.Lock()
	defer mu.Unlock()

	var spd C.ODBSPEED
	ret := C.cnc_rdspeed(C.ushort(h), -1, &spd)
	if ret != C.EW_OK {
		return Speed{}, focasErr("cnc_rdspeed", int(ret))
	}
	return Speed{
		ActualFeed: int32(spd.actf.data),
		JogFeed:    int32(spd.acts.data),
	}, nil
}

func (cgoBinding) FeedOverride(h Handle) (int, error) {
	return readOverride(h, "feed")
}

func (cgoBinding) JogOverride(h Handle) (int, error) {
	return readOverride(h, "jog")
}

func readOverride(h Handle, kind string) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	var val C.ODBTOFS
	// type 1 = feed override, type 2 = jog override
	t := C.short(1)
	if kind == "jog" {
		t = 2
	}
	ret := C.cnc_rdtofs(C.ushort(h), t, C.short(0), C.short(C.sizeof_ODBTOFS), &val)
	if ret != C.EW_OK {
		return 0, focasErr("cnc_rdtofs", int(ret))
	}
	return int(val.data), nil
}

func (cgoBinding) ExecProgName(h Handle) (ExecProg, error) {
	mu.Lock()
	defer mu.Unlock()

	var ep C.ODBEXEPRG
	ret := C.cnc_exeprgname(C.ushort(h), &ep)
	if ret != C.EW_OK {
		return ExecProg{}, focasErr("cnc_exeprgname", int(ret))
	}
	return ExecProg{
		Name:   C.GoString(&ep.name[0]),
		Number: int32(ep.o_num),
	}, nil
}

func (cgoBinding) ReadExecProg(h Handle, bufLen int) (ExecProgBlock, error) {
	mu.Lock()
	defer mu.Unlock()

	buf := make([]byte, bufLen)
	var blkno C.char
	length := C.short(bufLen)
	ulen := C.ushort(bufLen)
	ret := C.cnc_rdexecprog(C.ushort(h), &ulen, &length, (*C.char)(unsafe.Pointer(&buf[0])))
	if ret != C.EW_OK {
		return ExecProgBlock{}, focasErr("cnc_rdexecprog", int(ret))
	}
	_ = blkno
	return ExecProgBlock{Block: string(buf[:int(length)])}, nil
}

// AlarmStatus returns the bit-mask of active alarm categories from cnc_alarm2.
// Bit 30 indicates emergency stop on 30i/31i/32i controllers.
func (cgoBinding) AlarmStatus(h Handle) (uint32, error) {
	mu.Lock()
	defer mu.Unlock()

	var status C.long
	ret := C.cnc_alarm2(C.ushort(h), &status)
	if ret != C.EW_OK {
		return 0, focasErr("cnc_alarm2", int(ret))
	}
	return uint32(status), nil
}

func (cgoBinding) AlarmMsgs(h Handle, max int) ([]AlarmMsg, error) {
	mu.Lock()
	defer mu.Unlock()

	n := C.short(max)
	msgs := make([]C.ODBALMMSG2, max)
	ret := C.cnc_rdalmmsg2(C.ushort(h), -1, &n, &msgs[0])
	if ret != C.EW_OK {
		return nil, focasErr("cnc_rdalmmsg2", int(ret))
	}

	out := make([]AlarmMsg, int(n))
	for i := 0; i < int(n); i++ {
		m := &msgs[i]
		out[i] = AlarmMsg{
			AlmNo:  int32(m.alm_no),
			Type:   int16(m._type),
			Axis:   int16(m.axis),
			MsgLen: int16(m.msg_len),
			Msg:    C.GoStringN(&m.alm_msg[0], C.int(m.msg_len)),
		}
	}
	return out, nil
}

func (cgoBinding) Diag(h Handle, no, axis int) (DiagValue, error) {
	mu.Lock()
	defer mu.Unlock()

	var buf C.ODBDGN
	ret := C.cnc_diagnoss(C.ushort(h), C.short(no), C.short(axis), C.short(C.sizeof_ODBDGN), &buf)
	if ret != C.EW_OK {
		return DiagValue{}, focasErr("cnc_diagnoss", int(ret))
	}
	return DiagValue{
		No:   int16(buf.datano),
		Axis: int16(axis),
		Kind: int16(buf._type),
		Int:  int64(C.focas_dgn_ldata(&buf)),
	}, nil
}

func (cgoBinding) Param(h Handle, no, axis int) (ParamValue, error) {
	mu.Lock()
	defer mu.Unlock()

	var buf C.IODBPSD
	ret := C.cnc_rdparam(C.ushort(h), C.short(no), C.short(axis), C.short(C.sizeof_IODBPSD), &buf)
	if ret != C.EW_OK {
		return ParamValue{}, focasErr("cnc_rdparam", int(ret))
	}
	return ParamValue{
		No:   int16(buf.datano),
		Axis: int16(axis),
		Kind: int16(buf._type),
		Int:  int64(C.focas_psd_ldata(&buf)),
	}, nil
}

func (cgoBinding) ProductionParams(h Handle) (ProductionParams, error) {
	parts, err := (cgoBinding{}).Param(h, 6711, 0)
	if err != nil {
		return ProductionParams{}, err
	}
	powerOn, err := (cgoBinding{}).Param(h, 6750, 0)
	if err != nil {
		return ProductionParams{}, err
	}
	operating, err := (cgoBinding{}).Param(h, 6751, 0)
	if err != nil {
		return ProductionParams{}, err
	}
	cutting, err := (cgoBinding{}).Param(h, 6753, 0)
	if err != nil {
		return ProductionParams{}, err
	}
	cycle, err := (cgoBinding{}).Param(h, 6757, 0)
	if err != nil {
		return ProductionParams{}, err
	}
	return ProductionParams{
		PartsCount: parts.Int,
		PowerOn:    time.Duration(powerOn.Int) * time.Minute,
		Operating:  time.Duration(operating.Int) * time.Minute,
		Cutting:    time.Duration(cutting.Int) * time.Minute,
		Cycle:      time.Duration(cycle.Int) * time.Millisecond,
	}, nil
}

func focasErr(op string, code int) error {
	return &FocasError{Op: op, Code: code}
}
