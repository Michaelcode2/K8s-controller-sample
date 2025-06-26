package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/yourusername/k8s-controller-tutorial/pkg/logger"
)

var (
	namespace  string
	watch      bool
	kubeconfig string
	log        *logger.Logger
)

// controllerCmd represents the controller command
var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Monitor Kubernetes deployments and events",
	Long: `A Kubernetes controller that monitors deployments and shows their current state
along with recent events.

You can specify a custom kubeconfig file using the --kubeconfig flag, otherwise
it will use the KUBECONFIG environment variable or default to ~/.kube/config.`,
	Run: runController,
}

func init() {
	rootCmd.AddCommand(controllerCmd)
	controllerCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace to monitor")
	controllerCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes continuously")
	controllerCmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Path to kubeconfig")

	// Initialize logger
	log = logger.New()
}

func runController(cmd *cobra.Command, args []string) {
	namespaceLogger := log.WithNamespace(namespace)

	namespaceLogger.Info("Starting Kubernetes Controller", map[string]interface{}{
		"namespace":  namespace,
		"watch_mode": watch,
	})

	clientset, err := getKubernetesClient()
	if err != nil {
		log.Fatal("Failed to get Kubernetes client", err, map[string]interface{}{
			"namespace": namespace,
		})
	}

	if watch {
		watchDeployments(clientset, namespaceLogger)
	} else {
		showDeploymentStatus(clientset, namespaceLogger)
	}
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
	kubeconfigPath := kubeconfig
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("KUBECONFIG")
	}
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}

	log.Debug("Loading kubeconfig", map[string]interface{}{
		"kubeconfig_path": kubeconfigPath,
	})

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	log.Debug("Kubernetes client created successfully", nil)
	return clientset, nil
}

func showDeploymentStatus(clientset *kubernetes.Clientset, namespaceLogger *logger.Logger) {
	namespaceLogger.Info("Fetching deployment status", nil)

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		namespaceLogger.Error("Failed to get deployments", err, nil)
		return
	}

	namespaceLogger.Info("Deployment status retrieved", map[string]interface{}{
		"deployment_count": len(deployments.Items),
	})

	fmt.Println("DEPLOYMENT STATUS")
	fmt.Println("=================")

	for _, deployment := range deployments.Items {
		deploymentLogger := namespaceLogger.WithDeployment(deployment.Name)

		readyReplicas := deployment.Status.ReadyReplicas
		desiredReplicas := *deployment.Spec.Replicas
		availableReplicas := deployment.Status.AvailableReplicas

		deploymentLogger.Info("Deployment status", map[string]interface{}{
			"ready_replicas":     readyReplicas,
			"desired_replicas":   desiredReplicas,
			"available_replicas": availableReplicas,
			"updated_replicas":   deployment.Status.UpdatedReplicas,
		})

		fmt.Printf("\nDeployment: %s\n", deployment.Name)
		fmt.Printf("  Replicas: %d/%d (Available: %d, Ready: %d)\n",
			readyReplicas,
			desiredReplicas,
			availableReplicas,
			readyReplicas)

		// Log warnings for unhealthy deployments
		if readyReplicas < desiredReplicas {
			deploymentLogger.Warn("Deployment has fewer ready replicas than desired", map[string]interface{}{
				"ready_replicas":   readyReplicas,
				"desired_replicas": desiredReplicas,
			})
		}
	}

	showRecentEvents(clientset, namespaceLogger)
}

func showRecentEvents(clientset *kubernetes.Clientset, namespaceLogger *logger.Logger) {
	namespaceLogger.Info("Fetching recent events", nil)

	events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		Limit: 10,
	})
	if err != nil {
		namespaceLogger.Error("Failed to get events", err, nil)
		return
	}

	namespaceLogger.Info("Events retrieved", map[string]interface{}{
		"event_count": len(events.Items),
	})

	fmt.Println("\nRECENT EVENTS")
	fmt.Println("=============")

	for _, event := range events.Items {
		timestamp := event.LastTimestamp.Format("15:04:05")
		fmt.Printf("[%s] %s: %s\n", timestamp, event.Reason, event.Message)

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
}

func watchDeployments(clientset *kubernetes.Clientset, namespaceLogger *logger.Logger) {
	namespaceLogger.Info("Starting deployment watcher", map[string]interface{}{
		"watch_mode": true,
	})

	fmt.Println("Watching deployments for changes... (Press Ctrl+C to stop)")

	watcher, err := clientset.AppsV1().Deployments(namespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		namespaceLogger.Error("Failed to create deployment watcher", err, nil)
		return
	}
	defer watcher.Stop()

	namespaceLogger.Info("Deployment watcher started successfully", nil)

	for event := range watcher.ResultChan() {
		deployment := event.Object.(*appsv1.Deployment)
		timestamp := time.Now().Format("15:04:05")

		deploymentLogger := namespaceLogger.WithDeployment(deployment.Name)

		deploymentLogger.Info("Deployment event detected", map[string]interface{}{
			"event_type":       event.Type,
			"deployment_name":  deployment.Name,
			"ready_replicas":   deployment.Status.ReadyReplicas,
			"desired_replicas": *deployment.Spec.Replicas,
		})

		fmt.Printf("[%s] %s: %s\n", timestamp, event.Type, deployment.Name)
	}
}
