package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/domain/models"
	"time"
)

type Producer struct {
	producer *kafka.Producer
	topic    string
	timeout  time.Duration
}

func New(
	brokers []string,
	topic string,
	timeout time.Duration,
) (*Producer, error) {

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":      brokers,
		"acks":                   "all",
		"retries":                5,
		"linger.ms":              10,
		"enable.idempotence":     true,
		"max.in.flight.requests": 5,
	})

	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: p,
		topic:    topic,
		timeout:  timeout,
	}, nil
}

func (p *Producer) PublishStats(segStats *models.SegmentStats) error {
	const op = "producer.Kafka.PublishStats"

	data, err := json.Marshal(segStats)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)

	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:       []byte(segStats.Segment.Ticker),
		Value:     data,
		Timestamp: time.Now(),
	}, deliveryChan)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	select {
	case ev := <-deliveryChan:
		m := ev.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return fmt.Errorf(
				"%s: delivery failed: %w",
				op,
				m.TopicPartition.Error,
			)
		}
	case <-time.After(p.timeout):
		return fmt.Errorf("%s: delivery timeout", op)
	}

	return nil
}

func (p *Producer) Close() {
	p.producer.Flush(int(p.timeout.Milliseconds()))
	p.producer.Close()
}
