package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
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

// EsdbAdapter implements EventStoreAdapter using EventStoreDB
type EsdbAdapter struct {
	client *esdb.Client
}

// NewKurrentDBClient creates a new KurrentDB client with a real EventStore connection
func NewKurrentDBClient(connectionString string) (*KurrentDBClient, error) {
	if connectionString == "" {
		return nil, fmt.Errorf("connection string cannot be empty")
	}

	// Parse connection string
	opts, err := esdb.ParseConnectionString(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Create EventStore client
	client, err := esdb.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create event store client: %w", err)
	}

	adapter := &EsdbAdapter{
		client: client,
	}

	return &KurrentDBClient{
		adapter: adapter,
	}, nil
}

// NewEventStoreFromEnv creates a KurrentDB client from environment variables
// Falls back to NoopEventStore if connection fails
func NewEventStoreFromEnv() *KurrentDBClient {
	connString := os.Getenv("KURRENTDB_URL")
	if connString == "" {
		connString = "esdb://localhost:2113?tls=false"
	}

	client, err := NewKurrentDBClient(connString)
	if err != nil {
		log.Printf("Warning: failed to connect to KurrentDB at %s, using noop event store: %v\n", connString, err)
		return &KurrentDBClient{
			adapter: NewNoopEventStore(),
		}
	}

	return client
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

// --- EsdbAdapter Implementation ---

// AppendToStream appends events to a stream using EventStoreDB
func (e *EsdbAdapter) AppendToStream(ctx context.Context, streamID string, expectedRevision uint64, events []EventData) (*WriteResult, error) {
	// Convert EventData to esdb.EventData (values, not pointers)
	esdbEvents := make([]esdb.EventData, len(events))
	for i, evt := range events {
		esdbEvents[i] = esdb.EventData{
			ContentType: esdb.ContentTypeJson,
			EventType:   evt.Type,
			Data:        evt.Data,
			Metadata:    evt.Metadata,
		}
	}

	// Append to stream with expected revision
	writeResult, err := e.client.AppendToStream(ctx, streamID, esdb.AppendToStreamOptions{
		ExpectedRevision: esdb.Revision(expectedRevision),
	}, esdbEvents...)
	if err != nil {
		return nil, fmt.Errorf("failed to append to stream: %w", err)
	}

	return &WriteResult{
		NextExpectedRevision: writeResult.NextExpectedVersion,
		Position:             writeResult.CommitPosition,
	}, nil
}

// ReadStream reads events from a stream
func (e *EsdbAdapter) ReadStream(ctx context.Context, streamID string, direction Direction, from uint64, count uint64) ([]ResolvedEvent, error) {
	opts := esdb.ReadStreamOptions{
		Direction: esdb.Forwards,
		From:      esdb.Revision(from),
	}

	if direction == DirectionBackward {
		opts.Direction = esdb.Backwards
	}

	stream, err := e.client.ReadStream(ctx, streamID, opts, count)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream: %w", err)
	}
	defer stream.Close()

	var resolved []ResolvedEvent
	for {
		event, err := stream.Recv()
		if err != nil {
			// io.EOF means we've read all events
			break
		}

		if event.Event == nil {
			continue
		}

		resolvedEvent := ResolvedEvent{
			StreamName: event.Event.StreamID,
			EventType:  event.Event.EventType,
			EventData:  event.Event.Data,
			Metadata:   event.Event.UserMetadata,
			Revision:   event.Event.EventNumber,
			Timestamp:  event.Event.CreatedDate.Unix(),
		}

		if event.Event.Position.Commit > 0 {
			resolvedEvent.Position = event.Event.Position.Commit
		}

		resolved = append(resolved, resolvedEvent)
	}

	return resolved, nil
}

// SubscribeToStream subscribes to events from a specific stream
func (e *EsdbAdapter) SubscribeToStream(ctx context.Context, streamID string, from uint64, handler EventHandler) error {
	opts := esdb.SubscribeToStreamOptions{
		From: esdb.Revision(from),
	}

	subscription, err := e.client.SubscribeToStream(ctx, streamID, opts)
	if err != nil {
		return fmt.Errorf("failed to subscribe to stream: %w", err)
	}
	defer subscription.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		subEvent := subscription.Recv()
		if subEvent.EventAppeared == nil {
			if subEvent.SubscriptionDropped != nil {
				return fmt.Errorf("subscription dropped: %v", subEvent.SubscriptionDropped.Error)
			}
			continue
		}

		event := subEvent.EventAppeared
		resolvedEvent := &ResolvedEvent{
			StreamName: event.Event.StreamID,
			EventType:  event.Event.EventType,
			EventData:  event.Event.Data,
			Metadata:   event.Event.UserMetadata,
			Revision:   event.Event.EventNumber,
			Timestamp:  event.Event.CreatedDate.Unix(),
		}

		if err := handler(ctx, resolvedEvent); err != nil {
			return err
		}
	}
}

// SubscribeToAll subscribes to all events with optional filtering
func (e *EsdbAdapter) SubscribeToAll(ctx context.Context, filter *SubscriptionFilter, handler EventHandler) error {
	opts := esdb.SubscribeToAllOptions{}

	if filter != nil && len(filter.EventTypes) > 0 {
		opts.Filter = &esdb.SubscriptionFilter{
			Type:     esdb.EventFilterType,
			Prefixes: filter.EventTypes,
		}
	}

	subscription, err := e.client.SubscribeToAll(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to subscribe to all: %w", err)
	}
	defer subscription.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		subEvent := subscription.Recv()
		if subEvent.EventAppeared == nil {
			if subEvent.SubscriptionDropped != nil {
				return fmt.Errorf("subscription dropped: %v", subEvent.SubscriptionDropped.Error)
			}
			continue
		}

		event := subEvent.EventAppeared

		// Apply stream prefix filter client-side if specified
		if filter != nil && filter.StreamPrefix != "" {
			streamName := event.Event.StreamID
			if len(streamName) < len(filter.StreamPrefix) || streamName[:len(filter.StreamPrefix)] != filter.StreamPrefix {
				continue
			}
		}

		resolvedEvent := &ResolvedEvent{
			StreamName: event.Event.StreamID,
			EventType:  event.Event.EventType,
			EventData:  event.Event.Data,
			Metadata:   event.Event.UserMetadata,
			Revision:   event.Event.EventNumber,
			Timestamp:  event.Event.CreatedDate.Unix(),
		}

		if err := handler(ctx, resolvedEvent); err != nil {
			return err
		}
	}
}

// Close closes the EventStoreDB connection
func (e *EsdbAdapter) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}

// --- NoopEventStore Fallback Implementation ---

// NoopEventStore is a no-op implementation used when EventStoreDB is not available
type NoopEventStore struct{}

// NewNoopEventStore creates a new noop event store
func NewNoopEventStore() EventStoreAdapter {
	return &NoopEventStore{}
}

// AppendToStream implements EventStoreAdapter (noop)
func (n *NoopEventStore) AppendToStream(ctx context.Context, streamID string, expectedRevision uint64, events []EventData) (*WriteResult, error) {
	log.Printf("NoopEventStore: AppendToStream called for stream %s with %d events (noop)", streamID, len(events))
	return &WriteResult{
		NextExpectedRevision: expectedRevision + uint64(len(events)),
		Position:             expectedRevision + uint64(len(events)) - 1,
	}, nil
}

// ReadStream implements EventStoreAdapter (noop)
func (n *NoopEventStore) ReadStream(ctx context.Context, streamID string, direction Direction, from uint64, count uint64) ([]ResolvedEvent, error) {
	log.Printf("NoopEventStore: ReadStream called for stream %s (noop)", streamID)
	return []ResolvedEvent{}, nil
}

// SubscribeToStream implements EventStoreAdapter (noop)
func (n *NoopEventStore) SubscribeToStream(ctx context.Context, streamID string, from uint64, handler EventHandler) error {
	log.Printf("NoopEventStore: SubscribeToStream called for stream %s (noop)", streamID)
	return nil
}

// SubscribeToAll implements EventStoreAdapter (noop)
func (n *NoopEventStore) SubscribeToAll(ctx context.Context, filter *SubscriptionFilter, handler EventHandler) error {
	log.Printf("NoopEventStore: SubscribeToAll called (noop)")
	return nil
}

// Close implements EventStoreAdapter (noop)
func (n *NoopEventStore) Close() error {
	return nil
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
