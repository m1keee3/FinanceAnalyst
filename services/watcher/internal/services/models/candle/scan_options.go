package candle

type ScanOptions struct {
	TailLen         int
	ShadowTolerance float64
	BodyTolerance   float64
	DaysToWatch     int
}
