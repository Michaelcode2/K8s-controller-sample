package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/yourusername/k8s-controller-tutorial/pkg/logger"
)

var (
	namespace  string
	kubeconfig string
	inCluster  bool
	log        *logger.Logger
)

// controllerCmd represents the controller command
var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Monitor Kubernetes deployments and events with informers",
	Long: `A Kubernetes controller that monitors deployments and shows their current state
along with recent events using efficient informers.

Supports both in-cluster and kubeconfig authentication:
  • Use --in-cluster for running inside a Kubernetes cluster
  • Use --kubeconfig for external cluster access
  • Real-time monitoring with local caching and event deduplication`,
	Run: runController,
}

func init() {
	rootCmd.AddCommand(controllerCmd)
	controllerCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace to monitor")
	controllerCmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Path to kubeconfig")
	controllerCmd.Flags().BoolVarP(&inCluster, "in-cluster", "i", false, "Use in-cluster config")

	// Initialize logger
	log = logger.New()
}

func runController(cmd *cobra.Command, args []string) {
	namespaceLogger := log.WithNamespace(namespace)

	namespaceLogger.Info("Starting Kubernetes Controller", map[string]interface{}{
		"namespace":  namespace,
		"in_cluster": inCluster,
		"kubeconfig": kubeconfig,
		"mode":       "informer",
	})

	clientset, err := getKubernetesClient()
	if err != nil {
		log.Fatal("Failed to get Kubernetes client", err, map[string]interface{}{
			"namespace": namespace,
		})
	}

	// Show initial deployment status
	showDeploymentStatus(clientset, namespaceLogger)

	// Always use informer for real-time monitoring
	watchDeploymentsWithInformer(clientset, namespaceLogger)
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	if inCluster {
		log.Debug("Using in-cluster config", nil)
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load in-cluster config: %v", err)
		}
	} else {
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

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	log.Debug("Kubernetes client created successfully", map[string]interface{}{
		"auth_method": map[bool]string{true: "in_cluster", false: "kubeconfig"}[inCluster],
	})
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

func watchDeploymentsWithInformer(clientset *kubernetes.Clientset, namespaceLogger *logger.Logger) {
	namespaceLogger.Info("Starting deployment informer", map[string]interface{}{
		"namespace": namespace,
	})

	fmt.Println("Starting deployment informer... (Press Ctrl+C to stop)")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create shared informer factory
	factory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		30*time.Second, // resync period
		informers.WithNamespace(namespace),
	)

	// Get deployment informer
	deploymentInformer := factory.Apps().V1().Deployments().Informer()

	// Add event handlers
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			timestamp := time.Now().Format("15:04:05")

			deploymentLogger := namespaceLogger.WithDeployment(deployment.Name)
			deploymentLogger.Info("Deployment added", map[string]interface{}{
				"event_type":       "ADDED",
				"ready_replicas":   deployment.Status.ReadyReplicas,
				"desired_replicas": *deployment.Spec.Replicas,
			})

			fmt.Printf("[%s] ADDED: %s (%d/%d replicas)\n",
				timestamp,
				deployment.Name,
				deployment.Status.ReadyReplicas,
				*deployment.Spec.Replicas,
			)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDeployment := oldObj.(*appsv1.Deployment)
			newDeployment := newObj.(*appsv1.Deployment)
			timestamp := time.Now().Format("15:04:05")

			deploymentLogger := namespaceLogger.WithDeployment(newDeployment.Name)
			deploymentLogger.Info("Deployment updated", map[string]interface{}{
				"event_type":           "MODIFIED",
				"old_ready_replicas":   oldDeployment.Status.ReadyReplicas,
				"new_ready_replicas":   newDeployment.Status.ReadyReplicas,
				"old_desired_replicas": *oldDeployment.Spec.Replicas,
				"new_desired_replicas": *newDeployment.Spec.Replicas,
			})

			fmt.Printf("[%s] MODIFIED: %s (%d/%d -> %d/%d replicas)\n",
				timestamp,
				newDeployment.Name,
				oldDeployment.Status.ReadyReplicas,
				*oldDeployment.Spec.Replicas,
				newDeployment.Status.ReadyReplicas,
				*newDeployment.Spec.Replicas,
			)
		},
		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			timestamp := time.Now().Format("15:04:05")

			deploymentLogger := namespaceLogger.WithDeployment(deployment.Name)
			deploymentLogger.Info("Deployment deleted", map[string]interface{}{
				"event_type": "DELETED",
			})

			fmt.Printf("[%s] DELETED: %s\n", timestamp, deployment.Name)
		},
	})

	// Start the informer
	namespaceLogger.Info("Starting informer factory", nil)
	factory.Start(ctx.Done())

	// Wait for cache sync
	namespaceLogger.Info("Waiting for cache sync", nil)
	for t, ok := range factory.WaitForCacheSync(ctx.Done()) {
		if !ok {
			namespaceLogger.Error("Failed to sync informer", fmt.Errorf("sync failed for %v", t), nil)
			os.Exit(1)
		}
	}

	namespaceLogger.Info("Informer cache synced successfully", nil)
	fmt.Println("Informer cache synced. Watching for deployment events...")

	// Block until context is cancelled
	<-ctx.Done()
	namespaceLogger.Info("Informer stopped", nil)
}
