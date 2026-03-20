package events

import (
	"time"
)

// Event represents a domain event in the system
type Event interface {
	EventType() string
	EventID() string
	AggregateID() string
	AggregateType() string
	Timestamp() time.Time
	Data() interface{}
	Metadata() map[string]interface{}
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	ID            string
	Type          string
	AggID         string
	AggType       string
	CreatedAt     time.Time
	EventData     interface{}
	EventMetadata map[string]interface{}
}

func (e *BaseEvent) EventType() string {
	return e.Type
}

func (e *BaseEvent) EventID() string {
	return e.ID
}

func (e *BaseEvent) AggregateID() string {
	return e.AggID
}

func (e *BaseEvent) AggregateType() string {
	return e.AggType
}

func (e *BaseEvent) Timestamp() time.Time {
	return e.CreatedAt
}

func (e *BaseEvent) Data() interface{} {
	return e.EventData
}

func (e *BaseEvent) Metadata() map[string]interface{} {
	if e.EventMetadata == nil {
		return make(map[string]interface{})
	}
	return e.EventMetadata
}

// EventStore defines operations for event persistence
type EventStore interface {
	Append(event Event) error
	GetByAggregateID(aggregateID string) ([]Event, error)
	GetByAggregateType(aggregateType string) ([]Event, error)
	GetAll() ([]Event, error)
}

// EventPublisher defines operations for publishing events
type EventPublisher interface {
	Publish(event Event) error
	PublishBatch(events []Event) error
}
