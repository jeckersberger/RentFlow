// Package outbox implements the Transactional Outbox Pattern.
// Side-effects (email, PDF, exports) are written as events in the same DB transaction
// as the business operation, then dispatched asynchronously.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Event represents a pending side-effect in the outbox.
type Event struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	EventType     string          `json:"event_type"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   uuid.UUID       `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
	Status        string          `json:"status"`
	RetryCount    int             `json:"retry_count"`
	MaxRetries    int             `json:"max_retries"`
	ErrorMessage  string          `json:"error_message,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	ProcessedAt   *time.Time      `json:"processed_at,omitempty"`
}

// Publisher writes events to the outbox table.
type Publisher struct {
	pool *pgxpool.Pool
}

// NewPublisher creates a new outbox publisher.
func NewPublisher(pool *pgxpool.Pool) *Publisher {
	return &Publisher{pool: pool}
}

// Publish inserts an event into the outbox table.
// Should be called within the same transaction as the business operation.
func (p *Publisher) Publish(ctx context.Context, tenantID uuid.UUID, eventType, aggregateType string, aggregateID uuid.UUID, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("outbox: marshal payload: %w", err)
	}

	query := `INSERT INTO outbox_events (id, tenant_id, event_type, aggregate_type, aggregate_id, payload)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = p.pool.Exec(ctx, query, uuid.New(), tenantID, eventType, aggregateType, aggregateID, data)
	if err != nil {
		return fmt.Errorf("outbox: publish: %w", err)
	}
	return nil
}

// FetchPending retrieves pending events ready for processing.
func (p *Publisher) FetchPending(ctx context.Context, limit int) ([]*Event, error) {
	query := `
		UPDATE outbox_events
		SET status = 'processing'
		WHERE id IN (
			SELECT id FROM outbox_events
			WHERE status = 'pending'
			   OR (status = 'failed' AND retry_count < max_retries AND (next_retry_at IS NULL OR next_retry_at <= NOW()))
			ORDER BY created_at
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, tenant_id, event_type, aggregate_type, aggregate_id, payload, status, retry_count, max_retries, created_at`

	rows, err := p.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("outbox: fetch pending: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		e := &Event{}
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.EventType, &e.AggregateType, &e.AggregateID,
			&e.Payload, &e.Status, &e.RetryCount, &e.MaxRetries, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("outbox: scan: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}

// MarkCompleted marks an event as successfully processed.
func (p *Publisher) MarkCompleted(ctx context.Context, eventID uuid.UUID) error {
	_, err := p.pool.Exec(ctx,
		`UPDATE outbox_events SET status = 'completed', processed_at = NOW() WHERE id = $1`,
		eventID,
	)
	return err
}

// MarkFailed marks an event as failed with exponential backoff retry.
func (p *Publisher) MarkFailed(ctx context.Context, eventID uuid.UUID, errMsg string) error {
	_, err := p.pool.Exec(ctx,
		`UPDATE outbox_events SET status = 'failed', retry_count = retry_count + 1, error_message = $2,
		 next_retry_at = NOW() + (INTERVAL '1 minute' * POWER(2, retry_count))
		 WHERE id = $1`,
		eventID, errMsg,
	)
	return err
}
