package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/yourusername/k8s-controller-tutorial/pkg/logger"
)

var (
	managerNamespace     string
	managerKubeconfig    string
	managerInCluster     bool
	enableLeaderElection bool
	leaderElectionID     string
	metricsAddr          string
	healthAddr           string
	webhookPort          int
	managerLog           *logger.Logger
)

// managerCmd represents the manager command
var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Start controller-runtime manager with leader election",
	Long: `Start a controller-runtime manager that provides:
  • Leader election for high availability
  • Metrics and health endpoints
  • Structured controller lifecycle management
  • Webhook support
  • Graceful shutdown

This is the production-ready version using controller-runtime framework.`,
	Run: runManager,
}

func init() {
	rootCmd.AddCommand(managerCmd)

	// Manager configuration flags
	managerCmd.Flags().StringVarP(&managerNamespace, "namespace", "n", "", "Namespace to watch (empty for all namespaces)")
	managerCmd.Flags().StringVarP(&managerKubeconfig, "kubeconfig", "k", "", "Path to kubeconfig")
	managerCmd.Flags().BoolVarP(&managerInCluster, "in-cluster", "i", false, "Use in-cluster config")

	// Leader election flags
	managerCmd.Flags().BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager")
	managerCmd.Flags().StringVar(&leaderElectionID, "leader-election-id", "k8s-controller-sample", "Leader election ID")

	// Server configuration flags
	managerCmd.Flags().StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to")
	managerCmd.Flags().StringVar(&healthAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to")
	managerCmd.Flags().IntVar(&webhookPort, "webhook-port", 9443, "The port that the webhook server serves at")

	// Initialize logger
	managerLog = logger.New()
}

// DeploymentReconciler reconciles Deployment objects
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    *logger.Logger
}

// Reconcile handles Deployment events
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithNamespace(req.Namespace).WithDeployment(req.Name)
	reqLogger.Info("Reconciling Deployment", map[string]interface{}{
		"request": req.String(),
	})

	// Fetch the Deployment instance
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			reqLogger.Info("Deployment not found, probably deleted", map[string]interface{}{
				"deployment": req.Name,
			})
			return ctrl.Result{}, nil
		}
		reqLogger.Error("Failed to get Deployment", err, nil)
		return ctrl.Result{}, err
	}

	// Log deployment status
	reqLogger.Info("Deployment status", map[string]interface{}{
		"ready_replicas":      deployment.Status.ReadyReplicas,
		"desired_replicas":    *deployment.Spec.Replicas,
		"available_replicas":  deployment.Status.AvailableReplicas,
		"updated_replicas":    deployment.Status.UpdatedReplicas,
		"generation":          deployment.Generation,
		"observed_generation": deployment.Status.ObservedGeneration,
	})

	// Check if deployment is healthy
	isHealthy := deployment.Status.ReadyReplicas >= *deployment.Spec.Replicas
	if !isHealthy {
		reqLogger.Warn("Deployment is not healthy", map[string]interface{}{
			"ready_replicas":   deployment.Status.ReadyReplicas,
			"desired_replicas": *deployment.Spec.Replicas,
		})
	} else {
		reqLogger.Info("Deployment is healthy", map[string]interface{}{
			"replicas": deployment.Status.ReadyReplicas,
		})
	}

	// Requeue after 30 seconds for continuous monitoring
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}

func runManager(cmd *cobra.Command, args []string) {
	managerLog.Info("Starting controller-runtime manager", map[string]interface{}{
		"namespace":          managerNamespace,
		"in_cluster":         managerInCluster,
		"kubeconfig":         managerKubeconfig,
		"leader_election":    enableLeaderElection,
		"leader_election_id": leaderElectionID,
		"metrics_address":    metricsAddr,
		"health_address":     healthAddr,
		"webhook_port":       webhookPort,
	})

	// Setup scheme
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		managerLog.Fatal("Failed to add client-go scheme", err, nil)
	}

	// Setup manager configuration
	config := ctrl.GetConfigOrDie()
	if managerKubeconfig != "" {
		var err error
		config, err = ctrl.GetConfig()
		if err != nil {
			managerLog.Fatal("Failed to get config", err, nil)
		}
	}

	// Create manager options
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:                  scheme,
		Metrics:                 ctrl.Options{}.Metrics,
		WebhookServer:           ctrl.Options{}.WebhookServer,
		HealthProbeBindAddress:  healthAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        leaderElectionID,
		LeaderElectionNamespace: managerNamespace,
	})
	if err != nil {
		managerLog.Fatal("Failed to create manager", err, nil)
	}

	// Setup reconciler
	deploymentReconciler := &DeploymentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    managerLog,
	}

	if err = deploymentReconciler.SetupWithManager(mgr); err != nil {
		managerLog.Fatal("Failed to setup controller", err, map[string]interface{}{
			"controller": "Deployment",
		})
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		managerLog.Fatal("Failed to add health check", err, nil)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		managerLog.Fatal("Failed to add ready check", err, nil)
	}

	// Setup signal handler for graceful shutdown
	ctx := ctrl.SetupSignalHandler()

	managerLog.Info("Manager configured successfully", map[string]interface{}{
		"leader_election": enableLeaderElection,
		"namespace":       managerNamespace,
	})

	// Start the manager
	managerLog.Info("Starting manager", map[string]interface{}{
		"leader_election": enableLeaderElection,
	})
	if err := mgr.Start(ctx); err != nil {
		managerLog.Fatal("Manager failed to start", err, nil)
	}
}
