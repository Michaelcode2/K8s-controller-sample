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
)

var (
	namespace string
	watch     bool
)

// controllerCmd represents the controller command
var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Monitor Kubernetes deployments and events",
	Long: `A Kubernetes controller that monitors deployments and shows their current state
along with recent events.`,
	Run: runController,
}

func init() {
	rootCmd.AddCommand(controllerCmd)
	controllerCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace to monitor")
	controllerCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes continuously")
}

func runController(cmd *cobra.Command, args []string) {
	fmt.Printf("Starting Kubernetes Controller for namespace: %s\n", namespace)

	clientset, err := getKubernetesClient()
	if err != nil {
		fmt.Printf("Error getting Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	if watch {
		watchDeployments(clientset)
	} else {
		showDeploymentStatus(clientset)
	}
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return clientset, nil
}

func showDeploymentStatus(clientset *kubernetes.Clientset) {
	fmt.Println("DEPLOYMENT STATUS")
	fmt.Println("=================")

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting deployments: %v\n", err)
		return
	}

	for _, deployment := range deployments.Items {
		fmt.Printf("\nDeployment: %s\n", deployment.Name)
		fmt.Printf("  Replicas: %d/%d (Available: %d, Ready: %d)\n",
			deployment.Status.ReadyReplicas,
			*deployment.Spec.Replicas,
			deployment.Status.AvailableReplicas,
			deployment.Status.ReadyReplicas)
	}

	showRecentEvents(clientset)
}

func showRecentEvents(clientset *kubernetes.Clientset) {
	fmt.Println("\nRECENT EVENTS")
	fmt.Println("=============")

	events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		Limit: 10,
	})
	if err != nil {
		fmt.Printf("Error getting events: %v\n", err)
		return
	}

	for _, event := range events.Items {
		timestamp := event.LastTimestamp.Format("15:04:05")
		fmt.Printf("[%s] %s: %s\n", timestamp, event.Reason, event.Message)
	}
}

func watchDeployments(clientset *kubernetes.Clientset) {
	fmt.Println("Watching deployments for changes... (Press Ctrl+C to stop)")

	watcher, err := clientset.AppsV1().Deployments(namespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error watching deployments: %v\n", err)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case event := <-watcher.ResultChan():
			deployment := event.Object.(*appsv1.Deployment)
			timestamp := time.Now().Format("15:04:05")
			fmt.Printf("[%s] %s: %s\n", timestamp, event.Type, deployment.Name)
		}
	}
}
