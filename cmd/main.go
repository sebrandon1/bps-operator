package main

import (
	"flag"
	"fmt"
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
	"k8s.io/client-go/rest"
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

	// Register metrics
	_ "github.com/sebrandon1/bps-operator/internal/metrics"
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

// options holds parsed command-line flags.
type options struct {
	metricsAddr       string
	probeAddr         string
	operatorNamespace string
	probeImage        string
	probeExecTimeout  time.Duration
	certAPIURL        string
	nodeName          string
	catalogURLBase    string
}

// parseFlags parses command-line flags from the given FlagSet and arguments.
func parseFlags(fs *flag.FlagSet, args []string) (options, error) {
	var opts options

	fs.StringVar(&opts.metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	fs.StringVar(&opts.probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.StringVar(&opts.operatorNamespace, "operator-namespace", "", "Namespace where the operator runs (for probe DaemonSet).")
	fs.StringVar(&opts.probeImage, "probe-image", probe.ProbeImage, "Probe DaemonSet container image.")
	fs.DurationVar(&opts.probeExecTimeout, "probe-exec-timeout", probe.DefaultExecTimeout, "Timeout for probe command execution.")
	fs.StringVar(&opts.certAPIURL, "certification-api-url", "", "Red Hat Pyxis API base URL for certification checks (default: Red Hat Catalog API).")
	fs.StringVar(&opts.nodeName, "node-name", "", "Node name where the scanner pod runs.")
	fs.StringVar(&opts.catalogURLBase, "catalog-url-base", "https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md", "Base URL for check catalog documentation.")

	if err := fs.Parse(args); err != nil {
		return options{}, fmt.Errorf("parsing flags: %w", err)
	}
	return opts, nil
}

// run executes the operator with the given options and rest config.
func run(opts options, cfg *rest.Config) error {
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: opts.metricsAddr,
		},
		HealthProbeBindAddress: opts.probeAddr,
	})
	if err != nil {
		return fmt.Errorf("creating manager: %w", err)
	}

	// Create probe executor
	probeExecutor, err := probe.NewExecutor(cfg, opts.probeExecTimeout)
	if err != nil {
		setupLog.Error(err, "unable to create probe executor, probe-based checks will be skipped")
	}

	// Create discovery client for K8s version
	k8sClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("creating kubernetes clientset: %w", err)
	}

	// Create scale client for CRD scaling checks
	discoveryClient := k8sClientset.Discovery()
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return fmt.Errorf("getting API group resources for scale client: %w", err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	resolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)
	scaleClient := scale.New(discoveryClient.RESTClient(), mapper, dynamic.LegacyAPIPathResolverFunc, resolver)

	// Create certification validator
	certValidator := certification.NewPyxisValidator(opts.certAPIURL)

	if err := (&controller.ScannerReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		Recorder:          mgr.GetEventRecorderFor("scanner"), //nolint:staticcheck // TODO: migrate to GetEventRecorder (events.EventRecorder)
		ProbeExecutor:     probeExecutor,
		OperatorNamespace: opts.operatorNamespace,
		ProbeImage:        opts.probeImage,
		ScannerNodeName:   opts.nodeName,
		CertValidator:     certValidator,
		DiscoveryClient:   discoveryClient,
		K8sClientset:      k8sClientset,
		ScaleClient:       scaleClient,
		CatalogURLBase:    opts.catalogURLBase,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("creating controller: %w", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("setting up health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("setting up ready check: %w", err)
	}

	setupLog.Info("starting manager")
	return mgr.Start(ctrl.SetupSignalHandler())
}

func main() {
	zapOpts := zap.Options{Development: true}
	zapOpts.BindFlags(flag.CommandLine)

	opts, err := parseFlags(flag.CommandLine, os.Args[1:])
	if err != nil {
		setupLog.Error(err, "unable to parse flags")
		os.Exit(1)
	}

	// Fall back to environment variables when flags are not set
	if opts.operatorNamespace == "" {
		opts.operatorNamespace = os.Getenv("OPERATOR_NAMESPACE")
	}
	if opts.nodeName == "" {
		opts.nodeName = os.Getenv("NODE_NAME")
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zapOpts)))

	cfg := ctrl.GetConfigOrDie()

	if err := run(opts, cfg); err != nil {
		setupLog.Error(err, "operator failed")
		os.Exit(1)
	}
}
