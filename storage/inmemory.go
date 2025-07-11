package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/CloudNativeWorks/elchi-registry/models"
)

// InMemoryStorage implements Storage interface using in-memory maps
type InMemoryStorage struct {
	controllers     map[string]*models.ControllerInfo
	controllersMux  sync.RWMutex
	clientLocations map[string]*models.ClientLocation
	clientsMux      sync.RWMutex
}

// NewInMemoryStorage creates a new in-memory storage instance
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		controllers:     make(map[string]*models.ControllerInfo),
		controllersMux:  sync.RWMutex{},
		clientLocations: make(map[string]*models.ClientLocation),
		clientsMux:      sync.RWMutex{},
	}
}

// Controller operations
func (s *InMemoryStorage) RegisterController(ctx context.Context, controller *models.ControllerInfo) error {
	s.controllersMux.Lock()
	defer s.controllersMux.Unlock()

	// Copy to avoid external modifications
	controllerCopy := *controller
	s.controllers[controller.ID] = &controllerCopy

	return nil
}

func (s *InMemoryStorage) GetController(ctx context.Context, controllerID string) (*models.ControllerInfo, error) {
	s.controllersMux.RLock()
	defer s.controllersMux.RUnlock()

	controller, exists := s.controllers[controllerID]
	if !exists {
		return nil, fmt.Errorf("controller not found: %s", controllerID)
	}

	// Return a copy to avoid external modifications
	controllerCopy := *controller
	return &controllerCopy, nil
}

// Client location operations
func (s *InMemoryStorage) SetClientLocation(ctx context.Context, location *models.ClientLocation) error {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()

	// Copy to avoid external modifications
	locationCopy := *location
	s.clientLocations[location.ClientID] = &locationCopy

	return nil
}

func (s *InMemoryStorage) GetClientLocation(ctx context.Context, clientID string) (*models.ClientLocation, error) {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()

	location, exists := s.clientLocations[clientID]
	if !exists {
		return nil, fmt.Errorf("client location not found: %s", clientID)
	}

	// Return a copy to avoid external modifications
	locationCopy := *location
	return &locationCopy, nil
} 