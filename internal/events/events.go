package events

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Message events
	EventMessageReceived EventType = "message.received"
	EventMessageStored   EventType = "message.stored"
	EventMessageFailed   EventType = "message.failed"

	// Email events
	EventEmailSent       EventType = "email.sent"
	EventEmailFailed     EventType = "email.failed"
	EventEmailRetrying   EventType = "email.retrying"

	// Blog events
	EventBlogPublished   EventType = "blog.published"
	EventBlogUpdated     EventType = "blog.updated"
	EventBlogDeleted     EventType = "blog.deleted"

	// Booking events (for future use)
	EventBookingCreated  EventType = "booking.created"
	EventBookingConfirmed EventType = "booking.confirmed"
	EventBookingCancelled EventType = "booking.cancelled"

	// System events
	EventSystemStartup   EventType = "system.startup"
	EventSystemShutdown  EventType = "system.shutdown"
	EventHealthCheck     EventType = "system.health_check"
)

// Event represents an event in the system
type Event interface {
	ID() string
	Type() EventType
	Timestamp() time.Time
	Data() interface{}
	Context() context.Context
}

// EventHandler handles events
type EventHandler func(ctx context.Context, event Event) error

// EventBus manages event publishing and subscriptions
type EventBus interface {
	// Publish publishes an event to all subscribers
	Publish(ctx context.Context, event Event) error
	
	// Subscribe subscribes a handler to an event type
	Subscribe(eventType EventType, handler EventHandler) string
	
	// Unsubscribe removes a subscription
	Unsubscribe(subscriptionID string)
	
	// Start starts the event bus
	Start(ctx context.Context) error
	
	// Stop stops the event bus gracefully
	Stop() error
}

// BaseEvent provides a basic implementation of Event
type BaseEvent struct {
	id        string
	eventType EventType
	timestamp time.Time
	data      interface{}
	ctx       context.Context
}

func (e *BaseEvent) ID() string          { return e.id }
func (e *BaseEvent) Type() EventType     { return e.eventType }
func (e *BaseEvent) Timestamp() time.Time { return e.timestamp }
func (e *BaseEvent) Data() interface{}   { return e.data }
func (e *BaseEvent) Context() context.Context { return e.ctx }

// NewEvent creates a new event with background context  
// Use NewEventWithContext for proper context propagation
func NewEvent(eventType EventType, data interface{}) Event {
	return NewEventWithContext(context.Background(), eventType, data)
}

// NewEventWithContext creates a new event with context
func NewEventWithContext(ctx context.Context, eventType EventType, data interface{}) Event {
	return &BaseEvent{
		id:        generateEventID(),
		eventType: eventType,
		timestamp: time.Now(),
		data:      data,
		ctx:       ctx,
	}
}

// InMemoryEventBus is an in-memory implementation of EventBus
type InMemoryEventBus struct {
	mu            sync.RWMutex
	handlers      map[EventType]map[string]EventHandler
	subscriptions map[string]EventType
	eventQueue    chan Event
	workers       int
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *log.Logger
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus(workers int, logger *log.Logger) *InMemoryEventBus {
	if workers <= 0 {
		workers = 5 // Default workers
	}
	
	if logger == nil {
		logger = log.Default()
	}
	
	return &InMemoryEventBus{
		handlers:      make(map[EventType]map[string]EventHandler),
		subscriptions: make(map[string]EventType),
		eventQueue:    make(chan Event, 1000), // Buffer for 1000 events
		workers:       workers,
		logger:        logger,
	}
}

// Subscribe adds a handler for an event type
func (eb *InMemoryEventBus) Subscribe(eventType EventType, handler EventHandler) string {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	subscriptionID := generateSubscriptionID()
	
	if _, exists := eb.handlers[eventType]; !exists {
		eb.handlers[eventType] = make(map[string]EventHandler)
	}
	
	eb.handlers[eventType][subscriptionID] = handler
	eb.subscriptions[subscriptionID] = eventType
	
	eb.logger.Printf("EVENT_BUS: Subscribed handler %s to event %s", subscriptionID, eventType)
	
	return subscriptionID
}

// Unsubscribe removes a subscription
func (eb *InMemoryEventBus) Unsubscribe(subscriptionID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	if eventType, exists := eb.subscriptions[subscriptionID]; exists {
		delete(eb.handlers[eventType], subscriptionID)
		delete(eb.subscriptions, subscriptionID)
		
		eb.logger.Printf("EVENT_BUS: Unsubscribed handler %s from event %s", subscriptionID, eventType)
	}
}

// Publish publishes an event asynchronously
func (eb *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	// Log the event
	eb.logger.Printf("EVENT_BUS: Publishing event %s (ID: %s)", event.Type(), event.ID())
	
	select {
	case eb.eventQueue <- event:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while publishing event")
	default:
		// Queue is full
		eb.logger.Printf("EVENT_BUS: WARNING - Event queue full, dropping event %s", event.ID())
		return fmt.Errorf("event queue full")
	}
}

// Start starts the event bus workers
func (eb *InMemoryEventBus) Start(ctx context.Context) error {
	eb.ctx, eb.cancel = context.WithCancel(ctx)
	
	eb.logger.Printf("EVENT_BUS: Starting with %d workers", eb.workers)
	
	// Start worker goroutines
	for i := 0; i < eb.workers; i++ {
		eb.wg.Add(1)
		go eb.worker(i)
	}
	
	// Publish startup event
	eb.Publish(ctx, NewEvent(EventSystemStartup, nil))
	
	return nil
}

// Stop stops the event bus gracefully
func (eb *InMemoryEventBus) Stop() error {
	eb.logger.Printf("EVENT_BUS: Shutting down...")
	
	// Publish shutdown event
	eb.Publish(context.Background(), NewEvent(EventSystemShutdown, nil))
	
	// Give time for shutdown event to process
	time.Sleep(100 * time.Millisecond)
	
	// Cancel context to stop workers
	if eb.cancel != nil {
		eb.cancel()
	}
	
	// Close event queue
	close(eb.eventQueue)
	
	// Wait for workers to finish
	eb.wg.Wait()
	
	eb.logger.Printf("EVENT_BUS: Shutdown complete")
	
	return nil
}

// worker processes events from the queue
func (eb *InMemoryEventBus) worker(id int) {
	defer eb.wg.Done()
	
	eb.logger.Printf("EVENT_BUS: Worker %d started", id)
	
	for {
		select {
		case event, ok := <-eb.eventQueue:
			if !ok {
				eb.logger.Printf("EVENT_BUS: Worker %d stopping - queue closed", id)
				return
			}
			
			eb.processEvent(event)
			
		case <-eb.ctx.Done():
			eb.logger.Printf("EVENT_BUS: Worker %d stopping - context cancelled", id)
			return
		}
	}
}

// processEvent processes a single event
func (eb *InMemoryEventBus) processEvent(event Event) {
	eb.mu.RLock()
	handlers, exists := eb.handlers[event.Type()]
	eb.mu.RUnlock()
	
	if !exists || len(handlers) == 0 {
		eb.logger.Printf("EVENT_BUS: No handlers for event %s", event.Type())
		return
	}
	
	// Create a copy of handlers to avoid holding lock during execution
	handlersCopy := make(map[string]EventHandler)
	eb.mu.RLock()
	for id, handler := range handlers {
		handlersCopy[id] = handler
	}
	eb.mu.RUnlock()
	
	// Execute handlers
	for id, handler := range handlersCopy {
		eb.logger.Printf("EVENT_BUS: Executing handler %s for event %s", id, event.ID())
		
		// Run handler with timeout
		ctx, cancel := context.WithTimeout(event.Context(), 30*time.Second)
		err := handler(ctx, event)
		cancel()
		
		if err != nil {
			eb.logger.Printf("EVENT_BUS: ERROR - Handler %s failed for event %s: %v", 
				id, event.ID(), err)
			
			// Could publish a handler failure event here
			// eb.Publish(ctx, NewEvent(EventHandlerFailed, HandlerError{...}))
		} else {
			eb.logger.Printf("EVENT_BUS: Handler %s completed successfully for event %s", 
				id, event.ID())
		}
	}
}

// Helper functions

var (
	eventCounter      uint64
	subscriptionCounter uint64
	counterMu         sync.Mutex
)

func generateEventID() string {
	counterMu.Lock()
	eventCounter++
	count := eventCounter
	counterMu.Unlock()
	
	return fmt.Sprintf("evt_%d_%d", time.Now().Unix(), count)
}

func generateSubscriptionID() string {
	counterMu.Lock()
	subscriptionCounter++
	count := subscriptionCounter
	counterMu.Unlock()
	
	return fmt.Sprintf("sub_%d_%d", time.Now().Unix(), count)
}