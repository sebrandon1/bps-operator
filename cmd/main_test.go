package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/sebrandon1/bps-operator/internal/probe"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestParseFlags_Defaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	opts, err := parseFlags(fs, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.metricsAddr != ":8080" {
		t.Errorf("expected :8080, got %s", opts.metricsAddr)
	}
	if opts.probeAddr != ":8081" {
		t.Errorf("expected :8081, got %s", opts.probeAddr)
	}
	if opts.probeImage != probe.ProbeImage {
		t.Errorf("expected %s, got %s", probe.ProbeImage, opts.probeImage)
	}
	if opts.probeExecTimeout != probe.DefaultExecTimeout {
		t.Errorf("expected %v, got %v", probe.DefaultExecTimeout, opts.probeExecTimeout)
	}
	if opts.certAPIURL != "" {
		t.Errorf("expected empty, got %s", opts.certAPIURL)
	}
	if opts.operatorNamespace != "" {
		t.Errorf("expected empty, got %s", opts.operatorNamespace)
	}
	if opts.nodeName != "" {
		t.Errorf("expected empty, got %s", opts.nodeName)
	}
}

func TestParseFlags_CustomValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	opts, err := parseFlags(fs, []string{
		"--metrics-bind-address=:9090",
		"--health-probe-bind-address=:9091",
		"--operator-namespace=custom-ns",
		"--probe-image=my-image:v1",
		"--probe-exec-timeout=60s",
		"--certification-api-url=https://example.com/api",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.metricsAddr != ":9090" {
		t.Errorf("expected :9090, got %s", opts.metricsAddr)
	}
	if opts.probeAddr != ":9091" {
		t.Errorf("expected :9091, got %s", opts.probeAddr)
	}
	if opts.operatorNamespace != "custom-ns" {
		t.Errorf("expected custom-ns, got %s", opts.operatorNamespace)
	}
	if opts.probeImage != "my-image:v1" {
		t.Errorf("expected my-image:v1, got %s", opts.probeImage)
	}
	if opts.probeExecTimeout != 60*time.Second {
		t.Errorf("expected 60s, got %v", opts.probeExecTimeout)
	}
	if opts.certAPIURL != "https://example.com/api" {
		t.Errorf("expected https://example.com/api, got %s", opts.certAPIURL)
	}
}

func TestParseFlags_EnvNotReadDuringParsing(t *testing.T) {
	// parseFlags should NOT read env vars — that's main()'s job
	t.Setenv("OPERATOR_NAMESPACE", "env-ns")
	t.Setenv("NODE_NAME", "env-node")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	opts, err := parseFlags(fs, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.operatorNamespace != "" {
		t.Errorf("expected empty (env should not be read by parseFlags), got %s", opts.operatorNamespace)
	}
	if opts.nodeName != "" {
		t.Errorf("expected empty (env should not be read by parseFlags), got %s", opts.nodeName)
	}
}

func TestParseFlags_FlagSetsValue(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	opts, err := parseFlags(fs, []string{
		"--operator-namespace=flag-ns",
		"--node-name=flag-node",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.operatorNamespace != "flag-ns" {
		t.Errorf("expected flag-ns, got %s", opts.operatorNamespace)
	}
	if opts.nodeName != "flag-node" {
		t.Errorf("expected flag-node, got %s", opts.nodeName)
	}
}

func TestParseFlags_InvalidFlag(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_, err := parseFlags(fs, []string{"--nonexistent-flag=value"})
	if err == nil {
		t.Error("expected error for invalid flag")
	}
}

func TestSchemeRegistration(t *testing.T) {
	// Verify key types are registered in the scheme
	gvks := []schema.GroupVersionKind{
		{Group: "", Version: "v1", Kind: "Pod"},
		{Group: "apps", Version: "v1", Kind: "Deployment"},
		{Group: "bps.openshift.io", Version: "v1alpha1", Kind: "BestPracticeScanner"},
		{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"},
		{Group: "config.openshift.io", Version: "v1", Kind: "ClusterVersion"},
		{Group: "operators.coreos.com", Version: "v1alpha1", Kind: "ClusterServiceVersion"},
	}

	for _, gvk := range gvks {
		t.Run(gvk.Kind, func(t *testing.T) {
			if !scheme.Recognizes(gvk) {
				t.Errorf("scheme does not recognize %s", gvk)
			}
		})
	}
}

func TestMain(m *testing.M) {
	// Ensure tests don't pick up KUBECONFIG
	_ = os.Unsetenv("KUBECONFIG")
	os.Exit(m.Run())
}
