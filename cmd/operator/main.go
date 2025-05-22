package main

import (
	"context"
	"fmt"

	goruntime "runtime"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	api "github.com/ionos-cloud/octopinger/api/v1alpha1"
	"github.com/ionos-cloud/octopinger/pkg/controller"
	//+kubebuilder:scaffold:imports
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var build = fmt.Sprintf("%s (%s) (%s)", version, commit, date)

type flags struct {
	enableLeaderElection bool
	metricsAddr          string
	probeAddr            string
}

var f = &flags{}

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var rootCmd = &cobra.Command{
	Use:     "operator",
	Version: build,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd.Context())
	},
}

func printVersion() {
	setupLog.Info(fmt.Sprintf("Go Version: %s", goruntime.Version()))
	setupLog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", goruntime.GOOS, goruntime.GOARCH))
	setupLog.Info(fmt.Sprintf("Build: %s", build))
}

func init() {
	rootCmd.Flags().BoolVar(&f.enableLeaderElection, "leader-elect", f.enableLeaderElection, "only one controller")
	rootCmd.Flags().StringVar(&f.metricsAddr, "metrics-bind-address", ":8080", "metrics endpoint")
	rootCmd.Flags().StringVar(&f.probeAddr, "health-probe-bind-address", ":8081", "health probe")

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(api.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		setupLog.Error(err, "unable to run operator")
	}
}

func run(ctx context.Context) error {
	opts := zap.Options{
		Development: true,
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	printVersion()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: f.metricsAddr,
		},
		WebhookServer:              webhook.NewServer(webhook.Options{}),
		HealthProbeBindAddress:     f.probeAddr,
		LeaderElection:             f.enableLeaderElection,
		LeaderElectionID:           "j8yhqdnj.octopinger.io",
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	err = setupControllers(f, mgr)
	if err != nil {
		return err
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return err
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return err
	}

	setupLog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return err
	}

	return nil
}

func setupControllers(f *flags, mgr ctrl.Manager) error {
	err := controller.NewDaemonReconciler(mgr)
	if err != nil {
		return err
	}

	err = controller.NewConfigReconciler(mgr)
	if err != nil {
		return err
	}

	return nil
}
