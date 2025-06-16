package registry

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock service for testing
type mockService struct {
	name        string
	started     bool
	stopped     bool
	healthError error
}

func (m *mockService) Name() string {
	return m.name
}

func (m *mockService) Start(ctx context.Context) error {
	m.started = true
	return nil
}

func (m *mockService) Stop(ctx context.Context) error {
	m.stopped = true
	return nil
}

func (m *mockService) Health(ctx context.Context) error {
	return m.healthError
}

func TestServiceRegistry(t *testing.T) {
	// Create a test logger that discards output
	logger := log.New(&testWriter{}, "", 0)
	
	t.Run("register and get service", func(t *testing.T) {
		reg := NewServiceRegistry(logger)
		
		// Register a service
		service := &mockService{name: "test-service"}
		err := reg.Register("test", service)
		require.NoError(t, err)
		
		// Get the service
		retrieved, err := reg.Get("test")
		require.NoError(t, err)
		assert.Equal(t, service, retrieved)
		
		// List services
		list := reg.List()
		assert.Equal(t, []string{"test"}, list)
	})
	
	t.Run("get non-existent service", func(t *testing.T) {
		reg := NewServiceRegistry(logger)
		
		_, err := reg.Get("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
	
	t.Run("duplicate registration", func(t *testing.T) {
		reg := NewServiceRegistry(logger)
		
		service1 := &mockService{name: "service1"}
		service2 := &mockService{name: "service2"}
		
		// First registration should succeed
		err := reg.Register("test", service1)
		require.NoError(t, err)
		
		// Second registration with same name should fail
		err = reg.Register("test", service2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
	
	t.Run("start and stop services", func(t *testing.T) {
		reg := NewServiceRegistry(logger)
		ctx := context.Background()
		
		// Register multiple services
		service1 := &mockService{name: "service1"}
		service2 := &mockService{name: "service2"}
		
		err := reg.Register("svc1", service1)
		require.NoError(t, err)
		
		err = reg.Register("svc2", service2)
		require.NoError(t, err)
		
		// Start all services
		err = reg.Start(ctx)
		require.NoError(t, err)
		
		assert.True(t, service1.started)
		assert.True(t, service2.started)
		
		// Stop all services
		err = reg.Stop(ctx)
		require.NoError(t, err)
		
		assert.True(t, service1.stopped)
		assert.True(t, service2.stopped)
	})
	
	t.Run("health checks", func(t *testing.T) {
		reg := NewServiceRegistry(logger)
		ctx := context.Background()
		
		// Register services with different health states
		healthyService := &mockService{name: "healthy"}
		unhealthyService := &mockService{name: "unhealthy", healthError: fmt.Errorf("service is down")}
		
		err := reg.Register("healthy", healthyService)
		require.NoError(t, err)
		
		err = reg.Register("unhealthy", unhealthyService)
		require.NoError(t, err)
		
		// Start services
		err = reg.Start(ctx)
		require.NoError(t, err)
		defer reg.Stop(ctx)
		
		// Wait for health checks to run
		time.Sleep(100 * time.Millisecond)
		
		// Check health status
		health := reg.Health()
		
		assert.True(t, health["healthy"].Healthy)
		assert.False(t, health["unhealthy"].Healthy)
		assert.Contains(t, health["unhealthy"].Message, "service is down")
	})
	
	t.Run("typed get", func(t *testing.T) {
		reg := NewServiceRegistry(logger)
		
		// Register a service
		service := &mockService{name: "typed-service"}
		err := reg.Register("typed", service)
		require.NoError(t, err)
		
		// Get typed service
		var retrieved *mockService
		err = reg.GetTyped("typed", &retrieved)
		require.NoError(t, err)
		assert.Equal(t, service, retrieved)
		
		// Try to get with wrong type
		var wrongType string
		err = reg.GetTyped("typed", &wrongType)
		assert.Error(t, err)
	})
}

// testWriter is a writer that discards all output
type testWriter struct{}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}