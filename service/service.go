package service

import (
	"context"
	"fmt"
	"log"

	"github.com/CloudNativeWorks/elchi-registry/models"
	"github.com/CloudNativeWorks/elchi-registry/storage"
)

// RegistryService handles all registry operations
type RegistryService struct {
	storage storage.Storage
	logger  Logger
}

// Logger interface for dependency injection
type Logger interface {
	Info(msg string)
	Infof(format string, args ...any)
	Error(msg string)
	Errorf(format string, args ...any)
	Debug(msg string)
	Debugf(format string, args ...any)
}

// NewRegistryService creates a new registry service
func NewRegistryService(storage storage.Storage, logger Logger) *RegistryService {
	return &RegistryService{
		storage: storage,
		logger:  logger,
	}
}

// Controller operations
func (s *RegistryService) RegisterController(ctx context.Context, info *models.ControllerInfo) error {
	s.logger.Infof("Registering controller: %s (%s)", info.ID, info.GRPCAddress)

	if info.ID == "" {
		return fmt.Errorf("controller ID cannot be empty")
	}

	if info.GRPCAddress == "" {
		return fmt.Errorf("controller gRPC address cannot be empty")
	}

	return s.storage.RegisterController(ctx, info)
}

func (s *RegistryService) GetController(ctx context.Context, controllerID string) (*models.ControllerInfo, error) {
	return s.storage.GetController(ctx, controllerID)
}

// Client location operations
func (s *RegistryService) SetClientLocation(ctx context.Context, clientID, controllerID string) error {
	s.logger.Infof("Setting client location: %s -> %s", clientID, controllerID)

	if clientID == "" {
		return fmt.Errorf("client ID cannot be empty")
	}

	if controllerID == "" {
		return fmt.Errorf("controller ID cannot be empty")
	}

	// Verify controller exists
	_, err := s.storage.GetController(ctx, controllerID)
	if err != nil {
		return fmt.Errorf("controller not found: %s", controllerID)
	}

	location := &models.ClientLocation{
		ClientID:     clientID,
		ControllerID: controllerID,
	}

	return s.storage.SetClientLocation(ctx, location)
}

func (s *RegistryService) GetClientLocation(ctx context.Context, clientID string) (*models.ClientLocation, error) {
	clientLocation, err := s.storage.GetClientLocation(ctx, clientID)
	s.logger.Infof("Getting client location: %s, %v", clientLocation, err)
	return clientLocation, err
}

// Simple logger implementation
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string) { log.Printf("[INFO] %s", msg) }
func (l *SimpleLogger) Infof(format string, args ...any) {
	log.Printf("[INFO] "+format, args...)
}
func (l *SimpleLogger) Error(msg string) { log.Printf("[ERROR] %s", msg) }
func (l *SimpleLogger) Errorf(format string, args ...any) {
	log.Printf("[ERROR] "+format, args...)
}
func (l *SimpleLogger) Debug(msg string) { log.Printf("[DEBUG] %s", msg) }
func (l *SimpleLogger) Debugf(format string, args ...any) {
	log.Printf("[DEBUG] "+format, args...)
}
