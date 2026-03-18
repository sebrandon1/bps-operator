package main

import (
	"flag"
	"os"
	"time"

	netattdefv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmpackagev1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	apiserverv1 "github.com/openshift/api/apiserver/v1"
	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	bpsv1alpha1 "github.com/sebrandon1/bps-operator/api/v1alpha1"
	"github.com/sebrandon1/bps-operator/internal/certification"
	"github.com/sebrandon1/bps-operator/internal/controller"
	"github.com/sebrandon1/bps-operator/internal/probe"

	// Register checks via init()
	_ "github.com/redhat-best-practices-for-k8s/checks/accesscontrol"
	_ "github.com/redhat-best-practices-for-k8s/checks/certification"
	_ "github.com/redhat-best-practices-for-k8s/checks/lifecycle"
	_ "github.com/redhat-best-practices-for-k8s/checks/manageability"
	_ "github.com/redhat-best-practices-for-k8s/checks/networking"
	_ "github.com/redhat-best-practices-for-k8s/checks/observability"
	_ "github.com/redhat-best-practices-for-k8s/checks/operator"
	_ "github.com/redhat-best-practices-for-k8s/checks/performance"
	_ "github.com/redhat-best-practices-for-k8s/checks/platform"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(bpsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(apiextv1.AddToScheme(scheme))
	utilruntime.Must(configv1.Install(scheme))
	utilruntime.Must(apiserverv1.Install(scheme))
	utilruntime.Must(olmv1alpha1.AddToScheme(scheme))
	utilruntime.Must(olmpackagev1.AddToScheme(scheme))
	utilruntime.Must(netattdefv1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var probeAddr string
	var operatorNamespace string
	var probeImage string
	var probeExecTimeout time.Duration
	var certAPIURL string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&operatorNamespace, "operator-namespace", os.Getenv("OPERATOR_NAMESPACE"), "Namespace where the operator runs (for probe DaemonSet).")
	flag.StringVar(&probeImage, "probe-image", probe.ProbeImage, "Probe DaemonSet container image.")
	flag.DurationVar(&probeExecTimeout, "probe-exec-timeout", 30*time.Second, "Timeout for probe command execution.")
	flag.StringVar(&certAPIURL, "certification-api-url", "", "Red Hat Pyxis API base URL for certification checks (default: Red Hat Catalog API).")

	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	cfg := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
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
	if cfg != nil {
		probeExecutor, err = probe.NewExecutor(cfg, probeExecTimeout)
		if err != nil {
			setupLog.Error(err, "unable to create probe executor, probe-based checks will be skipped")
		}
	}

	// Create discovery client for K8s version
	k8sClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		setupLog.Error(err, "unable to create kubernetes clientset")
		os.Exit(1)
	}

	// Create scale client for CRD scaling checks
	discoveryClient := k8sClientset.Discovery()
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		setupLog.Error(err, "unable to get API group resources for scale client")
		os.Exit(1)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	resolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)
	scaleClient := scale.New(discoveryClient.RESTClient(), mapper, dynamic.LegacyAPIPathResolverFunc, resolver)

	// Create certification validator
	certValidator := certification.NewPyxisValidator(certAPIURL)

	if err := (&controller.ScannerReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		ProbeExecutor:     probeExecutor,
		OperatorNamespace: operatorNamespace,
		ProbeImage:        probeImage,
		ScannerNodeName:   os.Getenv("NODE_NAME"),
		CertValidator:     certValidator,
		DiscoveryClient:   discoveryClient,
		K8sClientset:      k8sClientset,
		ScaleClient:       scaleClient,
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
