package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/domain/models"
	"github.com/segmentio/kafka-go"
	"time"
)

type Producer struct {
	writer  *kafka.Writer
	topic   string
	timeout time.Duration
}

func New(
	brokers []string,
	topic string,
	timeout time.Duration,
) *Producer {

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &Producer{
		writer:  writer,
		timeout: timeout,
	}
}

func (p *Producer) PublishStats(segStats *models.SegmentStats) error {
	const op = "producer.Kafka.PublishStats"

	data, err := json.Marshal(segStats)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	err = p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte(segStats.Segment.Ticker),
			Value: data,
			Time:  time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Producer) Close() error {
	const op = "producer.Kafka.Close"

	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
