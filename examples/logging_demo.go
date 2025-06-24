package main

import (
	"errors"
	"time"

	"github.com/yourusername/k8s-controller-tutorial/pkg/logger"
)

func main() {
	// Create logger instance
	log := logger.New()

	// Demo different log levels
	log.Debug("This is a debug message", map[string]interface{}{
		"component": "demo",
		"timestamp": time.Now(),
	})

	log.Info("Application started successfully", map[string]interface{}{
		"version": "1.0.0",
		"port":    8080,
	})

	log.Warn("Resource usage is high", map[string]interface{}{
		"cpu_usage":    85.5,
		"memory_usage": 78.2,
		"threshold":    80.0,
	})

	// Simulate an error
	err := errors.New("connection timeout")
	log.Error("Failed to connect to database", err, map[string]interface{}{
		"database": "postgres",
		"host":     "localhost",
		"port":     5432,
	})

	// Demo with namespace context
	namespaceLogger := log.WithNamespace("kube-system")
	namespaceLogger.Info("Monitoring namespace", map[string]interface{}{
		"pod_count": 15,
		"services":  8,
	})

	// Demo with deployment context
	deploymentLogger := namespaceLogger.WithDeployment("nginx-deployment")
	deploymentLogger.Info("Deployment status updated", map[string]interface{}{
		"ready_replicas":   3,
		"desired_replicas": 3,
		"status":           "healthy",
	})

	// Show how to use the underlying zerolog logger for advanced usage
	zerologLogger := log.GetZerologLogger()
	zerologLogger.Info().
		Str("custom_field", "custom_value").
		Int("custom_number", 42).
		Msg("Using zerolog directly")
}
