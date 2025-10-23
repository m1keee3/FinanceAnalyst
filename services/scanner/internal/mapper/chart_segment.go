package mapper

import (
	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/scanner/proto-gen/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func ToProtoChartSegment(segment models.ChartSegment) *scannerv1.ChartSegment {

	candles := make([]*scannerv1.Candle, len(segment.Candles))

	for i, c := range segment.Candles {
		candles[i] = &scannerv1.Candle{
			Date:  timestamppb.New(c.Date),
			Open:  c.Open,
			High:  c.High,
			Low:   c.Low,
			Close: c.Close,
		}
	}

	return &scannerv1.ChartSegment{
		Ticker:  segment.Ticker,
		From:    timestamppb.New(segment.From),
		To:      timestamppb.New(segment.To),
		Candles: candles,
	}
}
