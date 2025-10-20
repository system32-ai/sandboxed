package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/altgen-ai/sandboxed/pkg/k8sclient"
	"github.com/altgen-ai/sandboxed/pkg/sdk"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// ExecuteRequest represents a code execution request
type ExecuteRequest struct {
	Language  string            `json:"language" binding:"required"`
	Code      string            `json:"code" binding:"required"`
	Namespace string            `json:"namespace,omitempty"`
	// Labels    map[string]string `json:"labels,omitempty"`
	SandboxID string            `json:"sandbox_id,omitempty"`
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
	Name string 		  `json:"name,omitempty"`
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
				"sandbox_create":   "POST /api/v1/sandbox/create - Create sandbox",
				"sandbox_execute":  "POST /api/v1/execute - Execute in sandbox",
				"sandbox_destroy":  "POST /api/v1/sandbox/destroy - Destroy sandbox",
			},
		})
	})
	
	// API v1 group
	if k8sClient != nil {
		v1 := r.Group("/api/v1")
		{
			// Sandbox endpoints
			v1.POST("/sandbox/create", func(c *gin.Context) {
				createSandboxHandler(c, k8sClient)
			})
			v1.POST("/execute", func(c *gin.Context) {
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

	id := req.Name + "-" + fmt.Sprintf("%d", time.Now().Unix())

	var opt sdk.SandboxOption

	opt.Name = "namespace"
	opt.Value = req.Namespace

	_, err := sdk.CreateSandbox(id, "python", opt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, SandboxResponse{
			Success:   false,
			Error:     fmt.Sprintf("Failed to create sandbox: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}
	
	c.JSON(http.StatusOK, SandboxResponse{
		Success:   true,
		SandboxID: id,
		Timestamp: time.Now().Format(time.RFC3339),		
	})
}

func executeInSandboxHandler(c *gin.Context, k8sClient *k8sclient.Client) {
	
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Invalid request: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	sdn, err := sdk.NewInstance(req.SandboxID, sdk.SandboxOption{
		Name:  "namespace",
		Value: req.Namespace,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Failed to create sandbox instance: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	output, err := sdn.Run(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ExecuteResponse{
			Success:   false,
			Error:     fmt.Sprintf("Code execution failed: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, ExecuteResponse{
		Success:   true,
		Output:    []string{output.Result},
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
	// for k, v := range req.Labels {
	// 	labels[k] = v
	// }
	
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


func init() {
	rootCmd.AddCommand(serverCmd)
	
	// Add flags
	serverCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")
	serverCmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
	serverCmd.Flags().StringP("namespace", "n", "", "Default Kubernetes namespace")
}