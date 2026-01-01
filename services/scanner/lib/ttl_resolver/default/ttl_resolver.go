package _default

import (
	"time"
)

const dayDuration = time.Hour * 24

type TTLResolver struct {
	candleTTL time.Duration
	chartTTL  time.Duration
}

func NewTTLResolver(candleTTL time.Duration, chartTTL time.Duration) *TTLResolver {
	return &TTLResolver{
		candleTTL: candleTTL,
		chartTTL:  chartTTL,
	}
}

func (T TTLResolver) CandleTTL(searchTo time.Time) time.Duration {
	if searchTo.Before(time.Now()) {
		return dayDuration
	}
	return T.candleTTL
}

func (T TTLResolver) ChartTTL(searchTo time.Time) time.Duration {
	if searchTo.Before(time.Now()) {
		return dayDuration
	}
	return T.chartTTL
}
