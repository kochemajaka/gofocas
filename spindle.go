package focas

// Spindle holds speed, load, and diagnostic data for one spindle.
type Spindle struct {
	Index           int
	SpeedRPM        int
	Load            float64 // %
	OverridePercent int
	PowerW          int
	Diag411         int // raw diag #411
}
