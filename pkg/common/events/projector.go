package events

import (
	"context"
	"fmt"
	"sync"
)

// Projector handles the projection of events onto read models
// This is part of the CQRS pattern
type Projector interface {
	// ProjectEvent projects a single event onto the read model
	ProjectEvent(ctx context.Context, event *ResolvedEvent) error

	// GetProjectionName returns the name of this projection
	GetProjectionName() string

	// Reset resets the projection state (for replay scenarios)
	Reset(ctx context.Context) error
}

// ProjectionManager coordinates multiple projectors
// It handles event subscriptions and distributes events to appropriate projectors
type ProjectionManager struct {
	projectors map[string][]Projector // eventType -> []Projector
	store      EventStoreAdapter
	mu         sync.RWMutex
	running    bool
	cancel     context.CancelFunc
}

// NewProjectionManager creates a new projection manager
func NewProjectionManager(store EventStoreAdapter) *ProjectionManager {
	return &ProjectionManager{
		projectors: make(map[string][]Projector),
		store:      store,
		running:    false,
	}
}

// Register registers a projector for a specific event type
// A projector can be registered for multiple event types
func (pm *ProjectionManager) Register(eventType string, projector Projector) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.projectors[eventType] = append(pm.projectors[eventType], projector)
}

// RegisterForAllEvents registers a projector for all events
func (pm *ProjectionManager) RegisterForAllEvents(projector Projector) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Use empty string as the key for "all events"
	pm.projectors[""] = append(pm.projectors[""], projector)
}

// Start starts the projection manager
// It subscribes to events from the event store and projects them
func (pm *ProjectionManager) Start(ctx context.Context) error {
	pm.mu.Lock()
	if pm.running {
		pm.mu.Unlock()
		return fmt.Errorf("projection manager already running")
	}
	pm.running = true
	pm.mu.Unlock()

	// Create a cancellable context
	subCtx, cancel := context.WithCancel(ctx)
	pm.cancel = cancel

	// Subscribe to all events
	handler := func(eventCtx context.Context, event *ResolvedEvent) error {
		return pm.projectEvent(eventCtx, event)
	}

	go func() {
		if err := pm.store.SubscribeToAll(subCtx, nil, handler); err != nil {
			// Log error but don't stop the manager
			fmt.Printf("projection manager subscription error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the projection manager
func (pm *ProjectionManager) Stop() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.running {
		return fmt.Errorf("projection manager not running")
	}

	if pm.cancel != nil {
		pm.cancel()
	}

	pm.running = false
	return nil
}

// IsRunning returns whether the projection manager is currently running
func (pm *ProjectionManager) IsRunning() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.running
}

// projectEvent handles projecting a single event to all registered projectors
func (pm *ProjectionManager) projectEvent(ctx context.Context, event *ResolvedEvent) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Project to specific event type projectors
	projectors := pm.projectors[event.EventType]

	// Also project to "all events" projectors
	projectors = append(projectors, pm.projectors[""]...)

	for _, projector := range projectors {
		if err := projector.ProjectEvent(ctx, event); err != nil {
			return fmt.Errorf("projection error for %s in %s: %w", event.EventType, projector.GetProjectionName(), err)
		}
	}

	return nil
}

// ReplayProjections replays all events through the projectors
// This is useful for rebuilding projections or testing
func (pm *ProjectionManager) ReplayProjections(ctx context.Context, filter *SubscriptionFilter) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	handler := func(eventCtx context.Context, event *ResolvedEvent) error {
		return pm.projectEvent(eventCtx, event)
	}

	return pm.store.SubscribeToAll(ctx, filter, handler)
}

// GetProjectorCount returns the number of registered projectors
func (pm *ProjectionManager) GetProjectorCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	count := 0
	for _, projectors := range pm.projectors {
		count += len(projectors)
	}
	return count
}

// BaseProjector provides a default implementation of Projector
// Concrete projectors can embed this and override ProjectEvent
type BaseProjector struct {
	name string
}

// NewBaseProjector creates a new base projector
func NewBaseProjector(name string) *BaseProjector {
	return &BaseProjector{
		name: name,
	}
}

// GetProjectionName implements Projector
func (bp *BaseProjector) GetProjectionName() string {
	return bp.name
}

// ProjectEvent is a no-op implementation that can be overridden
func (bp *BaseProjector) ProjectEvent(ctx context.Context, event *ResolvedEvent) error {
	return nil
}

// Reset is a no-op implementation that can be overridden
func (bp *BaseProjector) Reset(ctx context.Context) error {
	return nil
}

// EventTypeFilter is a helper that filters events by type
// Use this with projectors to only process specific event types
func EventTypeFilter(eventType string) func(*ResolvedEvent) bool {
	return func(event *ResolvedEvent) bool {
		return event.EventType == eventType
	}
}

// StreamPrefixFilter is a helper that filters events by stream prefix
// Use this with projectors to only process events from specific streams
func StreamPrefixFilter(prefix string) func(*ResolvedEvent) bool {
	return func(event *ResolvedEvent) bool {
		return len(event.StreamName) >= len(prefix) && event.StreamName[:len(prefix)] == prefix
	}
}
