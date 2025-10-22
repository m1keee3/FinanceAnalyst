package mapper

import (
	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/scanner/proto-gen/v1"
)

func FromProtoChartSegment(proto *scannerv1.ChartSegment) models.ChartSegment {

	candles := make([]models.Candle, len(proto.GetCandles()))

	for i, c := range proto.GetCandles() {
		candles[i] = models.Candle{
			Date:  c.GetDate().AsTime(),
			Open:  c.GetOpen(),
			High:  c.GetHigh(),
			Low:   c.GetLow(),
			Close: c.GetClose(),
		}
	}

	return models.ChartSegment{
		Ticker:  proto.Ticker,
		From:    proto.GetFrom().AsTime(),
		To:      proto.GetTo().AsTime(),
		Candles: candles,
	}
}
