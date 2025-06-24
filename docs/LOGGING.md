# Logging with Zerolog

This project uses [zerolog](https://github.com/rs/zerolog) for structured logging with environment-specific configurations.

## Features

- **Environment-aware logging**: Different configurations for development and production
- **Structured logging**: JSON format in production, pretty console output in development
- **Context-aware logging**: Add namespace and deployment context to logs
- **Multiple log levels**: Debug, Info, Warn, Error, Fatal
- **Structured fields**: Add custom fields to log messages

## Environment Configuration

The logging behavior is controlled by the `ENV` environment variable:

### Development Environment (default)
```bash
# Default behavior when ENV is not set or set to anything other than "prod"/"production"
export ENV=dev
```

**Features:**
- Debug level logging enabled
- Pretty console output with emojis
- Human-readable timestamps (15:04:05 format)
- Colored output

### Production Environment
```bash
export ENV=prod
# or
export ENV=production
```

**Features:**
- Info level and above only
- JSON format for better parsing
- RFC3339 timestamps
- Structured for log aggregation systems

## Usage

### Basic Logging

```go
import "github.com/yourusername/k8s-controller-tutorial/pkg/logger"

// Create logger instance
log := logger.New()

// Log with different levels
log.Debug("Debug message", map[string]interface{}{
    "component": "api",
    "user_id":   123,
})

log.Info("Application started", map[string]interface{}{
    "version": "1.0.0",
    "port":    8080,
})

log.Warn("High resource usage", map[string]interface{}{
    "cpu_usage": 85.5,
    "memory":    78.2,
})

log.Error("Database connection failed", err, map[string]interface{}{
    "database": "postgres",
    "host":     "localhost",
})

log.Fatal("Critical error", err, map[string]interface{}{
    "component": "auth",
})
```

### Context-Aware Logging

```go
// Add namespace context
namespaceLogger := log.WithNamespace("kube-system")

// Add deployment context
deploymentLogger := namespaceLogger.WithDeployment("nginx-deployment")

// Log with context
deploymentLogger.Info("Deployment updated", map[string]interface{}{
    "ready_replicas":   3,
    "desired_replicas": 3,
})
```

### Advanced Usage

```go
// Access the underlying zerolog logger for advanced features
zerologLogger := log.GetZerologLogger()

zerologLogger.Info().
    Str("custom_field", "value").
    Int("number", 42).
    Float64("percentage", 95.5).
    Msg("Custom structured log")
```

## Log Levels

| Level | Description | Development | Production |
|-------|-------------|-------------|------------|
| Debug | Detailed debugging information | ✅ Visible | ❌ Hidden |
| Info  | General information | ✅ Visible | ✅ Visible |
| Warn  | Warning messages | ✅ Visible | ✅ Visible |
| Error | Error messages | ✅ Visible | ✅ Visible |
| Fatal | Fatal errors (exits application) | ✅ Visible | ✅ Visible |

## Example Output

### Development Environment
```
🔍 [15:04:05] Debug message component=api user_id=123
ℹ️ [15:04:05] Application started version=1.0.0 port=8080
⚠️ [15:04:05] High resource usage cpu_usage=85.5 memory=78.2
❌ [15:04:05] Database connection failed database=postgres host=localhost error="connection timeout"
```

### Production Environment
```json
{"level":"info","service":"k8s-controller","environment":"prod","time":"2024-01-15T10:30:00Z","message":"Application started","version":"1.0.0","port":8080}
{"level":"warn","service":"k8s-controller","environment":"prod","time":"2024-01-15T10:30:01Z","message":"High resource usage","cpu_usage":85.5,"memory":78.2}
{"level":"error","service":"k8s-controller","environment":"prod","time":"2024-01-15T10:30:02Z","message":"Database connection failed","database":"postgres","host":"localhost","error":"connection timeout"}
```

## Running the Demo

To see the logging in action, run the demo:

```bash
# Development mode (default)
go run examples/logging_demo.go

# Production mode
ENV=prod go run examples/logging_demo.go
```

## Integration with Kubernetes Controller

The controller automatically uses structured logging for:

- Application startup and configuration
- Kubernetes client operations
- Deployment status monitoring
- Event watching and processing
- Error handling and recovery

All logs include relevant context like namespace, deployment name, and operation details. 