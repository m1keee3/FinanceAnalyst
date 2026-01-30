package scanner

type ScanStats struct {
	TotalMatches int     `json:"total_matches"`
	PriceChange  float64 `json:"price_change"`
	Probability  float64 `json:"probability"`
}
