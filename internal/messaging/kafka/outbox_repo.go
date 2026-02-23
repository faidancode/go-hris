package kafka

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	OutboxStatusPending = "pending"
	OutboxStatusSent    = "sent"
	OutboxStatusFailed  = "failed"
)

type OutboxEvent struct {
	ID            string
	RequestID     string
	AggregateType string
	AggregateID   string
	EventType     string
	Topic         string
	Payload       []byte
	Status        string
	RetryCount    int
	NextRetryAt   time.Time
}

//go:generate mockgen -source=outbox_repo.go -destination=mock/outbox_repo_mock.go -package=mock

type OutboxRepository interface {
	WithTx(tx *sql.Tx) OutboxRepository
	Create(ctx context.Context, event OutboxEvent) error
	ListPending(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkSent(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, reason string) error
}

type outboxRepository struct {
	db *sql.DB
	tx *sql.Tx
}

func NewOutboxRepository(db *sql.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) WithTx(tx *sql.Tx) OutboxRepository {
	return &outboxRepository{db: r.db, tx: tx}
}

func (r *outboxRepository) Create(ctx context.Context, event OutboxEvent) error {
	query := `
        INSERT INTO outbox_events (
            id, request_id, aggregate_type, aggregate_id, event_type, topic, payload, status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	exec := r.execer()
	_, err := exec.ExecContext(
		ctx, query,
		event.ID, event.RequestID, event.AggregateType,
		event.AggregateID, event.EventType, event.Topic, event.Payload, event.Status,
	)
	return err
}

func (r *outboxRepository) ListPending(ctx context.Context, limit int) ([]OutboxEvent, error) {
	query := `
SELECT
	id::text,
	aggregate_type,
	aggregate_id::text,
	event_type,
	topic,
	payload,
	status,
	retry_count,
	COALESCE(next_retry_at, created_at)
FROM outbox_events
WHERE status IN ($1, $2)
	AND (next_retry_at IS NULL OR next_retry_at <= NOW())
ORDER BY created_at ASC
LIMIT $3
`

	rows, err := r.db.QueryContext(ctx, query, OutboxStatusPending, OutboxStatusFailed, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]OutboxEvent, 0, limit)
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(
			&e.ID,
			&e.AggregateType,
			&e.AggregateID,
			&e.EventType,
			&e.Topic,
			&e.Payload,
			&e.Status,
			&e.RetryCount,
			&e.NextRetryAt,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *outboxRepository) MarkSent(ctx context.Context, id string) error {
	query := `
UPDATE outbox_events
SET
	status = $2,
	processed_at = NOW(),
	error_message = NULL,
	updated_at = NOW()
WHERE id = $1
`
	_, err := r.db.ExecContext(ctx, query, id, OutboxStatusSent)
	return err
}

func (r *outboxRepository) MarkFailed(ctx context.Context, id string, reason string) error {
	query := `
UPDATE outbox_events
SET
	status = $2,
	retry_count = retry_count + 1,
	error_message = LEFT($3, 500),
	next_retry_at = NOW() + (LEAST(retry_count + 1, 10) * INTERVAL '15 seconds'),
	updated_at = NOW()
WHERE id = $1
`
	_, err := r.db.ExecContext(ctx, query, id, OutboxStatusFailed, reason)
	return err
}

func (r *outboxRepository) execer() interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func ValidateOutboxEvent(event OutboxEvent) error {
	if event.ID == "" {
		return errors.New("outbox id is required")
	}
	if event.Topic == "" {
		return errors.New("outbox topic is required")
	}
	if len(event.Payload) == 0 {
		return errors.New("outbox payload is required")
	}
	switch event.Status {
	case OutboxStatusPending, OutboxStatusSent, OutboxStatusFailed:
		return nil
	default:
		return fmt.Errorf("invalid outbox status: %s", event.Status)
	}
}
