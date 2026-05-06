package focas

// Program describes the currently executing NC program.
type Program struct {
	Name      string
	Number    int64
	GCodeLine string // currently executing block
}
