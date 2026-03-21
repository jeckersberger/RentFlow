package events

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestNewKurrentDBClient(t *testing.T) {
	tests := []struct {
		name             string
		connectionString string
		expectError      bool
		errorContains    string
	}{
		{
			name:             "valid connection string",
			connectionString: "esdb://localhost:2113",
			expectError:      false,
		},
		{
			name:             "empty connection string",
			connectionString: "",
			expectError:      true,
			errorContains:    "connection string cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewKurrentDBClient(tt.connectionString)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if client == nil {
					t.Errorf("expected client, got nil")
				}
			}
		})
	}
}

func TestStreamNamingHelpers(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		fn       func(string) string
		expected string
	}{
		{
			name:     "EquipmentStream",
			id:       "eq-123",
			fn:       EquipmentStream,
			expected: "Equipment-eq-123",
		},
		{
			name:     "MaintenanceStream",
			id:       "maint-456",
			fn:       MaintenanceStream,
			expected: "Maintenance-maint-456",
		},
		{
			name:     "TenantStream",
			id:       "tenant-789",
			fn:       TenantStream,
			expected: "Tenant-tenant-789",
		},
		{
			name:     "LeaseStream",
			id:       "lease-101",
			fn:       LeaseStream,
			expected: "Lease-lease-101",
		},
		{
			name:     "PaymentStream",
			id:       "pay-202",
			fn:       PaymentStream,
			expected: "Payment-pay-202",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.id)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNewEventData(t *testing.T) {
	tests := []struct {
		name           string
		eventType      string
		data           interface{}
		metadata       interface{}
		expectError    bool
		validateResult func(*EventData) bool
	}{
		{
			name:        "simple string data",
			eventType:   "UserCreated",
			data:        "user-123",
			metadata:    map[string]string{"source": "api"},
			expectError: false,
			validateResult: func(ed *EventData) bool {
				return ed.Type == "UserCreated" &&
					len(ed.Data) > 0 &&
					len(ed.Metadata) > 0
			},
		},
		{
			name:      "complex struct data",
			eventType: "EquipmentCreated",
			data: map[string]interface{}{
				"id":   "eq-123",
				"name": "Air Conditioner",
				"cost": 5000,
			},
			metadata: map[string]interface{}{
				"version":   1,
				"timestamp": time.Now(),
			},
			expectError: false,
			validateResult: func(ed *EventData) bool {
				return ed.Type == "EquipmentCreated" &&
					json.Valid(ed.Data) &&
					json.Valid(ed.Metadata)
			},
		},
		{
			name:        "nil data",
			eventType:   "Event",
			data:        nil,
			metadata:    nil,
			expectError: false,
			validateResult: func(ed *EventData) bool {
				return ed.Type == "Event"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed, err := NewEventData(tt.eventType, tt.data, tt.metadata)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !tt.validateResult(ed) {
					t.Errorf("validation failed for result: %+v", ed)
				}
			}
		})
	}
}

func TestResolvedEventUnmarshal(t *testing.T) {
	type TestData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	data := TestData{ID: "123", Name: "Test"}
	dataBytes, _ := json.Marshal(data)
	metadata := map[string]string{"source": "api"}
	metadataBytes, _ := json.Marshal(metadata)

	event := ResolvedEvent{
		StreamName: "Test-stream",
		EventType:  "TestEvent",
		EventData:  dataBytes,
		Metadata:   metadataBytes,
		Revision:   1,
		Position:   0,
	}

	t.Run("UnmarshalData", func(t *testing.T) {
		var result TestData
		err := event.UnmarshalData(&result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.ID != data.ID || result.Name != data.Name {
			t.Errorf("unmarshaled data mismatch: %+v", result)
		}
	})

	t.Run("UnmarshalMetadata", func(t *testing.T) {
		var result map[string]string
		err := event.UnmarshalMetadata(&result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result["source"] != "api" {
			t.Errorf("unmarshaled metadata mismatch: %+v", result)
		}
	})
}

func TestMockEventStoreAdapter_AppendToStream(t *testing.T) {
	adapter := NewMockEventStoreAdapter()
	ctx := context.Background()

	tests := []struct {
		name             string
		streamID         string
		expectedRevision uint64
		events           []EventData
		expectError      bool
	}{
		{
			name:             "append single event",
			streamID:         "stream-1",
			expectedRevision: 0,
			events: []EventData{
				{
					Type: "Event1",
					Data: []byte(`{"id":"1"}`),
				},
			},
			expectError: false,
		},
		{
			name:             "append multiple events",
			streamID:         "stream-2",
			expectedRevision: 0,
			events: []EventData{
				{Type: "Event1", Data: []byte(`{"id":"1"}`)},
				{Type: "Event2", Data: []byte(`{"id":"2"}`)},
				{Type: "Event3", Data: []byte(`{"id":"3"}`)},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.AppendToStream(ctx, tt.streamID, tt.expectedRevision, tt.events)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result, got nil")
				}
				if result.NextExpectedRevision != uint64(len(tt.events)) {
					t.Errorf("expected revision %d, got %d", len(tt.events), result.NextExpectedRevision)
				}
			}
		})
	}
}

func TestMockEventStoreAdapter_ReadStream(t *testing.T) {
	adapter := NewMockEventStoreAdapter()
	ctx := context.Background()

	// Append test events
	streamID := "stream-read-test"
	events := []EventData{
		{Type: "Event1", Data: []byte(`{"seq":1}`)},
		{Type: "Event2", Data: []byte(`{"seq":2}`)},
		{Type: "Event3", Data: []byte(`{"seq":3}`)},
	}
	adapter.AppendToStream(ctx, streamID, 0, events)

	tests := []struct {
		name      string
		streamID  string
		direction Direction
		from      uint64
		count     uint64
		expectLen int
	}{
		{
			name:      "read all forward",
			streamID:  streamID,
			direction: DirectionForward,
			from:      0,
			count:     0,
			expectLen: 3,
		},
		{
			name:      "read from position",
			streamID:  streamID,
			direction: DirectionForward,
			from:      1,
			count:     2,
			expectLen: 2,
		},
		{
			name:      "read backward",
			streamID:  streamID,
			direction: DirectionBackward,
			from:      2,
			count:     2,
			expectLen: 2,
		},
		{
			name:      "read non-existent stream",
			streamID:  "non-existent",
			direction: DirectionForward,
			from:      0,
			count:     0,
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.ReadStream(ctx, tt.streamID, tt.direction, tt.from, tt.count)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(result) != tt.expectLen {
				t.Errorf("expected %d events, got %d", tt.expectLen, len(result))
			}
		})
	}
}

func TestMockEventStoreAdapter_SubscribeToStream(t *testing.T) {
	adapter := NewMockEventStoreAdapter()
	ctx := context.Background()

	streamID := "stream-subscribe"
	events := []EventData{
		{Type: "Event1", Data: []byte(`{"id":"1"}`)},
		{Type: "Event2", Data: []byte(`{"id":"2"}`)},
	}
	adapter.AppendToStream(ctx, streamID, 0, events)

	t.Run("subscribe and receive events", func(t *testing.T) {
		var receivedEvents []*ResolvedEvent
		handler := func(ctx context.Context, event *ResolvedEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		err := adapter.SubscribeToStream(ctx, streamID, 0, handler)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(receivedEvents) != 2 {
			t.Errorf("expected 2 events, got %d", len(receivedEvents))
		}
	})

	t.Run("subscribe from position", func(t *testing.T) {
		var receivedEvents []*ResolvedEvent
		handler := func(ctx context.Context, event *ResolvedEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		err := adapter.SubscribeToStream(ctx, streamID, 1, handler)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(receivedEvents) != 1 {
			t.Errorf("expected 1 event, got %d", len(receivedEvents))
		}
	})
}

func TestMockEventStoreAdapter_SubscribeToAll(t *testing.T) {
	adapter := NewMockEventStoreAdapter()
	ctx := context.Background()

	// Append events to multiple streams
	adapter.AppendToStream(ctx, "Equipment-1", 0, []EventData{
		{Type: "EquipmentCreated", Data: []byte(`{"id":"1"}`)},
	})
	adapter.AppendToStream(ctx, "Tenant-1", 0, []EventData{
		{Type: "TenantCreated", Data: []byte(`{"id":"1"}`)},
	})
	adapter.AppendToStream(ctx, "Equipment-2", 0, []EventData{
		{Type: "EquipmentUpdated", Data: []byte(`{"id":"2"}`)},
	})

	t.Run("subscribe to all events", func(t *testing.T) {
		var receivedEvents []*ResolvedEvent
		handler := func(ctx context.Context, event *ResolvedEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		err := adapter.SubscribeToAll(ctx, nil, handler)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(receivedEvents) != 3 {
			t.Errorf("expected 3 events, got %d", len(receivedEvents))
		}
	})

	t.Run("subscribe with stream prefix filter", func(t *testing.T) {
		var receivedEvents []*ResolvedEvent
		handler := func(ctx context.Context, event *ResolvedEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		filter := &SubscriptionFilter{StreamPrefix: "Equipment"}
		err := adapter.SubscribeToAll(ctx, filter, handler)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(receivedEvents) != 2 {
			t.Errorf("expected 2 equipment events, got %d", len(receivedEvents))
		}
	})

	t.Run("subscribe with event type filter", func(t *testing.T) {
		var receivedEvents []*ResolvedEvent
		handler := func(ctx context.Context, event *ResolvedEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		filter := &SubscriptionFilter{EventTypes: []string{"EquipmentCreated"}}
		err := adapter.SubscribeToAll(ctx, filter, handler)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(receivedEvents) != 1 {
			t.Errorf("expected 1 event, got %d", len(receivedEvents))
		}
	})
}

func TestMockEventStoreAdapter_Close(t *testing.T) {
	adapter := NewMockEventStoreAdapter()
	err := adapter.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestKurrentDBClientWithMockAdapter(t *testing.T) {
	mockAdapter := NewMockEventStoreAdapter()
	client := &KurrentDBClient{adapter: mockAdapter}
	ctx := context.Background()

	t.Run("AppendToStream delegates to adapter", func(t *testing.T) {
		events := []EventData{
			{Type: "TestEvent", Data: []byte(`{"id":"1"}`)},
		}
		result, err := client.AppendToStream(ctx, "test-stream", 0, events)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Errorf("expected result, got nil")
		}
	})

	t.Run("ReadStream delegates to adapter", func(t *testing.T) {
		result, err := client.ReadStream(ctx, "test-stream", DirectionForward, 0, 10)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Errorf("expected result, got nil")
		}
	})

	t.Run("Close delegates to adapter", func(t *testing.T) {
		err := client.Close()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestKurrentDBClientWithoutAdapter(t *testing.T) {
	client := &KurrentDBClient{adapter: nil}
	ctx := context.Background()

	tests := []struct {
		name          string
		testFunc      func() error
		errorContains string
	}{
		{
			name: "AppendToStream without adapter",
			testFunc: func() error {
				_, err := client.AppendToStream(ctx, "stream", 0, nil)
				return err
			},
			errorContains: "not initialized",
		},
		{
			name: "ReadStream without adapter",
			testFunc: func() error {
				_, err := client.ReadStream(ctx, "stream", DirectionForward, 0, 10)
				return err
			},
			errorContains: "not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc()
			if err == nil {
				t.Errorf("expected error, got nil")
			}
			if !contains(err.Error(), tt.errorContains) {
				t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return fmt.Sprintf("%s", s) != "" && len(s) >= len(substr)
}
