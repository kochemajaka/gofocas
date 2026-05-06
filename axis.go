package focas

// LengthUnit is the linear measurement unit reported by the controller.
type LengthUnit uint8

const (
	UnitMillimeter LengthUnit = iota
	UnitInch
)

func (u LengthUnit) String() string {
	if u == UnitInch {
		return "inch"
	}
	return "mm"
}

// Position holds the four coordinates reported by cnc_rdposition for one axis.
type Position struct {
	Absolute float64
	Machine  float64
	Relative float64
	Distance float64 // distance to go
	Unit     LengthUnit
}

// Axis aggregates position and servo diagnostic data for one controlled axis.
type Axis struct {
	Index      int
	Name       string
	Position   Position
	Load       float64 // servo load in %
	ServoTempC int
	CoderTempC int
	PowerW     int
	Diag301    float64 // raw diag #301 (error pulse); kept because users routinely need it
}
