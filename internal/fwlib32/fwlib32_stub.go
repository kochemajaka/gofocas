//go:build !focas_cgo

package fwlib32

import "errors"

var errUnsupported = errors.New("focas: fwlib32 not available (build without focas_cgo tag)")

// stub satisfies Binder when CGo / fwlib32 is not available.
type stub struct{}

// New returns the stub Binder when CGo is disabled.
func New() Binder { return stub{} }

func (stub) Startup(_ string) error                          { return nil }
func (stub) Alloc(_ string, _ uint16, _ uint32) (Handle, error) { return 0, errUnsupported }
func (stub) Free(_ Handle) error                              { return nil }
func (stub) SysInfo(_ Handle) (SysInfo, error)               { return SysInfo{}, errUnsupported }
func (stub) StatInfo(_ Handle) (StatInfo, error)             { return StatInfo{}, errUnsupported }
func (stub) Positions(_ Handle, _ int) ([]AxisPosition, error) { return nil, errUnsupported }
func (stub) SpindleMeters(_ Handle, _ int) ([]SpindleMeter, error) { return nil, errUnsupported }
func (stub) SpindleLoads(_ Handle, _ int) ([]SpindleLoad, error)   { return nil, errUnsupported }
func (stub) ActualFeed(_ Handle) (int32, error)              { return 0, errUnsupported }
func (stub) Speed(_ Handle) (Speed, error)                   { return Speed{}, errUnsupported }
func (stub) FeedOverride(_ Handle) (int, error)              { return 0, errUnsupported }
func (stub) JogOverride(_ Handle) (int, error)               { return 0, errUnsupported }
func (stub) ExecProgName(_ Handle) (ExecProg, error)         { return ExecProg{}, errUnsupported }
func (stub) ReadExecProg(_ Handle, _ int) (ExecProgBlock, error) { return ExecProgBlock{}, errUnsupported }
func (stub) AlarmMsgs(_ Handle, _ int) ([]AlarmMsg, error)   { return nil, errUnsupported }
func (stub) AlarmStatus(_ Handle) (uint32, error)            { return 0, errUnsupported }
func (stub) Diag(_ Handle, _, _ int) (DiagValue, error)      { return DiagValue{}, errUnsupported }
func (stub) Param(_ Handle, _, _ int) (ParamValue, error)    { return ParamValue{}, errUnsupported }
func (stub) ProductionParams(_ Handle) (ProductionParams, error) { return ProductionParams{}, errUnsupported }
