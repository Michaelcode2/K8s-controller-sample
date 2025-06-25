package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	serverPort int
	serverHost string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP server for Kubernetes controller",
	Long: `Start a FastHTTP server that exposes Kubernetes controller functionality
via HTTP endpoints. The server provides REST API access to deployment status,
events, and real-time monitoring.`,
	Run: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Port to listen on")
	serverCmd.Flags().StringVarP(&serverHost, "host", "H", "0.0.0.0", "Host to bind to")
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// DeploymentStatus represents deployment status information
type DeploymentStatus struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	ReadyReplicas     int32  `json:"ready_replicas"`
	DesiredReplicas   int32  `json:"desired_replicas"`
	AvailableReplicas int32  `json:"available_replicas"`
	UpdatedReplicas   int32  `json:"updated_replicas"`
	Healthy           bool   `json:"healthy"`
}

// Event represents a Kubernetes event
type Event struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Object    string    `json:"object"`
}

func runServer(cmd *cobra.Command, args []string) {
	log.Info("Starting HTTP server", map[string]interface{}{
		"host": serverHost,
		"port": serverPort,
	})

	clientset, err := getKubernetesClient()
	if err != nil {
		log.Fatal("Failed to get Kubernetes client", err, nil)
	}

	// Create FastHTTP server
	server := &fasthttp.Server{
		Handler: createHandler(clientset),
		Name:    "k8s-controller-server",
	}

	// Create listener
	addr := fmt.Sprintf("%s:%d", serverHost, serverPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Failed to create listener", err, map[string]interface{}{
			"address": addr,
		})
	}

	log.Info("HTTP server started successfully", map[string]interface{}{
		"address": addr,
	})

	// Start server
	if err := server.Serve(listener); err != nil {
		log.Fatal("Server error", err, nil)
	}
}

func createHandler(clientset *kubernetes.Clientset) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Set CORS headers
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if ctx.IsOptions() {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}

		// Set content type
		ctx.Response.Header.Set("Content-Type", "application/json")

		path := string(ctx.Path())
		method := string(ctx.Method())

		log.Debug("HTTP request", map[string]interface{}{
			"method": method,
			"path":   path,
			"remote": ctx.RemoteAddr(),
		})

		switch {
		case path == "/health" && method == "GET":
			handleHealth(ctx)
		case path == "/api/v1/deployments" && method == "GET":
			handleGetDeployments(ctx, clientset)
		case path == "/api/v1/deployments" && method == "POST":
			handleWatchDeployments(ctx)
		case path == "/api/v1/events" && method == "GET":
			handleGetEvents(ctx, clientset)
		case path == "/api/v1/status" && method == "GET":
			handleGetStatus(ctx, clientset)
		default:
			handleNotFound(ctx)
		}
	}
}

func handleHealth(ctx *fasthttp.RequestCtx) {
	response := Response{
		Success: true,
		Message: "Server is healthy",
		Data: map[string]interface{}{
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		},
	}

	jsonResponse, _ := json.Marshal(response)
	ctx.SetBody(jsonResponse)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func handleGetDeployments(ctx *fasthttp.RequestCtx, clientset *kubernetes.Clientset) {
	namespace := string(ctx.QueryArgs().Peek("namespace"))
	if namespace == "" {
		namespace = "default"
	}

	namespaceLogger := log.WithNamespace(namespace)
	namespaceLogger.Info("HTTP request: Get deployments", map[string]interface{}{
		"namespace": namespace,
	})

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		namespaceLogger.Error("Failed to get deployments", err, nil)
		sendErrorResponse(ctx, "Failed to get deployments", err, fasthttp.StatusInternalServerError)
		return
	}

	var deploymentStatuses []DeploymentStatus
	for _, deployment := range deployments.Items {
		readyReplicas := deployment.Status.ReadyReplicas
		desiredReplicas := *deployment.Spec.Replicas
		availableReplicas := deployment.Status.AvailableReplicas

		status := DeploymentStatus{
			Name:              deployment.Name,
			Namespace:         deployment.Namespace,
			ReadyReplicas:     readyReplicas,
			DesiredReplicas:   desiredReplicas,
			AvailableReplicas: availableReplicas,
			UpdatedReplicas:   deployment.Status.UpdatedReplicas,
			Healthy:           readyReplicas >= desiredReplicas,
		}

		deploymentStatuses = append(deploymentStatuses, status)

		// Log deployment status
		deploymentLogger := namespaceLogger.WithDeployment(deployment.Name)
		deploymentLogger.Info("Deployment status retrieved", map[string]interface{}{
			"ready_replicas":     readyReplicas,
			"desired_replicas":   desiredReplicas,
			"available_replicas": availableReplicas,
			"healthy":            status.Healthy,
		})
	}

	response := Response{
		Success: true,
		Data: map[string]interface{}{
			"deployments": deploymentStatuses,
			"namespace":   namespace,
			"count":       len(deploymentStatuses),
		},
	}

	jsonResponse, _ := json.Marshal(response)
	ctx.SetBody(jsonResponse)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func handleWatchDeployments(ctx *fasthttp.RequestCtx) {
	// This endpoint would implement Server-Sent Events (SSE) for real-time updates
	// For now, we'll return a simple response indicating the feature
	response := Response{
		Success: true,
		Message: "Watch endpoint - implement SSE for real-time updates",
		Data: map[string]interface{}{
			"feature": "deployment_watching",
			"status":  "not_implemented",
		},
	}

	jsonResponse, _ := json.Marshal(response)
	ctx.SetBody(jsonResponse)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func handleGetEvents(ctx *fasthttp.RequestCtx, clientset *kubernetes.Clientset) {
	namespace := string(ctx.QueryArgs().Peek("namespace"))
	if namespace == "" {
		namespace = "default"
	}

	limitStr := string(ctx.QueryArgs().Peek("limit"))
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	namespaceLogger := log.WithNamespace(namespace)
	namespaceLogger.Info("HTTP request: Get events", map[string]interface{}{
		"namespace": namespace,
		"limit":     limit,
	})

	events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		Limit: int64(limit),
	})
	if err != nil {
		namespaceLogger.Error("Failed to get events", err, nil)
		sendErrorResponse(ctx, "Failed to get events", err, fasthttp.StatusInternalServerError)
		return
	}

	var eventList []Event
	for _, event := range events.Items {
		k8sEvent := Event{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Timestamp: event.LastTimestamp.Time,
			Object:    event.InvolvedObject.Name,
		}
		eventList = append(eventList, k8sEvent)

		// Log events based on their type
		eventLogger := namespaceLogger.WithDeployment(event.InvolvedObject.Name)
		fields := map[string]interface{}{
			"event_type":    event.Type,
			"event_reason":  event.Reason,
			"event_message": event.Message,
			"timestamp":     event.LastTimestamp,
		}

		switch event.Type {
		case "Warning":
			eventLogger.Warn("Kubernetes event", fields)
		case "Normal":
			eventLogger.Debug("Kubernetes event", fields)
		default:
			eventLogger.Info("Kubernetes event", fields)
		}
	}

	response := Response{
		Success: true,
		Data: map[string]interface{}{
			"events":    eventList,
			"namespace": namespace,
			"count":     len(eventList),
		},
	}

	jsonResponse, _ := json.Marshal(response)
	ctx.SetBody(jsonResponse)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func handleGetStatus(ctx *fasthttp.RequestCtx, clientset *kubernetes.Clientset) {
	namespace := string(ctx.QueryArgs().Peek("namespace"))
	if namespace == "" {
		namespace = "default"
	}

	namespaceLogger := log.WithNamespace(namespace)
	namespaceLogger.Info("HTTP request: Get cluster status", map[string]interface{}{
		"namespace": namespace,
	})

	// Get deployments
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		namespaceLogger.Error("Failed to get deployments", err, nil)
		sendErrorResponse(ctx, "Failed to get deployments", err, fasthttp.StatusInternalServerError)
		return
	}

	// Get pods
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		namespaceLogger.Error("Failed to get pods", err, nil)
		sendErrorResponse(ctx, "Failed to get pods", err, fasthttp.StatusInternalServerError)
		return
	}

	// Get services
	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		namespaceLogger.Error("Failed to get services", err, nil)
		sendErrorResponse(ctx, "Failed to get services", err, fasthttp.StatusInternalServerError)
		return
	}

	// Calculate pod status
	podStatus := make(map[corev1.PodPhase]int32)
	for _, pod := range pods.Items {
		podStatus[pod.Status.Phase]++
	}

	// Calculate deployment health
	healthyDeployments := 0
	unhealthyDeployments := 0
	for _, deployment := range deployments.Items {
		if deployment.Status.ReadyReplicas >= *deployment.Spec.Replicas {
			healthyDeployments++
		} else {
			unhealthyDeployments++
		}
	}

	status := map[string]interface{}{
		"namespace": map[string]interface{}{
			"name": namespace,
		},
		"deployments": map[string]interface{}{
			"total":     len(deployments.Items),
			"healthy":   healthyDeployments,
			"unhealthy": unhealthyDeployments,
		},
		"pods": map[string]interface{}{
			"total":  len(pods.Items),
			"status": podStatus,
		},
		"services": map[string]interface{}{
			"total": len(services.Items),
		},
		"timestamp": time.Now().UTC(),
	}

	response := Response{
		Success: true,
		Data:    status,
	}

	jsonResponse, _ := json.Marshal(response)
	ctx.SetBody(jsonResponse)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func handleNotFound(ctx *fasthttp.RequestCtx) {
	response := Response{
		Success: false,
		Error:   "Endpoint not found",
		Message: "The requested endpoint does not exist",
	}

	jsonResponse, _ := json.Marshal(response)
	ctx.SetBody(jsonResponse)
	ctx.SetStatusCode(fasthttp.StatusNotFound)
}

func sendErrorResponse(ctx *fasthttp.RequestCtx, message string, err error, statusCode int) {
	response := Response{
		Success: false,
		Error:   message,
		Message: err.Error(),
	}

	jsonResponse, _ := json.Marshal(response)
	ctx.SetBody(jsonResponse)
	ctx.SetStatusCode(statusCode)
}
