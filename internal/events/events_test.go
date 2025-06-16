package events

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventBus(t *testing.T) {
	// Create a test logger that discards output
	logger := log.New(&testWriter{}, "", 0)
	bus := NewInMemoryEventBus(2, logger)
	
	ctx := context.Background()
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer bus.Stop()
	
	t.Run("publish and subscribe", func(t *testing.T) {
		var received Event
		var wg sync.WaitGroup
		wg.Add(1)
		
		// Subscribe to test event
		bus.Subscribe("test.event", func(ctx context.Context, event Event) error {
			received = event
			wg.Done()
			return nil
		})
		
		// Publish event
		testData := map[string]string{"key": "value"}
		event := NewEvent("test.event", testData)
		err := bus.Publish(ctx, event)
		require.NoError(t, err)
		
		// Wait for event to be processed
		wg.Wait()
		
		assert.NotNil(t, received)
		assert.Equal(t, "test.event", string(received.Type()))
		assert.Equal(t, testData, received.Data())
	})
	
	t.Run("multiple subscribers", func(t *testing.T) {
		var count int
		var mu sync.Mutex
		var wg sync.WaitGroup
		wg.Add(3)
		
		// Subscribe multiple handlers
		for i := 0; i < 3; i++ {
			bus.Subscribe("multi.event", func(ctx context.Context, event Event) error {
				mu.Lock()
				count++
				mu.Unlock()
				wg.Done()
				return nil
			})
		}
		
		// Publish event
		event := NewEvent("multi.event", nil)
		err := bus.Publish(ctx, event)
		require.NoError(t, err)
		
		// Wait for all handlers
		wg.Wait()
		
		assert.Equal(t, 3, count)
	})
	
	t.Run("unsubscribe", func(t *testing.T) {
		var count int
		var mu sync.Mutex
		
		// Subscribe handler
		subID := bus.Subscribe("unsub.event", func(ctx context.Context, event Event) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		})
		
		// Publish first event
		event1 := NewEvent("unsub.event", nil)
		bus.Publish(ctx, event1)
		
		// Wait a bit for processing
		time.Sleep(100 * time.Millisecond)
		
		// Unsubscribe
		bus.Unsubscribe(subID)
		
		// Publish second event
		event2 := NewEvent("unsub.event", nil)
		bus.Publish(ctx, event2)
		
		// Wait a bit for processing
		time.Sleep(100 * time.Millisecond)
		
		// Should have received only one event
		mu.Lock()
		finalCount := count
		mu.Unlock()
		
		assert.Equal(t, 1, finalCount)
	})
	
	t.Run("handler error", func(t *testing.T) {
		var successCount int
		var mu sync.Mutex
		var wg sync.WaitGroup
		wg.Add(1)
		
		// Subscribe failing handler
		bus.Subscribe("error.event", func(ctx context.Context, event Event) error {
			return assert.AnError
		})
		
		// Subscribe successful handler
		bus.Subscribe("error.event", func(ctx context.Context, event Event) error {
			mu.Lock()
			successCount++
			mu.Unlock()
			wg.Done()
			return nil
		})
		
		// Publish event
		event := NewEvent("error.event", nil)
		err := bus.Publish(ctx, event)
		require.NoError(t, err)
		
		// Wait for successful handler
		wg.Wait()
		
		// Successful handler should still run
		assert.Equal(t, 1, successCount)
	})
}

// testWriter is a writer that discards all output
type testWriter struct{}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}