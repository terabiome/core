package operations

import (
	"fmt"
	"sync"
)

type Factory func() Operation

type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

func (r *Registry) Register(name string, factory Factory) error {
	if name == "" {
		return fmt.Errorf("operation name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("operation factory for %q cannot be nil", name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("operation %q is already registered", name)
	}
	r.factories[name] = factory
	return nil
}

func (r *Registry) Get(name string) (Operation, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("operation %q not found in registry", name)
	}
	return factory(), nil
}
