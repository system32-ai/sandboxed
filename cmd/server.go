package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/altgen-ai/sandboxed/pkg/k8sclient"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// ExecuteRequest represents a code execution request
type ExecuteRequest struct {
	Language  string            `json:"language" binding:"required"`
	Code      string            `json:"code" binding:"required"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// ExecuteResponse represents a code execution response
type ExecuteResponse struct {
	Success   bool     `json:"success"`
	Output    []string `json:"output,omitempty"`
	Error     string   `json:"error,omitempty"`
	PodName   string   `json:"pod_name,omitempty"`
	Timestamp string   `json:"timestamp"`
}

// SandboxRequest represents a sandbox creation request
type SandboxRequest struct {
	Language  string            `json:"language" binding:"required"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// SandboxResponse represents a sandbox creation response
type SandboxResponse struct {
	Success     bool   `json:"success"`
	SandboxID   string `json:"sandbox_id,omitempty"`
	Error       string `json:"error,omitempty"`
	Timestamp   string `json:"timestamp"`
}

// PodListResponse represents a pod list response
type PodListResponse struct {
	Pods []PodInfo `json:"pods"`
}

// PodInfo represents basic pod information
type PodInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Image     string            `json:"image,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Created   string            `json:"created"`
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the sandboxed HTTP server",
	Long: `Start the sandboxed HTTP server to handle code execution and Kubernetes operations via REST API.
	
The server provides endpoints for:
- Code execution in Kubernetes pods
- Sandbox management (create, execute, destroy)
- Pod management (list, create, delete)
- Health checks

Examples:
  sandboxed server                    # Start on default port 8080
  sandboxed server --port 3000       # Start on custom port
  sandboxed server --debug           # Start in debug mode`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		debug, _ := cmd.Flags().GetBool("debug")
		namespace, _ := cmd.Flags().GetString("namespace")
		
		// Set gin mode
		if !debug {
			gin.SetMode(gin.ReleaseMode)
		}
		
		// Create gin router
		r := gin.Default()
		
		// Add middleware
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
		r.Use(corsMiddleware())
		
		// Initialize Kubernetes client
		k8sClient, err := k8sclient.NewClient(namespace)
		if err != nil {
			fmt.Printf("Warning: Kubernetes client initialization failed: %v\n", err)
			fmt.Println("Kubernetes endpoints will not be available")
			k8sClient = nil
		}
		
		// Setup routes
		setupRoutes(r, k8sClient)
		
		// Start server
		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Starting sandboxed server on %s\n", addr)
		if debug {
			fmt.Println("Debug mode enabled")
		}
		if k8sClient != nil {
			fmt.Printf("Kubernetes integration enabled (namespace: %s)\n", namespace)
		}
		
		if err := r.Run(addr); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	},
}

func setupRoutes(r *gin.Engine, k8sClient *k8sclient.Client) {
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
			"k8s_available": k8sClient != nil,
		})
	})
	
	// API documentation endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "Sandboxed API",
			"version":     "1.0.0",
			"description": "Code execution and Kubernetes management API",
			"endpoints": gin.H{
				"health":           "GET /health - Health check",
				"execute":          "POST /execute - Execute code directly",
				"sandbox_create":   "POST /api/v1/sandbox/create - Create sandbox",
				"sandbox_execute":  "POST /api/v1/execute/:sandboxID - Execute in sandbox",
				"sandbox_destroy":  "POST /api/v1/sandbox/destroy - Destroy sandbox",
	
			},
		})
	})
	
	// Direct code execution endpoint
	r.POST("/execute", func(c *gin.Context) {
		executeCodeHandler(c, k8sClient)
	})
	
	// API v1 group
	if k8sClient != nil {
		v1 := r.Group("/api/v1")
		{
			// Sandbox endpoints
			v1.POST("/sandbox/create", func(c *gin.Context) {
				createSandboxHandler(c, k8sClient)
			})
			v1.POST("/execute/:sandboxID", func(c *gin.Context) {
				executeInSandboxHandler(c, k8sClient)
			})
			v1.POST("/sandbox/destroy", func(c *gin.Context) {
				destroySandboxHandler(c, k8sClient)
			})
			
		}
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

func executeCodeHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	if k8sClient == nil {
		c.JSON(http.StatusServiceUnavailable, ExecuteResponse{
			Success:   false,
			Error:     "Kubernetes client not available",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Invalid request: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	// Execute code and return result
	result := executeCode(k8sClient, req)
	c.JSON(getStatusCode(result.Success), result)
}

func createSandboxHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	var req SandboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SandboxResponse{
			Success:   false,
			Error:     fmt.Sprintf("Invalid request: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	// Create sandbox pod
	sandboxID := fmt.Sprintf("sandbox-%d", time.Now().Unix())
	image := getImageForLanguage(req.Language)
	if image == "" {
		c.JSON(http.StatusBadRequest, SandboxResponse{
			Success:   false,
			Error:     fmt.Sprintf("Unsupported language: %s", req.Language),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	labels := map[string]string{
		"app":        "sandbox",
		"language":   req.Language,
		"created-by": "sandboxed-api",
		"sandbox-id": sandboxID,
	}
	
	// Add custom labels
	for k, v := range req.Labels {
		labels[k] = v
	}
	
	spec := k8sclient.PodSpec{
		Name:      sandboxID,
		Namespace: req.Namespace,
		Image:     image,
		Command:   []string{"sleep", "3600"}, // Keep container running
		Labels:    labels,
	}
	
	_, err := k8sClient.CreatePod(spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, SandboxResponse{
			Success:   false,
			Error:     fmt.Sprintf("Failed to create sandbox: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	// Wait for pod to be ready
	err = k8sClient.WaitForPodReady(sandboxID, req.Namespace, 2*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, SandboxResponse{
			Success:   false,
			Error:     fmt.Sprintf("Sandbox not ready: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	c.JSON(http.StatusCreated, SandboxResponse{
		Success:   true,
		SandboxID: sandboxID,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func executeInSandboxHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	sandboxID := c.Param("sandboxID")
	
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Invalid request: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	// Execute code in existing sandbox
	command := getCommandForLanguage(req.Language, req.Code)
	if len(command) == 0 {
		c.JSON(http.StatusBadRequest, ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Unsupported language: %s", req.Language),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	output, err := k8sClient.ExecCommand(sandboxID, req.Namespace, command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Execution failed: %v", err),
			PodName:   sandboxID,
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	c.JSON(http.StatusOK, ExecuteResponse{
		Success:   true,
		Output:    []string{output},
		PodName:   sandboxID,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func destroySandboxHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	type DestroyRequest struct {
		SandboxID string `json:"sandbox_id" binding:"required"`
		Namespace string `json:"namespace,omitempty"`
		Force     bool   `json:"force,omitempty"`
	}
	
	var req DestroyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}
	
	var err error
	if req.Force {
		err = k8sClient.ForceDeletePod(req.SandboxID, req.Namespace)
	} else {
		err = k8sClient.DeletePod(req.SandboxID, req.Namespace)
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to destroy sandbox: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Sandbox %s destroyed successfully", req.SandboxID),
	})
}

func executeCode(k8sClient *k8sclient.Client, req ExecuteRequest) ExecuteResponse {
	// Determine image and command based on language
	image := getImageForLanguage(req.Language)
	commands := getCommandsForLanguage(req.Language, req.Code)
	
	if image == "" || len(commands) == 0 {
		return ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Unsupported language: %s. Supported: python, node, go, bash, ruby", req.Language),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}
	
	// Create pod spec
	podName := fmt.Sprintf("api-exec-%d", time.Now().Unix())
	labels := map[string]string{
		"app":        "api-execution",
		"language":   req.Language,
		"created-by": "sandboxed-api",
	}
	
	// Add custom labels
	for k, v := range req.Labels {
		labels[k] = v
	}
	
	spec := k8sclient.PodSpec{
		Name:      podName,
		Namespace: req.Namespace,
		Image:     image,
		Labels:    labels,
	}
	
	// Execute code in pod
	results, err := k8sClient.CreateAndRunPod(spec, commands, true) // cleanup = true
	if err != nil {
		return ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Execution failed: %v", err),
			PodName:   podName,
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}
	
	return ExecuteResponse{
		Success:   true,
		Output:    results,
		PodName:   podName,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func getImageForLanguage(language string) string {
	switch language {
	case "python", "py":
		return "python:3.9-slim"
	case "node", "nodejs", "js":
		return "node:18-slim"
	case "go", "golang":
		return "golang:1.21-alpine"
	case "bash", "sh":
		return "alpine:latest"
	case "ruby", "rb":
		return "ruby:3.0-slim"
	default:
		return ""
	}
}

func getCommandsForLanguage(language, code string) [][]string {
	switch language {
	case "python", "py":
		return [][]string{{"python", "-c", code}}
	case "node", "nodejs", "js":
		return [][]string{{"node", "-e", code}}
	case "go", "golang":
		return [][]string{{"sh", "-c", fmt.Sprintf("echo '%s' > /tmp/main.go && cd /tmp && go run main.go", code)}}
	case "bash", "sh":
		return [][]string{{"sh", "-c", code}}
	case "ruby", "rb":
		return [][]string{{"ruby", "-e", code}}
	default:
		return nil
	}
}

func getCommandForLanguage(language, code string) []string {
	switch language {
	case "python", "py":
		return []string{"python", "-c", code}
	case "node", "nodejs", "js":
		return []string{"node", "-e", code}
	case "go", "golang":
		return []string{"sh", "-c", fmt.Sprintf("echo '%s' > /tmp/main.go && cd /tmp && go run main.go", code)}
	case "bash", "sh":
		return []string{"sh", "-c", code}
	case "ruby", "rb":
		return []string{"ruby", "-e", code}
	default:
		return nil
	}
}

func getStatusCode(success bool) int {
	if success {
		return http.StatusOK
	}
	return http.StatusInternalServerError
}

// Keep existing handlers for pod management
func listPodsHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	namespace := c.Query("namespace")
	
	pods, err := k8sClient.ListPods(namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to list pods: %v", err),
		})
		return
	}
	
	var podInfos []PodInfo
	for _, pod := range pods.Items {
		image := ""
		if len(pod.Spec.Containers) > 0 {
			image = pod.Spec.Containers[0].Image
		}
		
		podInfos = append(podInfos, PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Image:     image,
			Labels:    pod.Labels,
			Created:   pod.CreationTimestamp.Format(time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, PodListResponse{
		Pods: podInfos,
	})
}

func createPodHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	var spec k8sclient.PodSpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid pod spec: %v", err),
		})
		return
	}
	
	pod, err := k8sClient.CreatePod(spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create pod: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Pod created successfully",
		"pod_name":  pod.Name,
		"namespace": pod.Namespace,
	})
}

func deletePodHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	podName := c.Param("name")
	namespace := c.Query("namespace")
	force := c.Query("force") == "true"
	
	var err error
	if force {
		err = k8sClient.ForceDeletePod(podName, namespace)
	} else {
		err = k8sClient.DeletePod(podName, namespace)
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to delete pod: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Pod %s deleted successfully", podName),
	})
}

func getPodHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	podName := c.Param("name")
	namespace := c.Query("namespace")
	
	pod, err := k8sClient.GetPod(podName, namespace)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Pod not found: %v", err),
		})
		return
	}
	
	image := ""
	if len(pod.Spec.Containers) > 0 {
		image = pod.Spec.Containers[0].Image
	}
	
	c.JSON(http.StatusOK, PodInfo{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
		Image:     image,
		Labels:    pod.Labels,
		Created:   pod.CreationTimestamp.Format(time.RFC3339),
	})
}

func getPodLogsHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	podName := c.Param("name")
	namespace := c.Query("namespace")
	
	logs, err := k8sClient.GetPodLogs(podName, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get pod logs: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"pod_name": podName,
		"logs":     logs,
	})
}

func init() {
	rootCmd.AddCommand(serverCmd)
	
	// Add flags
	serverCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")
	serverCmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
	serverCmd.Flags().StringP("namespace", "n", "", "Default Kubernetes namespace")
}(
	"fmt"
	"net/http"
	"time"

	"github.com/altgen-ai/sandboxed/pkg/k8sclient"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// ExecuteRequest represents a code execution request
type ExecuteRequest struct {
	Language  string            `json:"language" binding:"required"`
	Code      string            `json:"code" binding:"required"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// ExecuteResponse represents a code execution response
type ExecuteResponse struct {
	Success   bool     `json:"success"`
	Output    []string `json:"output,omitempty"`
	Error     string   `json:"error,omitempty"`
	PodName   string   `json:"pod_name,omitempty"`
	Timestamp string   `json:"timestamp"`
}

// PodListResponse represents a pod list response
type PodListResponse struct {
	Pods []PodInfo `json:"pods"`
}

// PodInfo represents basic pod information
type PodInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Image     string            `json:"image,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Created   string            `json:"created"`
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the sandboxed HTTP server",
	Long: `Start the sandboxed HTTP server to handle code execution and Kubernetes operations via REST API.
	
The server provides endpoints for:
- Code execution in Kubernetes pods
- Pod management (list, create, delete)
- Health checks

Examples:
  sandboxed server                    # Start on default port 8080
  sandboxed server --port 3000       # Start on custom port
  sandboxed server --debug           # Start in debug mode`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		debug, _ := cmd.Flags().GetBool("debug")
		namespace, _ := cmd.Flags().GetString("namespace")
		
		// Set gin mode
		if !debug {
			gin.SetMode(gin.ReleaseMode)
		}
		
		// Create gin router
		r := gin.Default()
		
		// Add middleware
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
		r.Use(corsMiddleware())
		
		// Initialize Kubernetes client
		k8sClient, err := k8sclient.NewClient(namespace)
		if err != nil {
			fmt.Printf("Warning: Kubernetes client initialization failed: %v\n", err)
			fmt.Println("Kubernetes endpoints will not be available")
			k8sClient = nil
		}
		
		// Health check endpoint
		r.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().Format(time.RFC3339),
				"version":   "1.0.0",
			})
		})
		
		// Kubernetes pod endpoints
		if k8sClient != nil {
			k8sGroup := r.Group("/api/v1")
			{
				k8sGroup.GET("/sandbox/create", func(c *gin.Context) {
					listPodsHandler(c, k8sClient)
				})
				k8sGroup.POST("/execute/:sandboxID", func(c *gin.Context) {
					createPodHandler(c, k8sClient)
				})
				k8sGroup.POST("/sandbox/destroy", func(c *gin.Context) {
					deletePodHandler(c, k8sClient)
				})
			}
		}
		
		// API documentation endpoint
		r.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"name":        "Sandboxed API",
				"version":     "1.0.0",
				"description": "Code execution and Kubernetes management API",
				"endpoints": gin.H{
					"health":     "GET /health - Health check",
					"execute":    "POST /execute - Execute code",
					"pods":       "GET /k8s/pods - List pods",
					"create_pod": "POST /k8s/pods - Create pod",
					"delete_pod": "DELETE /k8s/pods/:name - Delete pod",
					"get_pod":    "GET /k8s/pods/:name - Get pod details",
					"pod_logs":   "GET /k8s/pods/:name/logs - Get pod logs",
				},
			})
		})
		
		// Start server
		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Starting sandboxed server on %s\n", addr)
		if debug {
			fmt.Println("Debug mode enabled")
		}
		if k8sClient != nil {
			fmt.Printf("Kubernetes integration enabled (namespace: %s)\n", namespace)
		}
		
		if err := r.Run(addr); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	},
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

func executeCodeHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	if k8sClient == nil {
		c.JSON(http.StatusServiceUnavailable, ExecuteResponse{
			Success:   false,
			Error:     "Kubernetes client not available",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Invalid request: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	// Determine image based on language
	var image string
	var commands [][]string
	
	switch req.Language {
	case "python", "py":
		image = "python:3.9-slim"
		commands = [][]string{{"python", "-c", req.Code}}
	case "node", "nodejs", "js":
		image = "node:18-slim"
		commands = [][]string{{"node", "-e", req.Code}}
	case "go", "golang":
		image = "golang:1.21-alpine"
		commands = [][]string{{"sh", "-c", fmt.Sprintf("echo '%s' > /tmp/main.go && cd /tmp && go run main.go", req.Code)}}
	case "bash", "sh":
		image = "alpine:latest"
		commands = [][]string{{"sh", "-c", req.Code}}
	case "ruby", "rb":
		image = "ruby:3.0-slim"
		commands = [][]string{{"ruby", "-e", req.Code}}
	default:
		c.JSON(http.StatusBadRequest, ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Unsupported language: %s. Supported: python, node, go, bash, ruby", req.Language),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	// Create pod spec
	podName := fmt.Sprintf("api-exec-%d", time.Now().Unix())
	labels := map[string]string{
		"app":        "api-execution",
		"language":   req.Language,
		"created-by": "sandboxed-api",
	}
	
	// Add custom labels
	for k, v := range req.Labels {
		labels[k] = v
	}
	
	spec := k8sclient.PodSpec{
		Name:      podName,
		Namespace: req.Namespace,
		Image:     image,
		Labels:    labels,
	}
	
	// Execute code in pod
	results, err := k8sClient.CreateAndRunPod(spec, commands, true) // cleanup = true
	if err != nil {
		c.JSON(http.StatusInternalServerError, ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Execution failed: %v", err),
			PodName:   podName,
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	c.JSON(http.StatusOK, ExecuteResponse{
		Success:   true,
		Output:    results,
		PodName:   podName,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func listPodsHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	namespace := c.Query("namespace")
	
	pods, err := k8sClient.ListPods(namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to list pods: %v", err),
		})
		return
	}
	
	var podInfos []PodInfo
	for _, pod := range pods.Items {
		image := ""
		if len(pod.Spec.Containers) > 0 {
			image = pod.Spec.Containers[0].Image
		}
		
		podInfos = append(podInfos, PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Image:     image,
			Labels:    pod.Labels,
			Created:   pod.CreationTimestamp.Format(time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, PodListResponse{
		Pods: podInfos,
	})
}

func createPodHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	var spec k8sclient.PodSpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid pod spec: %v", err),
		})
		return
	}
	
	pod, err := k8sClient.CreatePod(spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create pod: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Pod created successfully",
		"pod_name":  pod.Name,
		"namespace": pod.Namespace,
	})
}

func deletePodHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	podName := c.Param("name")
	namespace := c.Query("namespace")
	force := c.Query("force") == "true"
	
	var err error
	if force {
		err = k8sClient.ForceDeletePod(podName, namespace)
	} else {
		err = k8sClient.DeletePod(podName, namespace)
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to delete pod: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Pod %s deleted successfully", podName),
	})
}

func getPodHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	podName := c.Param("name")
	namespace := c.Query("namespace")
	
	pod, err := k8sClient.GetPod(podName, namespace)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Pod not found: %v", err),
		})
		return
	}
	
	image := ""
	if len(pod.Spec.Containers) > 0 {
		image = pod.Spec.Containers[0].Image
	}
	
	c.JSON(http.StatusOK, PodInfo{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
		Image:     image,
		Labels:    pod.Labels,
		Created:   pod.CreationTimestamp.Format(time.RFC3339),
	})
}

func getPodLogsHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	podName := c.Param("name")
	namespace := c.Query("namespace")
	
	logs, err := k8sClient.GetPodLogs(podName, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get pod logs: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"pod_name": podName,
		"logs":     logs,
	})
}

func init() {
	rootCmd.AddCommand(serverCmd)
	
	// Add flags
	serverCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")
	serverCmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
	serverCmd.Flags().StringP("namespace", "n", "", "Default Kubernetes namespace")
}