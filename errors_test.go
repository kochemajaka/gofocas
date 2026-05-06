package focas

import (
	"errors"
	"testing"
)

func TestErrorString(t *testing.T) {
	e := &Error{Op: "cnc_rdposition", Code: CodeSocket, Msg: "connection reset"}
	got := e.Error()
	if got == "" {
		t.Fatal("Error() returned empty string")
	}
	if !contains(got, "cnc_rdposition") {
		t.Errorf("Error() missing op: %q", got)
	}
}

func TestErrorIs(t *testing.T) {
	e := &Error{Op: "cnc_sysinfo", Code: CodeHandle}
	if !errors.Is(e, &Error{Code: CodeHandle}) {
		t.Error("errors.Is should match on Code")
	}
	if errors.Is(e, &Error{Code: CodeSocket}) {
		t.Error("errors.Is should not match different Code")
	}
}

func TestCodeOf(t *testing.T) {
	e := &Error{Op: "op", Code: CodeBusy}
	code, ok := CodeOf(e)
	if !ok {
		t.Fatal("CodeOf should find *Error")
	}
	if code != CodeBusy {
		t.Errorf("want CodeBusy, got %v", code)
	}

	// CodeOf on a plain error returns false.
	_, ok = CodeOf(ErrClosed)
	if ok {
		t.Error("CodeOf(ErrClosed) should return false")
	}
}

func TestIsTransient(t *testing.T) {
	for _, code := range []Code{CodeHandle, CodeSocket, CodeBusy, CodeReset} {
		if !code.IsTransient() {
			t.Errorf("Code %v should be transient", code)
		}
	}
	for _, code := range []Code{CodeOK, CodeFunc, CodeData} {
		if code.IsTransient() {
			t.Errorf("Code %v should not be transient", code)
		}
	}
}

func TestCodeString(t *testing.T) {
	cases := []struct {
		code Code
		want string
	}{
		{CodeOK, "EW_OK"},
		{CodeSocket, "EW_SOCKET"},
		{CodeHandle, "EW_HANDLE"},
		{Code(999), "EW_UNKNOWN(999)"},
	}
	for _, tc := range cases {
		if got := tc.code.String(); got != tc.want {
			t.Errorf("Code(%d).String() = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
