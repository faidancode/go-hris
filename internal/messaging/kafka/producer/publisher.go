package producer

import (
	"context"
	"go-hris/internal/messaging/kafka"

	kafkago "github.com/segmentio/kafka-go"
)

func publishEvent(ctx context.Context, writer *kafkago.Writer, event kafka.OutboxEvent) error {
	msg := kafkago.Message{
		Topic: event.Topic,
		Key:   []byte(event.AggregateID),
		Value: event.Payload,
		Headers: []kafkago.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "aggregate_type", Value: []byte(event.AggregateType)},
		},
	}

	return writer.WriteMessages(ctx, msg)
}
