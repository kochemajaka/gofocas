package focas

import (
	"errors"
	"fmt"

	"github.com/kochemajaka/gofocas/internal/fwlib32"
)

// Code maps to the EW_* return values from fwlib32.
type Code int

const (
	CodeOK         Code = 0
	CodeFunc       Code = 1
	CodeLength     Code = 2
	CodeNumber     Code = 3
	CodeAttrib     Code = 4
	CodeData       Code = 5
	CodeNoOpt      Code = 6
	CodeProtect    Code = 7
	CodeOverflow   Code = 8
	CodeParam      Code = 9
	CodeBuffer     Code = 10
	CodePath       Code = 11
	CodeMode       Code = 12
	CodeReject     Code = 13
	CodeDataSrv    Code = 14
	CodeAlarm      Code = 15
	CodeStop       Code = 16
	CodePasswd     Code = 17
	CodeReset      Code = -2
	CodeBusy       Code = -1
	CodeSystem     Code = -5
	CodeParity     Code = -4
	CodeMMCSys     Code = -3
	CodeHandle     Code = -8
	CodeVersion    Code = -7
	CodeUnexpected Code = -6
	CodeHSSB       Code = -9
	CodeSystem2    Code = -10
	CodeBus        Code = -11
	CodeNoDLL      Code = -15
	CodeSocket     Code = -16
	CodeProtocol   Code = -17
	CodeNoPMC      Code = -101
	CodeRange      Code = -103
	CodeType       Code = -104
)

func (c Code) String() string {
	switch c {
	case CodeOK:
		return "EW_OK"
	case CodeFunc:
		return "EW_FUNC"
	case CodeLength:
		return "EW_LENGTH"
	case CodeNumber:
		return "EW_NUMBER"
	case CodeAttrib:
		return "EW_ATTRIB"
	case CodeData:
		return "EW_DATA"
	case CodeNoOpt:
		return "EW_NOOPT"
	case CodeProtect:
		return "EW_PROTECT"
	case CodeOverflow:
		return "EW_OVERFLOW"
	case CodeParam:
		return "EW_PARAM"
	case CodeBuffer:
		return "EW_BUFFER"
	case CodePath:
		return "EW_PATH"
	case CodeMode:
		return "EW_MODE"
	case CodeReject:
		return "EW_REJECT"
	case CodeDataSrv:
		return "EW_DATASRV"
	case CodeAlarm:
		return "EW_ALARM"
	case CodeStop:
		return "EW_STOP"
	case CodePasswd:
		return "EW_PASSWD"
	case CodeReset:
		return "EW_RESET"
	case CodeBusy:
		return "EW_BUSY"
	case CodeSystem:
		return "EW_SYSTEM"
	case CodeParity:
		return "EW_PARITY"
	case CodeMMCSys:
		return "EW_MMCSYS"
	case CodeHandle:
		return "EW_HANDLE"
	case CodeVersion:
		return "EW_VERSION"
	case CodeUnexpected:
		return "EW_UNEXPECTED"
	case CodeHSSB:
		return "EW_HSSB"
	case CodeSystem2:
		return "EW_SYSTEM2"
	case CodeBus:
		return "EW_BUS"
	case CodeNoDLL:
		return "EW_NODLL"
	case CodeSocket:
		return "EW_SOCKET"
	case CodeProtocol:
		return "EW_PROTOCOL"
	case CodeNoPMC:
		return "EW_NOPMC"
	case CodeRange:
		return "EW_RANGE"
	case CodeType:
		return "EW_TYPE"
	default:
		return fmt.Sprintf("EW_UNKNOWN(%d)", int(c))
	}
}

// IsTransient returns true for codes that indicate a recoverable connection
// loss; the reconnect logic uses this to decide whether to re-dial.
func (c Code) IsTransient() bool {
	switch c {
	case CodeHandle, CodeSocket, CodeBusy, CodeReset:
		return true
	}
	return false
}

// Error is the structured error type returned by all Client methods.
type Error struct {
	Op  string // FOCAS function name or high-level op, e.g. "cnc_rdposition"
	Code Code
	Msg string
	Err error // wrapped underlying error
}

func (e *Error) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("focas: %s: %s (%s)", e.Op, e.Msg, e.Code)
	}
	if e.Err != nil && e.Code == CodeOK {
		return fmt.Sprintf("focas: %s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("focas: %s: %s", e.Op, e.Code)
}

func (e *Error) Unwrap() error { return e.Err }

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	if t.Code != 0 && t.Code != e.Code {
		return false
	}
	if t.Op != "" && t.Op != e.Op {
		return false
	}
	return true
}

// Sentinel errors for use with errors.Is.
var (
	ErrClosed       = errors.New("focas: client closed")
	ErrNotConnected = errors.New("focas: not connected")
	ErrUnsupported  = errors.New("focas: unsupported on this series")
	ErrTimeout      = errors.New("focas: timeout")
)

// CodeOf extracts the Code from any error in the chain.
// Returns (CodeOK, false) if no *Error is found.
func CodeOf(err error) (Code, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e.Code, true
	}
	return CodeOK, false
}

// IsTransient reports whether any *Error in the chain has a transient code.
func IsTransient(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code.IsTransient()
	}
	return false
}

func wrapErr(op string, code Code, msg string) error {
	return &Error{Op: op, Code: code, Msg: msg}
}

// wrapFocasErr wraps a *fwlib32.FocasError (or any error) into *Error,
// extracting the EW_* code so it appears correctly in the message.
func wrapFocasErr(op, msg string, err error) error {
	var fe *fwlib32.FocasError
	if errors.As(err, &fe) {
		return &Error{Op: op, Code: Code(fe.Code), Msg: msg, Err: err}
	}
	return &Error{Op: op, Msg: msg, Err: err}
}
