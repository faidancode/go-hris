package producer

import (
	"context"
	"go-hris/internal/messaging/kafka"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func ProcessOutboxEvents(
	ctx context.Context,
	repo kafka.OutboxRepository,
	writer *kafkago.Writer,
	logger *zap.Logger,
	pollInterval time.Duration,
) {
	if pollInterval <= 0 {
		pollInterval = 3 * time.Second
	}

	log := logger.Named("kafka.producer.worker")
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Info("outbox worker started", zap.Duration("poll_interval", pollInterval))

	for {
		select {
		case <-ctx.Done():
			log.Info("outbox worker stopped")
			return
		case <-ticker.C:
			if err := processPendingEvents(ctx, repo, writer, log); err != nil {
				log.Error("process outbox events failed", zap.Error(err))
			}
		}
	}
}

func processPendingEvents(
	ctx context.Context,
	repo kafka.OutboxRepository,
	writer *kafkago.Writer,
	logger *zap.Logger,
) error {
	events, err := repo.ListPending(ctx, 50)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	logger.Info("processing pending outbox events", zap.Int("count", len(events)))

	for _, event := range events {
		if err := publishEvent(ctx, writer, event); err != nil {
			logger.Error("publish outbox event failed",
				zap.String("outbox_id", event.ID),
				zap.String("event_type", event.EventType),
				zap.String("topic", event.Topic),
				zap.Error(err),
			)
			_ = repo.MarkFailed(ctx, event.ID, err.Error())
			continue
		}

		if err := repo.MarkSent(ctx, event.ID); err != nil {
			logger.Error("mark outbox sent failed",
				zap.String("outbox_id", event.ID),
				zap.Error(err),
			)
			continue
		}

		logger.Info("outbox event sent",
			zap.String("outbox_id", event.ID),
			zap.String("event_type", event.EventType),
			zap.String("topic", event.Topic),
		)
	}

	return nil
}
