package models

// ScanOptions определяет параметры сравнения свечей
type ScanOptions struct {
	TailLen         int
	BodyTolerance   float64
	ShadowTolerance float64
}

func (o *ScanOptions) WithDefaults() ScanOptions {
	out := ScanOptions{TailLen: 0, BodyTolerance: 0.1, ShadowTolerance: 0.1}
	if o == nil {
		return out
	}
	if o.TailLen > 0 {
		out.TailLen = o.TailLen
	}
	if o.BodyTolerance > 0 {
		out.BodyTolerance = o.BodyTolerance
	}
	if o.ShadowTolerance > 0 {
		out.ShadowTolerance = o.ShadowTolerance
	}
	return out
}
