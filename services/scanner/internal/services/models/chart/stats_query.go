package chart

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type StatsQuery struct {
	ScanQuery   *ScanQuery
	DaysToWatch int
}

func (q *StatsQuery) Hash() string {
	h := sha256.New()
	enc := json.NewEncoder(h)
	if q.ScanQuery != nil {
		_ = enc.Encode(q.ScanQuery)
	}
	_ = enc.Encode(q.DaysToWatch)
	return hex.EncodeToString(h.Sum(nil))
}
