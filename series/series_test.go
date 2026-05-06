package series_test

import (
	"testing"

	"github.com/kochemajaka/gofocas/series"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want series.Series
	}{
		{"0i", series.S0i},
		{"0", series.S0i},
		{"30i", series.S30i},
		{"30I", series.S30i},
		{"31i", series.S31i},
		{"32i", series.S32i},
		{"15", series.S15},
		{"15i", series.S15i},
		{"16i", series.S16i},
		{"18i", series.S18i},
		{"21", series.S21},
		{"unknown_xyz", series.Unknown},
		{"", series.Unknown},
	}
	for _, tc := range cases {
		got := series.Parse(tc.in)
		if got != tc.want {
			t.Errorf("Parse(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestString(t *testing.T) {
	cases := []struct {
		s    series.Series
		want string
	}{
		{series.S0i, "0i"},
		{series.S30i, "30i"},
		{series.Unknown, "unknown"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("%v.String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}
