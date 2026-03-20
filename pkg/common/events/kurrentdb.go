package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// Direction specifies the direction for reading events
type Direction int

const (
	// DirectionForward reads events forward
	DirectionForward Direction = iota
	// DirectionBackward reads events backward
	DirectionBackward
)

// EventData represents an event to be stored
type EventData struct {
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
	Metadata json.RawMessage `json:"metadata"`
}

// ResolvedEvent represents an event read from the event store
type ResolvedEvent struct {
	StreamName string
	EventType  string
	EventData  json.RawMessage
	Metadata   json.RawMessage
	Revision   uint64
	Position   uint64
	Timestamp  int64
}

// WriteResult represents the result of appending events
type WriteResult struct {
	NextExpectedRevision uint64
	Position             uint64
}

// SubscriptionFilter defines criteria for subscribing to events
type SubscriptionFilter struct {
	StreamPrefix string
	EventTypes   []string
}

// EventHandler handles an event during subscription
type EventHandler func(context.Context, *ResolvedEvent) error

// EventStoreAdapter defines the interface for event store operations
// This adapter pattern allows swapping EventStore implementations
type EventStoreAdapter interface {
	// AppendToStream appends events to a stream
	AppendToStream(ctx context.Context, streamID string, expectedRevision uint64, events []EventData) (*WriteResult, error)

	// ReadStream reads events from a stream
	ReadStream(ctx context.Context, streamID string, direction Direction, from uint64, count uint64) ([]ResolvedEvent, error)

	// SubscribeToStream subscribes to events from a specific stream
	SubscribeToStream(ctx context.Context, streamID string, from uint64, handler EventHandler) error

	// SubscribeToAll subscribes to events from all streams with optional filtering
	SubscribeToAll(ctx context.Context, filter *SubscriptionFilter, handler EventHandler) error

	// Close closes the connection to the event store
	Close() error
}

// KurrentDBClient wraps the EventStore client
// It provides a high-level interface for event sourcing operations
type KurrentDBClient struct {
	adapter EventStoreAdapter
}

// NewKurrentDBClient creates a new KurrentDB client
// Note: This is a placeholder for the actual EventStore Go client initialization
// In production, this would use: github.com/EventStore/EventStore-Client-Go/v4/esdb
func NewKurrentDBClient(connectionString string) (*KurrentDBClient, error) {
	if connectionString == "" {
		return nil, fmt.Errorf("connection string cannot be empty")
	}

	// TODO: Initialize actual EventStore client
	// opts, err := esdb.ParseConnectionString(connectionString)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to parse connection string: %w", err)
	// }
	//
	// client, err := esdb.NewClient(opts)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create event store client: %w", err)
	// }

	return &KurrentDBClient{
		adapter: nil, // Will be set to actual EventStore adapter
	}, nil
}

// AppendToStream appends events to a stream
func (c *KurrentDBClient) AppendToStream(ctx context.Context, streamName string, expectedRevision uint64, events []EventData) (*WriteResult, error) {
	if c.adapter == nil {
		return nil, fmt.Errorf("event store adapter not initialized")
	}
	return c.adapter.AppendToStream(ctx, streamName, expectedRevision, events)
}

// ReadStream reads events from a stream
func (c *KurrentDBClient) ReadStream(ctx context.Context, streamName string, direction Direction, from uint64, count uint64) ([]ResolvedEvent, error) {
	if c.adapter == nil {
		return nil, fmt.Errorf("event store adapter not initialized")
	}
	return c.adapter.ReadStream(ctx, streamName, direction, from, count)
}

// SubscribeToStream subscribes to events from a specific stream
func (c *KurrentDBClient) SubscribeToStream(ctx context.Context, streamName string, from uint64, handler EventHandler) error {
	if c.adapter == nil {
		return fmt.Errorf("event store adapter not initialized")
	}
	return c.adapter.SubscribeToStream(ctx, streamName, from, handler)
}

// SubscribeToAll subscribes to all events with optional filtering
func (c *KurrentDBClient) SubscribeToAll(ctx context.Context, filter *SubscriptionFilter, handler EventHandler) error {
	if c.adapter == nil {
		return fmt.Errorf("event store adapter not initialized")
	}
	return c.adapter.SubscribeToAll(ctx, filter, handler)
}

// Close closes the event store connection
func (c *KurrentDBClient) Close() error {
	if c.adapter == nil {
		return nil
	}
	return c.adapter.Close()
}

// Stream naming convention helpers

// EquipmentStream returns the stream name for an equipment aggregate
func EquipmentStream(id string) string {
	return fmt.Sprintf("Equipment-%s", id)
}

// MaintenanceStream returns the stream name for a maintenance aggregate
func MaintenanceStream(id string) string {
	return fmt.Sprintf("Maintenance-%s", id)
}

// TenantStream returns the stream name for a tenant aggregate
func TenantStream(id string) string {
	return fmt.Sprintf("Tenant-%s", id)
}

// LeaseStream returns the stream name for a lease aggregate
func LeaseStream(id string) string {
	return fmt.Sprintf("Lease-%s", id)
}

// PaymentStream returns the stream name for a payment aggregate
func PaymentStream(id string) string {
	return fmt.Sprintf("Payment-%s", id)
}

// NewEventData creates a new EventData from arbitrary data
func NewEventData(eventType string, data interface{}, metadata interface{}) (*EventData, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event metadata: %w", err)
	}

	return &EventData{
		Type:     eventType,
		Data:     dataBytes,
		Metadata: metadataBytes,
	}, nil
}

// UnmarshalData unmarshals the event data into the provided value
func (e *ResolvedEvent) UnmarshalData(v interface{}) error {
	return json.Unmarshal(e.EventData, v)
}

// UnmarshalMetadata unmarshals the event metadata into the provided value
func (e *ResolvedEvent) UnmarshalMetadata(v interface{}) error {
	return json.Unmarshal(e.Metadata, v)
}

// MockEventStoreAdapter is a simple in-memory implementation for testing
// This allows unit tests to work without a real EventStore
type MockEventStoreAdapter struct {
	streams map[string][]ResolvedEvent
}

// NewMockEventStoreAdapter creates a new mock adapter for testing
func NewMockEventStoreAdapter() *MockEventStoreAdapter {
	return &MockEventStoreAdapter{
		streams: make(map[string][]ResolvedEvent),
	}
}

// AppendToStream implements EventStoreAdapter
func (m *MockEventStoreAdapter) AppendToStream(ctx context.Context, streamID string, expectedRevision uint64, events []EventData) (*WriteResult, error) {
	if _, exists := m.streams[streamID]; !exists {
		m.streams[streamID] = []ResolvedEvent{}
	}

	stream := m.streams[streamID]
	for i, evt := range events {
		resolved := ResolvedEvent{
			StreamName: streamID,
			EventType:  evt.Type,
			EventData:  evt.Data,
			Metadata:   evt.Metadata,
			Revision:   uint64(len(stream)) + uint64(i),
			Position:   uint64(len(stream)) + uint64(i),
		}
		stream = append(stream, resolved)
	}
	m.streams[streamID] = stream

	return &WriteResult{
		NextExpectedRevision: uint64(len(stream)),
		Position:             uint64(len(stream) - 1),
	}, nil
}

// ReadStream implements EventStoreAdapter
func (m *MockEventStoreAdapter) ReadStream(ctx context.Context, streamID string, direction Direction, from uint64, count uint64) ([]ResolvedEvent, error) {
	stream, exists := m.streams[streamID]
	if !exists {
		return []ResolvedEvent{}, nil
	}

	if count == 0 {
		count = uint64(len(stream))
	}

	var result []ResolvedEvent
	if direction == DirectionForward {
		for i := from; i < from+count && i < uint64(len(stream)); i++ {
			result = append(result, stream[i])
		}
	} else {
		for i := int(from); i >= 0 && uint64(len(result)) < count; i-- {
			if uint64(i) < uint64(len(stream)) {
				result = append(result, stream[i])
			}
		}
	}

	return result, nil
}

// SubscribeToStream implements EventStoreAdapter
func (m *MockEventStoreAdapter) SubscribeToStream(ctx context.Context, streamID string, from uint64, handler EventHandler) error {
	stream, exists := m.streams[streamID]
	if !exists {
		return nil
	}

	for i := from; i < uint64(len(stream)); i++ {
		if err := handler(ctx, &stream[i]); err != nil {
			return err
		}
	}

	return nil
}

// SubscribeToAll implements EventStoreAdapter
func (m *MockEventStoreAdapter) SubscribeToAll(ctx context.Context, filter *SubscriptionFilter, handler EventHandler) error {
	for _, stream := range m.streams {
		for _, event := range stream {
			if filter != nil && filter.StreamPrefix != "" {
				if len(event.StreamName) < len(filter.StreamPrefix) {
					continue
				}
				if event.StreamName[:len(filter.StreamPrefix)] != filter.StreamPrefix {
					continue
				}
			}

			if filter != nil && len(filter.EventTypes) > 0 {
				found := false
				for _, et := range filter.EventTypes {
					if et == event.EventType {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			if err := handler(ctx, &event); err != nil {
				return err
			}
		}
	}

	return nil
}

// Close implements EventStoreAdapter
func (m *MockEventStoreAdapter) Close() error {
	return nil
}

// CopyReaderToRawMessage reads from an io.Reader and returns json.RawMessage
func CopyReaderToRawMessage(r io.Reader) (json.RawMessage, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
