package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stacklok/toolhive/pkg/registryapi"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	// Version information set at build time
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var (
		port             = flag.Int("port", 8080, "Port to listen on")
		registryName     = flag.String("registry-name", "", "Name of the MCPRegistry resource to serve")
		registryNS       = flag.String("registry-namespace", "", "Namespace of the MCPRegistry resource")
		metricsAddr      = flag.String("metrics-bind-address", ":8081", "The address the metric endpoint binds to")
		enableLeaderElection = flag.Bool("leader-elect", false, "Enable leader election for controller manager")
		printVersion     = flag.Bool("version", false, "Print version information and exit")
	)
	flag.Parse()

	// Print version information if requested
	if *printVersion {
		fmt.Printf("thv-registry-api\n")
		fmt.Printf("  Version: %s\n", version)
		fmt.Printf("  Commit:  %s\n", commit)
		fmt.Printf("  Date:    %s\n", date)
		os.Exit(0)
	}

	// Validate required parameters
	if *registryName == "" {
		fmt.Fprintf(os.Stderr, "Error: --registry-name is required\n")
		os.Exit(1)
	}
	if *registryNS == "" {
		fmt.Fprintf(os.Stderr, "Error: --registry-namespace is required\n")
		os.Exit(1)
	}

	// Setup logging
	opts := zap.Options{
		Development: true,
	}
	logger := zap.New(zap.UseFlagOptions(&opts))
	log.SetLogger(logger)

	setupLog := log.Log.WithName("setup")
	setupLog.Info("Starting thv-registry-api",
		"version", version,
		"commit", commit,
		"date", date,
		"registry-name", *registryName,
		"registry-namespace", *registryNS,
		"port", *port,
	)

	// Create server configuration
	config := &registryapi.ServerConfig{
		Port:              *port,
		RegistryName:      *registryName,
		RegistryNamespace: *registryNS,
		MetricsAddr:       *metricsAddr,
		EnableLeaderElection: *enableLeaderElection,
	}

	// Create and start the server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server, err := registryapi.NewServer(config)
	if err != nil {
		setupLog.Error(err, "failed to create server")
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		setupLog.Info("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	// Start the server
	if err := server.Start(ctx); err != nil {
		setupLog.Error(err, "failed to start server")
		os.Exit(1)
	}

	setupLog.Info("Server shutdown complete")
}