package focas

// Feed holds the actual feed rate and override from cnc_rdspeed / param #20.
type Feed struct {
	ActualMMPerMin  int
	OverridePercent int
}
