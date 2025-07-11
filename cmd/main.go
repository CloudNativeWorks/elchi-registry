package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/CloudNativeWorks/elchi-registry/server"
	"github.com/CloudNativeWorks/elchi-registry/service"
	"github.com/CloudNativeWorks/elchi-registry/storage"
)

// Version information
var (
	Version   = "1.0.0"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	// Command line flags
	var (
		grpcPort    = flag.Int("grpc-port", 9090, "gRPC server port")
		showVersion = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("Registry Service\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Time: %s\n", BuildTime)
		os.Exit(0)
	}

	// Initialize logger
	logger := &service.SimpleLogger{}
	logger.Infof("Starting Registry Service v%s", Version)

	// Initialize storage
	logger.Info("Initializing in-memory storage...")
	storageInstance := storage.NewInMemoryStorage()

	// Initialize registry service
	logger.Info("Initializing registry service...")
	registryService := service.NewRegistryService(storageInstance, logger)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start gRPC server in goroutine
	go func() {
		logger.Infof("Starting gRPC server on port %d", *grpcPort)
		if err := server.StartGRPCServer(*grpcPort, registryService, logger); err != nil {
			logger.Errorf("gRPC server error: %v", err)
		}
	}()

	logger.Info("Registry service is ready")
	logger.Infof("gRPC endpoint: localhost:%d", *grpcPort)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Infof("Received signal: %v", sig)

	// Graceful shutdown
	logger.Info("Shutting down...")
	logger.Info("Registry service shutdown completed")
} 