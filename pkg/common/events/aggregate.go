package events

import (
	"fmt"
)

// AggregateRoot is the base class for all aggregate roots in event sourcing
// It manages the aggregate's identity, version, and uncommitted changes
type AggregateRoot struct {
	ID      string
	Type    string
	Version int64
	Changes []EventData // uncommitted events
}

// NewAggregateRoot creates a new aggregate root with the given ID and type
func NewAggregateRoot(id, aggregateType string) *AggregateRoot {
	return &AggregateRoot{
		ID:      id,
		Type:    aggregateType,
		Version: 0,
		Changes: []EventData{},
	}
}

// Apply applies an event to the aggregate and records it as an uncommitted change
// This is used when the aggregate is producing new events
func (a *AggregateRoot) Apply(event EventData) {
	a.Changes = append(a.Changes, event)
}

// ApplyEvent applies an event to the aggregate without recording it as an uncommitted change
// This is used when loading historical events from the event store
func (a *AggregateRoot) ApplyEvent(event EventData) {
	// This should be overridden by concrete aggregate implementations
	// to apply the event to their state
}

// GetUncommittedChanges returns the list of uncommitted events
func (a *AggregateRoot) GetUncommittedChanges() []EventData {
	return a.Changes
}

// HasUncommittedChanges returns true if there are uncommitted changes
func (a *AggregateRoot) HasUncommittedChanges() bool {
	return len(a.Changes) > 0
}

// MarkChangesAsCommitted clears the uncommitted changes list
// This should be called after the events have been successfully persisted
func (a *AggregateRoot) MarkChangesAsCommitted() {
	a.Changes = []EventData{}
}

// IncrementVersion increments the version of the aggregate
// This should be called each time an event is applied
func (a *AggregateRoot) IncrementVersion() {
	a.Version++
}

// LoadFromHistory reconstructs the aggregate state from a history of events
// The applyFunc is a function that applies each event to the aggregate's state
func (a *AggregateRoot) LoadFromHistory(events []ResolvedEvent, applyFunc func(*AggregateRoot, *ResolvedEvent)) {
	for _, event := range events {
		applyFunc(a, &event)
		a.IncrementVersion()
	}
}

// ValidateStateAfterEvent validates that the aggregate is in a valid state
// This can be overridden by concrete aggregates for more specific validation
func (a *AggregateRoot) ValidateStateAfterEvent(event EventData) error {
	if a.ID == "" {
		return fmt.Errorf("aggregate ID cannot be empty")
	}
	if a.Type == "" {
		return fmt.Errorf("aggregate type cannot be empty")
	}
	return nil
}

// GetStreamName returns the stream name for this aggregate in the event store
// Stream names follow the convention: AggregateType-ID
func (a *AggregateRoot) GetStreamName() string {
	return fmt.Sprintf("%s-%s", a.Type, a.ID)
}

// AggregateEvent is a more strongly-typed event that includes aggregate information
// This can be used as a base for domain-specific events
type AggregateEvent struct {
	AggregateID   string
	AggregateType string
	EventType     string
	EventVersion  int
	Timestamp     int64
	Data          interface{}
	Metadata      map[string]interface{}
}

// ToEventData converts an AggregateEvent to EventData for storage
func (ae *AggregateEvent) ToEventData() (*EventData, error) {
	return NewEventData(ae.EventType, ae.Data, ae.Metadata)
}
