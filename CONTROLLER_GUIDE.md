# Kubernetes Controller Implementation Guide

This guide shows how to implement a Kubernetes controller that monitors deployments and events.

## What We Built

A CLI tool that can:
- ✅ Show deployment status (replicas, availability, readiness)
- ✅ Display recent events in a namespace
- ✅ Watch for real-time deployment changes
- ✅ Connect to any Kubernetes cluster via kubeconfig

## Step-by-Step Implementation

### Step 1: Add Kubernetes Dependencies
```bash
go get k8s.io/client-go@latest k8s.io/apimachinery@latest k8s.io/api@latest
go mod tidy
```

### Step 2: Create Controller Command
Created `cmd/controller.go` with:
- Kubernetes client setup
- Deployment status monitoring
- Event logging
- Real-time watching capability

### Step 3: Key Components

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

## Next Steps

You can extend this controller by:

1. **Adding More Resources**: Monitor pods, services, configmaps
2. **Custom Resources**: Watch your own CRDs
3. **Filtering**: Add label selectors and field selectors
4. **Metrics**: Export Prometheus metrics
5. **Webhooks**: Send notifications on events
6. **Persistence**: Store events in a database

## Testing

To test the controller:

1. **Build**: `go build -o controller`
2. **Run**: `./controller controller`
3. **Watch**: `./controller controller -w`

Make sure you have a Kubernetes cluster running and accessible via your kubeconfig.

## Architecture Benefits

This implementation demonstrates:
- **Separation of Concerns**: Each function has a single responsibility
- **Error Handling**: Proper error handling and user feedback
- **Flexibility**: Easy to extend with new features
- **CLI Design**: Good UX with flags and help text
- **Real-time Monitoring**: Efficient watching mechanism

The controller is now ready to use and can be extended based on your specific needs! 