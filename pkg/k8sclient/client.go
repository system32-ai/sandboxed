package k8sclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
)

// Client wraps the Kubernetes clientset
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
	namespace string
}

// PodSpec represents the configuration for creating a pod
type PodSpec struct {
	Name      string
	Namespace string
	Image     string
	Command   []string
	Args      []string
	Labels    map[string]string
}

// NewClient creates a new Kubernetes client
func NewClient(namespace string) (*Client, error) {
	// Try to get the kubeconfig from the default location
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	var config *rest.Config
	var err error

	// Try to use out-of-cluster config first (kubeconfig file)
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		// If that fails, try in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %v", err)
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	if namespace == "" {
		namespace = "default"
	}

	return &Client{
		clientset: clientset,
		config:    config,
		namespace: namespace,
	}, nil
}

// CreatePod creates a new pod in the cluster
func (c *Client) CreatePod(spec PodSpec) (*corev1.Pod, error) {
	if spec.Namespace == "" {
		spec.Namespace = c.namespace
	}

	if spec.Labels == nil {
		spec.Labels = make(map[string]string)
	}
	
	// Add default labels
	spec.Labels["app"] = spec.Name
	spec.Labels["created-by"] = "sandboxed-cli"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: spec.Namespace,
			Labels:    spec.Labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:  spec.Name,
					Image: spec.Image,
				},
			},
		},
	}

	// Add command and args if specified
	if len(spec.Command) > 0 {
		pod.Spec.Containers[0].Command = spec.Command
	}
	if len(spec.Args) > 0 {
		pod.Spec.Containers[0].Args = spec.Args
	}

	// Create the pod
	createdPod, err := c.clientset.CoreV1().Pods(spec.Namespace).Create(
		context.TODO(),
		pod,
		metav1.CreateOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	return createdPod, nil
}

// DeletePod deletes a pod from the cluster
func (c *Client) DeletePod(name, namespace string) error {
	return c.DeletePodWithOptions(name, namespace, false)
}

// ForceDeletePod forcefully deletes a pod from the cluster
func (c *Client) ForceDeletePod(name, namespace string) error {
	return c.DeletePodWithOptions(name, namespace, true)
}

// DeletePodWithOptions deletes a pod with specific options
func (c *Client) DeletePodWithOptions(name, namespace string, force bool) error {
	if namespace == "" {
		namespace = c.namespace
	}

	deleteOptions := metav1.DeleteOptions{}
	
	if force {
		// Set grace period to 0 for immediate deletion
		gracePeriodSeconds := int64(0)
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
		
		// Set propagation policy to foreground for immediate deletion
		foregroundDeletion := metav1.DeletePropagationForeground
		deleteOptions.PropagationPolicy = &foregroundDeletion
	}

	err := c.clientset.CoreV1().Pods(namespace).Delete(
		context.TODO(),
		name,
		deleteOptions,
	)
	if err != nil {
		return fmt.Errorf("failed to delete pod %s in namespace %s: %v", name, namespace, err)
	}

	return nil
}

// GetPod retrieves a pod by name
func (c *Client) GetPod(name, namespace string) (*corev1.Pod, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	pod, err := c.clientset.CoreV1().Pods(namespace).Get(
		context.TODO(),
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s in namespace %s: %v", name, namespace, err)
	}

	return pod, nil
}

// ListPods lists all pods in the namespace
func (c *Client) ListPods(namespace string) (*corev1.PodList, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	pods, err := c.clientset.CoreV1().Pods(namespace).List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %v", namespace, err)
	}

	return pods, nil
}

// WaitForPodReady waits for a pod to be in Ready state
func (c *Client) WaitForPodReady(name, namespace string, timeout time.Duration) error {
	if namespace == "" {
		namespace = c.namespace
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pod %s to be ready", name)
		default:
			pod, err := c.GetPod(name, namespace)
			if err != nil {
				return err
			}

			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
					return nil
				}
			}

			time.Sleep(2 * time.Second)
		}
	}
}

// GetPodLogs retrieves logs from a pod
func (c *Client) GetPodLogs(name, namespace string) (string, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	podLogOpts := corev1.PodLogOptions{}
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(name, &podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to get logs for pod %s: %v", name, err)
	}
	defer podLogs.Close()

	buf := make([]byte, 2048)
	var logs string
	for {
		numBytes, err := podLogs.Read(buf)
		if numBytes == 0 {
			break
		}
		if err != nil {
			break
		}
		logs += string(buf[:numBytes])
	}

	return logs, nil
}

// ExecOptions represents options for executing commands in a pod
type ExecOptions struct {
	Command   []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	TTY       bool
	Container string
}

// ExecInPod executes a command in a running pod
func (c *Client) ExecInPod(podName, namespace string, options ExecOptions) error {
	if namespace == "" {
		namespace = c.namespace
	}

	if options.Stdout == nil {
		options.Stdout = os.Stdout
	}
	if options.Stderr == nil {
		options.Stderr = os.Stderr
	}

	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	// Set up the exec options
	execOptions := &corev1.PodExecOptions{
		Command: options.Command,
		Stdin:   options.Stdin != nil,
		Stdout:  true,
		Stderr:  true,
		TTY:     options.TTY,
	}

	if options.Container != "" {
		execOptions.Container = options.Container
	}

	req.VersionedParams(
		execOptions,
		scheme.ParameterCodec,
	)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %v", err)
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  options.Stdin,
		Stdout: options.Stdout,
		Stderr: options.Stderr,
		Tty:    options.TTY,
	})
	if err != nil {
		return fmt.Errorf("failed to execute command in pod: %v", err)
	}

	return nil
}

// ExecCommand executes a command in a pod and returns the output
func (c *Client) ExecCommand(podName, namespace string, command []string) (string, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	var stdout, stderr bytes.Buffer

	err := c.ExecInPod(podName, namespace, ExecOptions{
		Command: command,
		Stdout:  &stdout,
		Stderr:  &stderr,
		TTY:     true,
	})
	if err != nil {
		return "", fmt.Errorf("exec failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// CreateAndRunPod creates a pod, waits for it to be ready, and optionally executes commands
func (c *Client) CreateAndRunPod(spec PodSpec, commands [][]string, cleanup bool) ([]string, error) {
	// Create the pod
	pod, err := c.CreatePod(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	// If cleanup is requested, delete the pod when done
	if cleanup {
		defer func() {
			_ = c.DeletePod(pod.Name, pod.Namespace)
		}()
	}

	// Wait for pod to be ready
	err = c.WaitForPodReady(pod.Name, pod.Namespace, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("pod not ready: %v", err)
	}

	// Execute commands if provided
	var results []string
	for _, command := range commands {
		output, err := c.ExecCommand(pod.Name, pod.Namespace, command)
		if err != nil {
			return results, fmt.Errorf("command execution failed: %v", err)
		}
		results = append(results, output)
	}

	return results, nil
}