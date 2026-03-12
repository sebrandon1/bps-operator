package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	bpsv1alpha1 "github.com/sebrandon1/bps-operator/api/v1alpha1"
	"github.com/sebrandon1/bps-operator/internal/controller"
	"github.com/sebrandon1/bps-operator/internal/probe"

	// Register checks via init()
	_ "github.com/sebrandon1/bps-operator/internal/checks/accesscontrol"
	_ "github.com/sebrandon1/bps-operator/internal/checks/lifecycle"
	_ "github.com/sebrandon1/bps-operator/internal/checks/manageability"
	_ "github.com/sebrandon1/bps-operator/internal/checks/networking"
	_ "github.com/sebrandon1/bps-operator/internal/checks/observability"
	_ "github.com/sebrandon1/bps-operator/internal/checks/performance"
	_ "github.com/sebrandon1/bps-operator/internal/checks/platform"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(bpsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(apiextv1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var probeAddr string
	var operatorNamespace string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&operatorNamespace, "operator-namespace", os.Getenv("OPERATOR_NAMESPACE"), "Namespace where the operator runs (for probe DaemonSet).")

	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		os.Exit(1)
	}

	// Create probe executor
	var probeExecutor *probe.Executor
	if cfg := mgr.GetConfig(); cfg != nil {
		probeExecutor, err = probe.NewExecutor(cfg)
		if err != nil {
			setupLog.Error(err, "unable to create probe executor, probe-based checks will be skipped")
		}
	}

	if err := (&controller.ScannerReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		ProbeExecutor:     probeExecutor,
		OperatorNamespace: operatorNamespace,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Scanner")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
