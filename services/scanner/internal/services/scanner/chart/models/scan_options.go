package models

// ScanOptions определяет параметры сравнения графиков
type ScanOptions struct {
	// MinScale минимальная длина сегмента относительно входного
	MinScale float64
	// MaxScale максимальная длина сегмента относительно входного
	MaxScale  float64
	Tolerance float64
}

func (o *ScanOptions) WithDefaults() ScanOptions {
	out := ScanOptions{
		MinScale:  0.75,
		MaxScale:  1.5,
		Tolerance: 0.1,
	}
	if o == nil {
		return out
	}
	if o.MinScale > 0 {
		out.MinScale = o.MinScale
	}
	if o.MaxScale > 0 {
		out.MaxScale = o.MaxScale
	}
	if o.Tolerance > 0 && o.Tolerance <= 1.0 {
		out.Tolerance = o.Tolerance
	}
	return out
}
