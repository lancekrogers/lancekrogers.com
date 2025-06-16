package registry

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"
)

// ServiceStatus represents the health status of a service
type ServiceStatus struct {
	Healthy       bool      `json:"healthy"`
	LastChecked   time.Time `json:"last_checked"`
	Message       string    `json:"message,omitempty"`
	ResponseTime  time.Duration `json:"response_time_ms"`
}

// Service represents a service that can be registered
type Service interface {
	// Name returns the service name
	Name() string
	
	// Health checks the health of the service
	Health(ctx context.Context) error
	
	// Start starts the service
	Start(ctx context.Context) error
	
	// Stop stops the service
	Stop(ctx context.Context) error
}

// ServiceRegistry manages service registration and lifecycle
type ServiceRegistry interface {
	// Register registers a service
	Register(name string, service interface{}) error
	
	// Get retrieves a service by name
	Get(name string) (interface{}, error)
	
	// GetTyped retrieves a typed service (requires type assertion)
	GetTyped(name string, target interface{}) error
	
	// Health returns health status of all services
	Health() map[string]ServiceStatus
	
	// Start starts all registered services
	Start(ctx context.Context) error
	
	// Stop stops all registered services
	Stop(ctx context.Context) error
	
	// List returns all registered service names
	List() []string
}

// DefaultServiceRegistry is the default implementation of ServiceRegistry
type DefaultServiceRegistry struct {
	mu            sync.RWMutex
	services      map[string]interface{}
	healthChecks  map[string]func(context.Context) error
	startOrder    []string
	started       bool
	logger        *log.Logger
	healthMu      sync.RWMutex
	healthStatus  map[string]ServiceStatus
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(logger *log.Logger) *DefaultServiceRegistry {
	if logger == nil {
		logger = log.Default()
	}
	
	return &DefaultServiceRegistry{
		services:     make(map[string]interface{}),
		healthChecks: make(map[string]func(context.Context) error),
		healthStatus: make(map[string]ServiceStatus),
		logger:       logger,
	}
}

// Register registers a service
func (r *DefaultServiceRegistry) Register(name string, service interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.started {
		return fmt.Errorf("cannot register service after registry has started")
	}
	
	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}
	
	r.services[name] = service
	r.startOrder = append(r.startOrder, name)
	
	// If service implements Service interface, register health check
	if svc, ok := service.(Service); ok {
		r.healthChecks[name] = svc.Health
	}
	
	r.logger.Printf("REGISTRY: Registered service '%s' (type: %T)", name, service)
	
	return nil
}

// Get retrieves a service by name
func (r *DefaultServiceRegistry) Get(name string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	service, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	
	return service, nil
}

// GetTyped retrieves a typed service using reflection
func (r *DefaultServiceRegistry) GetTyped(name string, target interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	service, exists := r.services[name]
	if !exists {
		return fmt.Errorf("service %s not found", name)
	}
	
	// Use reflection to set the target
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}
	
	targetElem := targetValue.Elem()
	serviceValue := reflect.ValueOf(service)
	
	// Check if types are compatible
	if !serviceValue.Type().AssignableTo(targetElem.Type()) {
		return fmt.Errorf("service %s (type %T) cannot be assigned to target type %v", 
			name, service, targetElem.Type())
	}
	
	targetElem.Set(serviceValue)
	
	return nil
}

// Health returns health status of all services
func (r *DefaultServiceRegistry) Health() map[string]ServiceStatus {
	r.healthMu.RLock()
	defer r.healthMu.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]ServiceStatus)
	for name, status := range r.healthStatus {
		result[name] = status
	}
	
	return result
}

// Start starts all registered services
func (r *DefaultServiceRegistry) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.started {
		return fmt.Errorf("registry already started")
	}
	
	r.logger.Printf("REGISTRY: Starting %d services...", len(r.startOrder))
	
	// Start services in registration order
	for _, name := range r.startOrder {
		service := r.services[name]
		
		// If service implements Service interface, call Start
		if svc, ok := service.(Service); ok {
			r.logger.Printf("REGISTRY: Starting service '%s'...", name)
			
			if err := svc.Start(ctx); err != nil {
				// Stop already started services
				r.stopStartedServices(ctx, name)
				return fmt.Errorf("failed to start service %s: %w", name, err)
			}
			
			r.logger.Printf("REGISTRY: Service '%s' started successfully", name)
		}
	}
	
	r.started = true
	
	// Start health check goroutine
	go r.healthCheckLoop(ctx)
	
	r.logger.Printf("REGISTRY: All services started successfully")
	
	return nil
}

// Stop stops all registered services
func (r *DefaultServiceRegistry) Stop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if !r.started {
		return nil
	}
	
	r.logger.Printf("REGISTRY: Stopping %d services...", len(r.startOrder))
	
	// Stop services in reverse order
	for i := len(r.startOrder) - 1; i >= 0; i-- {
		name := r.startOrder[i]
		service := r.services[name]
		
		// If service implements Service interface, call Stop
		if svc, ok := service.(Service); ok {
			r.logger.Printf("REGISTRY: Stopping service '%s'...", name)
			
			// Create timeout context for each service stop
			stopCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			err := svc.Stop(stopCtx)
			cancel()
			
			if err != nil {
				r.logger.Printf("REGISTRY: Error stopping service '%s': %v", name, err)
				// Continue stopping other services
			} else {
				r.logger.Printf("REGISTRY: Service '%s' stopped successfully", name)
			}
		}
	}
	
	r.started = false
	r.logger.Printf("REGISTRY: All services stopped")
	
	return nil
}

// List returns all registered service names
func (r *DefaultServiceRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]string, len(r.startOrder))
	copy(result, r.startOrder)
	
	return result
}

// stopStartedServices stops services that were already started (used during failed startup)
func (r *DefaultServiceRegistry) stopStartedServices(ctx context.Context, failedService string) {
	// Find index of failed service
	failedIndex := -1
	for i, name := range r.startOrder {
		if name == failedService {
			failedIndex = i
			break
		}
	}
	
	if failedIndex == -1 {
		return
	}
	
	// Stop services in reverse order up to the failed service
	for i := failedIndex - 1; i >= 0; i-- {
		name := r.startOrder[i]
		service := r.services[name]
		
		if svc, ok := service.(Service); ok {
			r.logger.Printf("REGISTRY: Rolling back - stopping service '%s'...", name)
			
			stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			svc.Stop(stopCtx)
			cancel()
		}
	}
}

// healthCheckLoop runs periodic health checks
func (r *DefaultServiceRegistry) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	// Run initial health check
	r.runHealthChecks(ctx)
	
	for {
		select {
		case <-ticker.C:
			r.runHealthChecks(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// runHealthChecks runs health checks for all services
func (r *DefaultServiceRegistry) runHealthChecks(ctx context.Context) {
	r.mu.RLock()
	healthChecks := make(map[string]func(context.Context) error)
	for name, check := range r.healthChecks {
		healthChecks[name] = check
	}
	r.mu.RUnlock()
	
	for name, check := range healthChecks {
		start := time.Now()
		
		// Run health check with timeout
		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := check(checkCtx)
		cancel()
		
		responseTime := time.Since(start)
		
		status := ServiceStatus{
			Healthy:      err == nil,
			LastChecked:  time.Now(),
			ResponseTime: responseTime,
		}
		
		if err != nil {
			status.Message = err.Error()
		}
		
		r.healthMu.Lock()
		r.healthStatus[name] = status
		r.healthMu.Unlock()
		
		if err != nil {
			r.logger.Printf("REGISTRY: Health check failed for service '%s': %v", name, err)
		}
	}
}