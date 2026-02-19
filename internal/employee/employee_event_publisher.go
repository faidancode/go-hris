package employee

import (
	"context"
	"encoding/json"
	"go-hris/internal/events"

	"github.com/segmentio/kafka-go"
)

type EventPublisher interface {
	PublishEmployeeCreated(ctx context.Context, event events.EmployeeCreatedEvent) error
}

type noopEventPublisher struct{}

func (noopEventPublisher) PublishEmployeeCreated(context.Context, events.EmployeeCreatedEvent) error {
	return nil
}

type kafkaEventPublisher struct {
	writer *kafka.Writer
}

func NewKafkaEventPublisher(writer *kafka.Writer) EventPublisher {
	return &kafkaEventPublisher{writer: writer}
}

func (p *kafkaEventPublisher) PublishEmployeeCreated(
	ctx context.Context,
	event events.EmployeeCreatedEvent,
) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: events.EmployeeCreatedTopic,
		Key:   []byte(event.EmployeeID),
		Value: payload,
	})
}
