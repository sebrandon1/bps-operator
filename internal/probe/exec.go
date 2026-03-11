package probe

import (
	"bytes"
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const execTimeout = 30 * time.Second

// Executor implements checks.ProbeExecutor using the Kubernetes pods/exec API.
type Executor struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewExecutor creates a new Executor from the given REST config.
func NewExecutor(config *rest.Config) (*Executor, error) {
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating clientset: %w", err)
	}
	return &Executor{clientset: cs, config: config}, nil
}

// ExecCommand runs a command on the given probe pod and returns stdout/stderr.
func (e *Executor) ExecCommand(ctx context.Context, pod *corev1.Pod, command string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	req := e.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: ProbeName,
			Command:   []string{"/bin/sh", "-c", command},
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(e.config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("creating executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	if err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}); err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("executing command: %w", err)
	}

	return stdout.String(), stderr.String(), nil
}
