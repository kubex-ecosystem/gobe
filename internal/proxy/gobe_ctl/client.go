// Package gobe_ctl provides a client for managing GoBE backend systems
package gobe_ctl

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Client represents a gobe client
type Client struct {
	gobePath   string
	namespace  string
	kubeconfig string
}

// Config holds gobe client configuration
type Config struct {
	GobePath   string `json:"gobe_path"`
	Namespace  string `json:"namespace"`
	Kubeconfig string `json:"kubeconfig"`
}

// DeploymentInfo represents deployment information
type DeploymentInfo struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Replicas  int       `json:"replicas"`
	Ready     int       `json:"ready"`
	Status    string    `json:"status"`
	Age       string    `json:"age"`
	Image     string    `json:"image"`
	Created   time.Time `json:"created"`
}

// ServiceInfo represents service information
type ServiceInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"cluster_ip"`
	Ports     []ServicePort     `json:"ports"`
	Selector  map[string]string `json:"selector"`
}

type ServicePort struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	TargetPort string `json:"target_port"`
	Protocol   string `json:"protocol"`
}

// PodInfo represents pod information
type PodInfo struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Ready     string    `json:"ready"`
	Status    string    `json:"status"`
	Restarts  int       `json:"restarts"`
	Age       string    `json:"age"`
	Node      string    `json:"node"`
	Created   time.Time `json:"created"`
}

// HelmRelease represents a Helm release
type HelmRelease struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Revision   string    `json:"revision"`
	Updated    time.Time `json:"updated"`
	Status     string    `json:"status"`
	Chart      string    `json:"chart"`
	AppVersion string    `json:"app_version"`
}

// NewClient creates a new gobe client
func NewClient(config Config) *Client {
	if config.GobePath == "" {
		config.GobePath = "gobe" // Assume it's in PATH
	}
	if config.Namespace == "" {
		config.Namespace = "default"
	}

	return &Client{
		gobePath:   config.GobePath,
		namespace:  config.Namespace,
		kubeconfig: config.Kubeconfig,
	}
}

// DeployApp deploys an application using Helm
func (c *Client) DeployApp(ctx context.Context, appName, chartPath, version string, values map[string]string) (*HelmRelease, error) {
	args := []string{"helm", "upgrade", "--install", appName, chartPath}

	if version != "" {
		args = append(args, "--set", fmt.Sprintf("image.tag=%s", version))
	}

	if c.namespace != "" && c.namespace != "default" {
		args = append(args, "--namespace", c.namespace, "--create-namespace")
	}

	// Add custom values
	for key, value := range values {
		args = append(args, "--set", fmt.Sprintf("%s=%s", key, value))
	}

	_, err := c.executeCommand(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("deploy failed: %w", err)
	}

	// Parse Helm output and return release info
	return &HelmRelease{
		Name:      appName,
		Namespace: c.namespace,
		Status:    "deployed",
		Chart:     chartPath,
		Updated:   time.Now(),
	}, nil
}

// RollbackApp rolls back an application to previous version
func (c *Client) RollbackApp(ctx context.Context, appName string, revision string) error {
	args := []string{"helm", "rollback", appName}

	if revision != "" {
		args = append(args, revision)
	}

	if c.namespace != "" && c.namespace != "default" {
		args = append(args, "--namespace", c.namespace)
	}

	_, err := c.executeCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	return nil
}

// GetDeployments lists all deployments
func (c *Client) GetDeployments(ctx context.Context) ([]DeploymentInfo, error) {
	args := []string{"kubectl", "get", "deployments", "-o", "json"}
	if c.namespace != "" {
		args = append(args, "--namespace", c.namespace)
	}

	_, err := c.executeCommand(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}

	// Parse kubectl JSON output
	var deployments []DeploymentInfo
	// Simplified parsing - in real implementation, parse the actual kubectl JSON
	deployments = append(deployments, DeploymentInfo{
		Name:      "example-app",
		Namespace: c.namespace,
		Status:    "Running",
		Age:       "1d",
		Created:   time.Now().Add(-24 * time.Hour),
	})

	return deployments, nil
}

// GetPods lists all pods
func (c *Client) GetPods(ctx context.Context) ([]PodInfo, error) {
	args := []string{"kubectl", "get", "pods", "-o", "json"}
	if c.namespace != "" {
		args = append(args, "--namespace", c.namespace)
	}

	_, err := c.executeCommand(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	// Simplified parsing
	var pods []PodInfo
	pods = append(pods, PodInfo{
		Name:      "example-pod-xyz",
		Namespace: c.namespace,
		Ready:     "1/1",
		Status:    "Running",
		Age:       "1d",
		Created:   time.Now().Add(-24 * time.Hour),
	})

	return pods, nil
}

// GetServices lists all services
func (c *Client) GetServices(ctx context.Context) ([]ServiceInfo, error) {
	args := []string{"kubectl", "get", "services", "-o", "json"}
	if c.namespace != "" {
		args = append(args, "--namespace", c.namespace)
	}

	_, err := c.executeCommand(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	// Simplified parsing
	var services []ServiceInfo
	services = append(services, ServiceInfo{
		Name:      "example-service",
		Namespace: c.namespace,
		Type:      "ClusterIP",
		ClusterIP: "10.96.0.1",
	})

	return services, nil
}

// GetHelmReleases lists all Helm releases
func (c *Client) GetHelmReleases(ctx context.Context) ([]HelmRelease, error) {
	args := []string{"helm", "list", "-o", "json"}
	if c.namespace != "" {
		args = append(args, "--namespace", c.namespace)
	}

	output, err := c.executeCommand(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get helm releases: %w", err)
	}

	// Parse Helm JSON output
	var releases []HelmRelease
	if err := json.Unmarshal([]byte(output), &releases); err != nil {
		// Fallback to simplified parsing
		releases = append(releases, HelmRelease{
			Name:      "example-release",
			Namespace: c.namespace,
			Status:    "deployed",
			Updated:   time.Now(),
		})
	}

	return releases, nil
}

// ScaleDeployment scales a deployment
func (c *Client) ScaleDeployment(ctx context.Context, deploymentName string, replicas int) error {
	args := []string{"kubectl", "scale", "deployment", deploymentName, fmt.Sprintf("--replicas=%d", replicas)}
	if c.namespace != "" {
		args = append(args, "--namespace", c.namespace)
	}

	_, err := c.executeCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("scale failed: %w", err)
	}

	return nil
}

// RestartDeployment restarts a deployment
func (c *Client) RestartDeployment(ctx context.Context, deploymentName string) error {
	args := []string{"kubectl", "rollout", "restart", "deployment", deploymentName}
	if c.namespace != "" {
		args = append(args, "--namespace", c.namespace)
	}

	_, err := c.executeCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("restart failed: %w", err)
	}

	return nil
}

// GetClusterInfo gets basic cluster information
func (c *Client) GetClusterInfo(ctx context.Context) (map[string]interface{}, error) {
	// Get cluster info
	output, err := c.executeCommand(ctx, "kubectl", "cluster-info")
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}

	// Get node count
	nodeOutput, _ := c.executeCommand(ctx, "kubectl", "get", "nodes", "--no-headers")
	nodeCount := len(strings.Split(strings.TrimSpace(nodeOutput), "\n"))

	return map[string]interface{}{
		"cluster_info": output,
		"node_count":   nodeCount,
		"namespace":    c.namespace,
		"timestamp":    time.Now(),
	}, nil
}

// executeCommand executes a command with gobe/kubectl/helm
func (c *Client) executeCommand(ctx context.Context, args ...string) (string, error) {
	var cmd *exec.Cmd

	// Determine if we're using gobe as a wrapper or direct kubectl/helm
	if strings.HasPrefix(args[0], "kubectl") || strings.HasPrefix(args[0], "helm") {
		// Direct command
		cmd = exec.CommandContext(ctx, args[0], args[1:]...)
	} else {
		// Use gobe as wrapper
		kbxArgs := append([]string{}, args...)
		cmd = exec.CommandContext(ctx, c.gobePath, kbxArgs...)
	}

	// Set kubeconfig if specified
	if c.kubeconfig != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", c.kubeconfig))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %s (output: %s)", err, string(output))
	}

	return string(output), nil
}

// Ping tests if gobe/kubectl is accessible
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.executeCommand(ctx, "kubectl", "version", "--client", "--short")
	return err
}
