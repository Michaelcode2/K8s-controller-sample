# Kubernetes Controller Implementation Guide

This guide shows how to implement a Kubernetes controller that monitors deployments and events using efficient informers with structured logging support.

## What We Built

A CLI tool that can:
- ✅ Monitor deployments with efficient informers (local caching, event deduplication)
- ✅ Display recent events in a namespace
- ✅ Real-time deployment monitoring with detailed replica information
- ✅ Connect to any Kubernetes cluster via kubeconfig or in-cluster authentication
- ✅ **Structured logging with environment support**
- ✅ **Context-aware logging (namespace, deployment)**
- ✅ **Production-ready JSON logging**
- ✅ **In-cluster authentication support**

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
- Kubernetes client setup with in-cluster and kubeconfig support
- Deployment status monitoring
- Event logging
- **Informer-based real-time monitoring**
- **Structured logging integration**

### Step 4: Key Components

#### 1. Kubernetes Client Setup (Enhanced)
```go
func getKubernetesClient() (*kubernetes.Clientset, error) {
    var config *rest.Config
    var err error

    if inCluster {
        config, err = rest.InClusterConfig()
    } else {
        kubeconfigPath := kubeconfig
        if kubeconfigPath == "" {
            kubeconfigPath = os.Getenv("KUBECONFIG")
        }
        if kubeconfigPath == "" {
            kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
        }
        
        config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
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

#### 4. Informer-based Real-time Monitoring
```go
func watchDeploymentsWithInformer(clientset *kubernetes.Clientset) {
    factory := informers.NewSharedInformerFactoryWithOptions(
        clientset,
        30*time.Second, // resync period
        informers.WithNamespace(namespace),
    )
    
    deploymentInformer := factory.Apps().V1().Deployments().Informer()
    deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
        AddFunc: func(obj interface{}) {
            // Handle deployment creation
        },
        UpdateFunc: func(oldObj, newObj interface{}) {
            // Handle deployment updates with detailed replica changes
        },
        DeleteFunc: func(obj interface{}) {
            // Handle deployment deletion
        },
    })
    
    factory.Start(ctx.Done())
    factory.WaitForCacheSync(ctx.Done())
}
```

## Authentication Support

The controller supports two authentication methods:

### 1. In-Cluster Authentication
Use when running the controller as a pod inside a Kubernetes cluster:
```bash
./k8s-controller-tutorial controller -i
```

### 2. Kubeconfig Authentication
Use for external cluster access:
```bash
# Use default kubeconfig
./k8s-controller-tutorial controller

# Use specific kubeconfig file
./k8s-controller-tutorial controller -k /path/to/kubeconfig

# Use environment variable
export KUBECONFIG=/path/to/kubeconfig
./k8s-controller-tutorial controller
```

**Kubeconfig precedence order:**
1. CLI flag `--kubeconfig` (highest priority)
2. `KUBECONFIG` environment variable
3. Default `~/.kube/config` (lowest priority)

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
- `auth_method`: Authentication method used (in_cluster/kubeconfig)

### Logging Demo

Test the logging system:

```bash
# Development mode
go run examples/logging_demo.go

# Production mode
ENV=prod go run examples/logging_demo.go
```

## Usage Examples

### 1. Basic Usage (Informer is Default)
```bash
# Monitor deployments in default namespace with informer
./k8s-controller-tutorial controller

# Monitor deployments in specific namespace
./k8s-controller-tutorial controller -n kube-system
```

### 2. Authentication Examples
```bash
# Use in-cluster authentication (when running as a pod)
./k8s-controller-tutorial controller -i

# Use specific kubeconfig file
./k8s-controller-tutorial controller -k /path/to/kubeconfig

# Use environment variable for kubeconfig
export KUBECONFIG=/path/to/kubeconfig
./k8s-controller-tutorial controller
```

### 3. Combined Examples
```bash
# Monitor specific namespace with in-cluster auth
./k8s-controller-tutorial controller -n my-app -i

# Monitor with custom kubeconfig and namespace
./k8s-controller-tutorial controller -k ~/.kube/config-prod -n production
```

### 4. Help
```bash
./k8s-controller-tutorial controller --help
```

### 5. Environment-Specific Logging
```bash
# Development mode with detailed logging
./scripts/run_dev.sh controller -n default

# Production mode with JSON logging
./scripts/run_prod.sh controller -n default
```

## Informer Benefits

The controller now uses Kubernetes informers by default, providing:

### 1. **Performance Benefits**
- **Local Caching**: Reduces API server load
- **Efficient Watching**: Uses watch connections instead of polling
- **Shared Resources**: Multiple components can share the same informer factory

### 2. **Reliability Features**
- **Event Deduplication**: Prevents duplicate events
- **Resync Capability**: 30-second resync period to catch missed events
- **Better Error Handling**: Automatic reconnection on connection loss
- **Cache Sync**: Waits for cache to sync before processing events

### 3. **Enhanced Event Information**
The informer provides detailed event information:

```
[15:04:05] ADDED: nginx-deployment (0/3 replicas)
[15:04:05] MODIFIED: nginx-deployment (0/3 -> 2/3 replicas)
[15:04:05] MODIFIED: nginx-deployment (2/3 -> 3/3 replicas)
[15:04:05] DELETED: old-deployment
```

## Prerequisites

1. **Kubernetes Cluster**: Access to a Kubernetes cluster
2. **Authentication**: Either kubeconfig file or in-cluster service account
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

### 4. In-Cluster Authentication Issues
**Error**: Failed to load in-cluster config

**Solution**: Ensure the controller is running inside a Kubernetes cluster with proper service account:
```bash
# Check if running in cluster
kubectl get pods -n your-namespace

# Verify service account exists
kubectl get serviceaccount -n your-namespace
```

### 5. Logging Issues
**Error**: No logs appearing in production mode

**Solution**: Check environment variable:
```bash
# Ensure ENV is set correctly
export ENV=prod
./k8s-controller-tutorial controller
```

## Next Steps

You can extend this controller by:

1. **Adding More Resources**: Monitor pods, services, configmaps with informers
2. **Custom Resources**: Watch your own CRDs with custom informers
3. **Filtering**: Add label selectors and field selectors
4. **Metrics**: Export Prometheus metrics
5. **Webhooks**: Send notifications on events
6. **Persistence**: Store events in a database
7. **Advanced Logging**: Add log rotation, file output, remote logging
8. **Multi-Namespace**: Monitor multiple namespaces simultaneously

## Testing

To test the controller:

1. **Build**: `go build -o k8s-controller-tutorial`
2. **Run with kubeconfig**: `./k8s-controller-tutorial controller`
3. **Run with in-cluster**: `./k8s-controller-tutorial controller -i`
4. **Test Logging**: `go run examples/logging_demo.go`

Make sure you have a Kubernetes cluster running and accessible via your chosen authentication method.

## Architecture Benefits

This implementation demonstrates:
- **Separation of Concerns**: Each function has a single responsibility
- **Error Handling**: Proper error handling and user feedback
- **Flexibility**: Easy to extend with new features
- **CLI Design**: Good UX with flags and help text
- **Real-time Monitoring**: Efficient informer-based watching mechanism
- **Structured Logging**: Production-ready logging with environment support
- **Context Awareness**: Logs include relevant Kubernetes context
- **Authentication Flexibility**: Support for both in-cluster and external access
- **Performance**: Local caching and event deduplication

The controller is now ready to use with comprehensive logging, efficient informers, and flexible authentication options! 