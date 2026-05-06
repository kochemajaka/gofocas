package focas

import "github.com/kochemajaka/gofocas/series"

// System holds static information about the CNC controller.
type System struct {
	Manufacturer string
	Model        string
	Series       series.Series
	Version      string
	MaxAxes      int
	Axes         int // currently controlled axes
}
