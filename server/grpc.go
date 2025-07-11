package server

import (
	"context"
	"fmt"
	"net"

	pb "github.com/CloudNativeWorks/elchi-proto/client"
	"github.com/CloudNativeWorks/elchi-registry/models"
	"github.com/CloudNativeWorks/elchi-registry/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegistryGRPCServer implements the gRPC registry service
type RegistryGRPCServer struct {
	pb.UnimplementedControllerServiceServer
	registryService *service.RegistryService
	logger          service.Logger
}

// NewRegistryGRPCServer creates a new gRPC server instance
func NewRegistryGRPCServer(registryService *service.RegistryService, logger service.Logger) *RegistryGRPCServer {
	return &RegistryGRPCServer{
		registryService: registryService,
		logger:          logger,
	}
}

// RegisterController handles controller registration
func (s *RegistryGRPCServer) RegisterController(ctx context.Context, req *pb.ControllerInfo) (*pb.ControllerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Convert proto to internal model
	controllerInfo := &models.ControllerInfo{
		ID:          req.ControllerId,
		GRPCAddress: req.GrpcAddress,
	}

	if err := s.registryService.RegisterController(ctx, controllerInfo); err != nil {
		s.logger.Errorf("Failed to register controller %s: %v", req.ControllerId, err)
		return &pb.ControllerResponse{
			Success: "failed: " + err.Error(),
		}, nil
	}

	s.logger.Infof("Controller registered successfully: %s", req.ControllerId)
	return &pb.ControllerResponse{
		Success: "controller registered successfully",
	}, nil
}

// GetClientLocation handles client location queries
func (s *RegistryGRPCServer) GetClientLocation(ctx context.Context, req *pb.ClientLocationRequest) (*pb.ClientLocationResponse, error) {
	if req == nil || req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "client ID cannot be empty")
	}

	location, err := s.registryService.GetClientLocation(ctx, req.ClientId)
	if err != nil {
		s.logger.Debugf("Client location not found: %s", req.ClientId)
		return &pb.ClientLocationResponse{
			Found: false,
		}, nil
	}

	// Get controller GRPC address
	controller, err := s.registryService.GetController(ctx, location.ControllerID)
	if err != nil {
		s.logger.Errorf("Controller not found for client %s: %v", req.ClientId, err)
		return &pb.ClientLocationResponse{
			Found: false,
		}, nil
	}

	return &pb.ClientLocationResponse{
		Found:          true,
		ControllerId:   location.ControllerID,
		ControllerFqdn: controller.GRPCAddress,
	}, nil
}

// SetClientLocation handles setting client location
func (s *RegistryGRPCServer) SetClientLocation(ctx context.Context, req *pb.SetClientLocationRequest) (*pb.SetClientLocationResponse, error) {
	if req == nil || req.ClientId == "" || req.ControllerId == "" {
		return nil, status.Error(codes.InvalidArgument, "client ID and controller ID cannot be empty")
	}

	if err := s.registryService.SetClientLocation(ctx, req.ClientId, req.ControllerId); err != nil {
		s.logger.Errorf("Failed to set client location for %s: %v", req.ClientId, err)
		return &pb.SetClientLocationResponse{
			Success: "failed: " + err.Error(),
		}, nil
	}

	s.logger.Infof("Client location set: %s -> %s", req.ClientId, req.ControllerId)
	return &pb.SetClientLocationResponse{
		Success: "client location set successfully",
	}, nil
}

// RequestClientRefresh asks a controller to refresh its client list
func (s *RegistryGRPCServer) RequestClientRefresh(ctx context.Context, req *pb.ClientRefreshRequest) (*pb.ClientRefreshResponse, error) {
	if req == nil || req.ControllerId == "" {
		return nil, status.Error(codes.InvalidArgument, "controller ID cannot be empty")
	}

	// Verify controller exists
	controller, err := s.registryService.GetController(ctx, req.ControllerId)
	if err != nil {
		s.logger.Errorf("Controller not found for refresh request: %s", req.ControllerId)
		return &pb.ClientRefreshResponse{
			Success:     "failed: controller not found",
			ClientCount: 0,
		}, nil
	}

	// TODO: This would actually call the controller to refresh its clients
	// For now, just log and return success
	s.logger.Infof("Client refresh requested for controller %s at %s", controller.ID, controller.GRPCAddress)
	
	return &pb.ClientRefreshResponse{
		Success:     "refresh request sent",
		ClientCount: 0, // Will be updated when controller responds
	}, nil
}

// StartGRPCServer starts the gRPC server
func StartGRPCServer(port int, registryService *service.RegistryService, logger service.Logger) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer()
	
	// Register service
	registryGRPCServer := NewRegistryGRPCServer(registryService, logger)
	pb.RegisterControllerServiceServer(grpcServer, registryGRPCServer)

	logger.Infof("gRPC server starting on port %d", port)
	
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}


// IsControllerRegistered checks if a controller is registered
func (s *RegistryGRPCServer) IsControllerRegistered(ctx context.Context, req *pb.IsControllerRegisteredRequest) (*pb.IsControllerRegisteredResponse, error) {
	if req == nil || req.ControllerId == "" {
		return nil, status.Error(codes.InvalidArgument, "controller ID cannot be empty")
	}

	// Check if controller exists
	_, err := s.registryService.GetController(ctx, req.ControllerId)
	if err != nil {
		s.logger.Debugf("Controller registration check failed: %s", req.ControllerId)
		return &pb.IsControllerRegisteredResponse{
			Registered: false,
		}, nil
	}

	return &pb.IsControllerRegisteredResponse{
		Registered: true,
	}, nil
}

// BulkSetClientLocations sets multiple client locations efficiently
func (s *RegistryGRPCServer) BulkSetClientLocations(ctx context.Context, req *pb.BulkSetClientLocationsRequest) (*pb.BulkSetClientLocationsResponse, error) {
	if req == nil || req.ControllerId == "" {
		return nil, status.Error(codes.InvalidArgument, "controller ID cannot be empty")
	}

	if len(req.ClientIds) == 0 {
		return &pb.BulkSetClientLocationsResponse{
			Success:      true,
			Error:        "",
			UpdatedCount: 0,
		}, nil
	}

	// Verify controller exists
	_, err := s.registryService.GetController(ctx, req.ControllerId)
	if err != nil {
		s.logger.Errorf("Controller not found for bulk client update: %s", req.ControllerId)
		return &pb.BulkSetClientLocationsResponse{
			Success:      false,
			Error:        "controller not found",
			UpdatedCount: 0,
		}, nil
	}

	// Set each client location
	successCount := int32(0)
	for _, clientID := range req.ClientIds {
		if err := s.registryService.SetClientLocation(ctx, clientID, req.ControllerId); err != nil {
			s.logger.Errorf("Failed to set location for client %s: %v", clientID, err)
			// Continue with other clients, don't fail completely
		} else {
			successCount++
		}
	}

	s.logger.Infof("Bulk client location update: %d/%d clients updated for controller %s", 
		successCount, len(req.ClientIds), req.ControllerId)

	return &pb.BulkSetClientLocationsResponse{
		Success:      true,
		Error:        "",
		UpdatedCount: successCount,
	}, nil
}