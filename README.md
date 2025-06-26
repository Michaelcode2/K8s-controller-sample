# Kubernetes Controller Implementation Guide

This guide shows how to implement a Kubernetes controller that monitors deployments and events with structured logging support.

## What We Built

A CLI tool that can:
- ✅ Show deployment status (replicas, availability, readiness)
- ✅ Display recent events in a namespace
- ✅ Watch for real-time deployment changes
- ✅ Connect to any Kubernetes cluster via kubeconfig
- ✅ **Structured logging with environment support**
- ✅ **Context-aware logging (namespace, deployment)**
- ✅ **Production-ready JSON logging**

## Step-by-Step Implementation

### Step 1: Add Kubernetes Dependencies
```bash
go get k8s.io/client-go@latest k8s.io/apimachinery@latest k8s.io/api@latest
go mod tidy
```

### Step 2: Add Logging Dependencies
```bash
go get github.com/rs/zerolog
go mod tidy
```

### Step 3: Create Controller Command
Created `cmd/controller.go` with:
- Kubernetes client setup
- Deployment status monitoring
- Event logging
- Real-time watching capability
- **Structured logging integration**

### Step 4: Key Components

#### 1. Kubernetes Client Setup
```go
func getKubernetesClient() (*kubernetes.Clientset, error) {
    kubeconfig := os.Getenv("KUBECONFIG")
    if kubeconfig == "" {
        kubeconfig = os.Getenv("HOME") + "/.kube/config"
    }
    
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
    }
    
    return kubernetes.NewForConfig(config)
}
```

#### 2. Deployment Status Monitoring
```go
func showDeploymentStatus(clientset *kubernetes.Clientset) {
    deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
    // ... display deployment info
}
```

#### 3. Event Monitoring
```go
func showRecentEvents(clientset *kubernetes.Clientset) {
    events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
        Limit: 10,
    })
    // ... display events
}
```

#### 4. Real-time Watching
```go
func watchDeployments(clientset *kubernetes.Clientset) {
    watcher, err := clientset.AppsV1().Deployments(namespace).Watch(context.TODO(), metav1.ListOptions{})
    // ... watch for changes
}
```

## Logging System

The controller includes a comprehensive logging system built with [zerolog](https://github.com/rs/zerolog) that provides environment-specific configurations.

### Environment Modes

#### Development Mode (Default)
```bash
# Pretty console output with emojis and debug level
./scripts/run_dev.sh controller -n default
```

**Features:**
- Debug level logging enabled
- Pretty console output with emojis
- Human-readable timestamps (15:04:05 format)
- Colored output for better readability

**Example Output:**
```
🔍 [15:04:05] Debug message component=api user_id=123
ℹ️ [15:04:05] Application started version=1.0.0 port=8080
⚠️ [15:04:05] High resource usage cpu_usage=85.5 memory=78.2
❌ [15:04:05] Database connection failed error="connection timeout"
```

#### Production Mode
```bash
# JSON format for log aggregation systems
./scripts/run_prod.sh controller -n default
```

**Features:**
- Info level and above only
- JSON format for better parsing
- RFC3339 timestamps
- Structured for log aggregation systems

**Example Output:**
```json
{"level":"info","service":"k8s-controller","environment":"prod","time":"2024-01-15T10:30:00Z","message":"Application started","version":"1.0.0","port":8080}
{"level":"warn","service":"k8s-controller","environment":"prod","time":"2024-01-15T10:30:01Z","message":"High resource usage","cpu_usage":85.5,"memory":78.2}
```

### Manual Environment Control

```bash
# Development mode
ENV=dev ./controller controller

# Production mode  
ENV=prod ./controller controller
```

### Logging Features

#### 1. Context-Aware Logging
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

#### 2. Multiple Log Levels
```go
log.Debug("Debug message", map[string]interface{}{"component": "api"})
log.Info("Application started", map[string]interface{}{"version": "1.0.0"})
log.Warn("High resource usage", map[string]interface{}{"cpu_usage": 85.5})
log.Error("Database connection failed", err, map[string]interface{}{"database": "postgres"})
log.Fatal("Critical error", err, map[string]interface{}{"component": "auth"})
```

#### 3. Structured Fields
All logs include structured fields for better parsing and filtering:
- `service`: Always set to "k8s-controller"
- `environment`: Current environment (dev/prod)
- `namespace`: Kubernetes namespace (when applicable)
- `deployment`: Deployment name (when applicable)
- `error`: Error details (for error logs)

### Logging Demo

Test the logging system:

```bash
# Development mode
go run examples/logging_demo.go

# Production mode
ENV=prod go run examples/logging_demo.go
```

## Usage Examples

### 1. Show Current Status
```bash
# Show deployments in default namespace
./controller controller

# Show deployments in specific namespace
./controller controller -n kube-system
```

### 2. Watch for Changes
```bash
# Watch deployments in real-time
./controller controller -w

# Watch specific namespace
./controller controller -n my-app -w
```

### 3. Help
```bash
./controller controller --help
```

### 4. Environment-Specific Logging
```bash
# Development mode with detailed logging
./scripts/run_dev.sh controller -n default

# Production mode with JSON logging
./scripts/run_prod.sh controller -n default
```

## Prerequisites

1. **Kubernetes Cluster**: Access to a Kubernetes cluster
2. **kubeconfig**: Properly configured kubeconfig file
3. **Go**: Go 1.24+ installed

## Common Issues & Solutions

### 1. Type Assertion Error
**Error**: `impossible type assertion: event.Object.(*metav1.ObjectMeta)`

**Solution**: Use the correct type:
```go
// ❌ Wrong
deployment := event.Object.(*metav1.ObjectMeta)

// ✅ Correct
deployment := event.Object.(*appsv1.Deployment)
```

### 2. Missing Dependencies
**Error**: Missing go.sum entries

**Solution**: Run `go mod tidy` to download all dependencies

### 3. Connection Issues
**Error**: Failed to load kubeconfig

**Solution**: Ensure your kubeconfig is properly set up:
```bash
export KUBECONFIG=/path/to/your/kubeconfig
# or
kubectl config use-context your-context
```

### 4. Logging Issues
**Error**: No logs appearing in production mode

**Solution**: Check environment variable:
```bash
# Ensure ENV is set correctly
export ENV=prod
./controller controller
```

## Next Steps

You can extend this controller by:

1. **Adding More Resources**: Monitor pods, services, configmaps
2. **Custom Resources**: Watch your own CRDs
3. **Filtering**: Add label selectors and field selectors
4. **Metrics**: Export Prometheus metrics
5. **Webhooks**: Send notifications on events
6. **Persistence**: Store events in a database
7. **Advanced Logging**: Add log rotation, file output, remote logging

## Testing

To test the controller:

1. **Build**: `go build -o controller`
2. **Run**: `./controller controller`
3. **Watch**: `./controller controller -w`
4. **Test Logging**: `go run examples/logging_demo.go`

Make sure you have a Kubernetes cluster running and accessible via your kubeconfig.

## Architecture Benefits

This implementation demonstrates:
- **Separation of Concerns**: Each function has a single responsibility
- **Error Handling**: Proper error handling and user feedback
- **Flexibility**: Easy to extend with new features
- **CLI Design**: Good UX with flags and help text
- **Real-time Monitoring**: Efficient watching mechanism
- **Structured Logging**: Production-ready logging with environment support
- **Context Awareness**: Logs include relevant Kubernetes context

The controller is now ready to use with comprehensive logging and can be extended based on your specific needs! 

## New Features

### 1. kubeconfig Parameter

The controller now supports a `--kubeconfig` flag to specify a custom kubeconfig file. The precedence order for kubeconfig resolution is:
1. CLI flag `--kubeconfig` (highest priority)
2. `KUBECONFIG` environment variable
3. Default `~/.kube/config` (lowest priority)

This gives users full flexibility to specify which Kubernetes cluster configuration to use while maintaining backward compatibility with the existing behavior.

### Usage Examples

#### 1. Use a specific kubeconfig file
```bash
# Use a specific kubeconfig file
./your-app controller --kubeconfig /path/to/custom/kubeconfig

# Use short flag
./your-app controller -k /path/to/custom/kubeconfig

# Use with other flags
./your-app controller --kubeconfig /path/to/kubeconfig --namespace my-namespace --watch
```

#### 2. Show Current Status
```bash
# Show deployments in default namespace
./controller controller

# Show deployments in specific namespace
./controller controller -n kube-system
```

#### 3. Watch for Changes
```bash
# Watch deployments in real-time
./controller controller -w

# Watch specific namespace
./controller controller -n my-app -w
```

#### 4. Help
```bash
./controller controller --help
```

#### 5. Environment-Specific Logging
```bash
# Development mode with detailed logging
./scripts/run_dev.sh controller -n default

# Production mode with JSON logging
./scripts/run_prod.sh controller -n default
``` 