package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/altgen-ai/sandboxed/pkg/sdk"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SandboxManager manages the state of all active sandboxes
type SandboxManager struct {
	mu        sync.RWMutex
	sandboxes map[string]sdk.Sandboxed
}

// NewSandboxManager creates a new sandbox manager
func NewSandboxManager() *SandboxManager {
	return &SandboxManager{
		sandboxes: make(map[string]sdk.Sandboxed),
	}
}

// AddSandbox adds a sandbox to the manager
func (sm *SandboxManager) AddSandbox(name string, sandbox sdk.Sandboxed) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sandboxes[name] = sandbox
}

// GetSandbox retrieves a sandbox by name
func (sm *SandboxManager) GetSandbox(name string) (sdk.Sandboxed, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sandbox, exists := sm.sandboxes[name]
	return sandbox, exists
}

// RemoveSandbox removes a sandbox from the manager
func (sm *SandboxManager) RemoveSandbox(name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sandboxes, name)
}

// ListSandboxes returns all sandbox names
func (sm *SandboxManager) ListSandboxes() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	names := make([]string, 0, len(sm.sandboxes))
	for name := range sm.sandboxes {
		names = append(names, name)
	}
	return names
}

// NewServer creates a new MCP server with sandbox tools
func NewServer() *mcp.Server {
	// Create MCP server with proper Implementation struct
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "sandboxed",
		Version: "1.0.0",
	}, nil)

	// Create sandbox manager for state tracking
	sandboxManager := NewSandboxManager()

	// Register sandbox tools
	registerSandboxTools(server, sandboxManager)

	return server
}

// registerSandboxTools registers all sandbox-related tools using the MCP SDK
func registerSandboxTools(server *mcp.Server, sandboxManager *SandboxManager) {
	// Register create_sandbox tool
	type CreateSandboxArgs struct {
		Name      string            `json:"name"`
		Language  string            `json:"language"`
		Namespace string            `json:"namespace,omitempty"`
		Labels    map[string]string `json:"labels,omitempty"`
	}

	type CreateSandboxResult struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_sandbox",
		Description: "Creates a new sandbox environment for code execution",
	}, func(ctx context.Context, request *mcp.CallToolRequest, args CreateSandboxArgs) (*mcp.CallToolResult, CreateSandboxResult, error) {
		// Check if sandbox already exists
		if _, exists := sandboxManager.GetSandbox(args.Name); exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Sandbox '%s' already exists", args.Name)},
				},
			}, CreateSandboxResult{Success: false, Message: "Sandbox already exists"}, nil
		}

		// Prepare sandbox options
		var opts []sdk.SandboxOption
		if args.Namespace != "" {
			opts = append(opts, sdk.SandboxOption{Name: "namespace", Value: args.Namespace})
		}
		if args.Labels != nil {
			opts = append(opts, sdk.SandboxOption{Name: "labels", Value: args.Labels})
		}

		lang, err := sdk.ToLanguage(args.Language)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Invalid language: %v", err)},
				},
			}, CreateSandboxResult{Success: false, Message: err.Error()}, nil
		}

		// Create sandbox
		sandbox, err := sdk.CreateSandbox(args.Name, lang, opts...)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to create sandbox '%s': %v", args.Name, err)},
				},
			}, CreateSandboxResult{Success: false, Message: err.Error()}, nil
		}

		// Add to manager
		sandboxManager.AddSandbox(args.Name, sandbox)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Successfully created sandbox '%s' with language '%s'", args.Name, args.Language)},
			},
		}, CreateSandboxResult{Success: true, Message: "Sandbox created successfully"}, nil
	})

	// Register run_code tool
	type RunCodeArgs struct {
		SandboxName string `json:"sandbox_name"`
		Code        string `json:"code"`
	}

	type RunCodeResult struct {
		Success  bool   `json:"success"`
		Output   string `json:"output,omitempty"`
		ExitCode int    `json:"exit_code,omitempty"`
		Error    string `json:"error,omitempty"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_code",
		Description: "Executes code in an existing sandbox environment",
	}, func(ctx context.Context, request *mcp.CallToolRequest, args RunCodeArgs) (*mcp.CallToolResult, RunCodeResult, error) {
		// Get sandbox
		sandbox, exists := sandboxManager.GetSandbox(args.SandboxName)
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Sandbox '%s' not found. Use create_sandbox first.", args.SandboxName)},
				},
			}, RunCodeResult{Success: false, Error: "Sandbox not found"}, nil
		}

		// Run code
		output, err := sandbox.Run(args.Code)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to execute code in sandbox '%s': %v", args.SandboxName, err)},
				},
			}, RunCodeResult{Success: false, Error: err.Error()}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Code executed successfully in sandbox '%s':\n\nOutput:\n%s\n\nExit Code: %d", 
					args.SandboxName, output.Result, output.ExitCode)},
			},
		}, RunCodeResult{Success: true, Output: output.Result, ExitCode: output.ExitCode}, nil
	})

	// Register destroy_sandbox tool
	type DestroySandboxArgs struct {
		SandboxName string `json:"sandbox_name"`
	}

	type DestroySandboxResult struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "destroy_sandbox",
		Description: "Destroys an existing sandbox environment and cleans up resources",
	}, func(ctx context.Context, request *mcp.CallToolRequest, args DestroySandboxArgs) (*mcp.CallToolResult, DestroySandboxResult, error) {
		// Get sandbox
		sandbox, exists := sandboxManager.GetSandbox(args.SandboxName)
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Sandbox '%s' not found", args.SandboxName)},
				},
			}, DestroySandboxResult{Success: false, Message: "Sandbox not found"}, nil
		}

		// Destroy sandbox
		if err := sandbox.Destroy(); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to destroy sandbox '%s': %v", args.SandboxName, err)},
				},
			}, DestroySandboxResult{Success: false, Message: err.Error()}, nil
		}

		// Remove from manager
		sandboxManager.RemoveSandbox(args.SandboxName)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Successfully destroyed sandbox '%s'", args.SandboxName)},
			},
		}, DestroySandboxResult{Success: true, Message: "Sandbox destroyed successfully"}, nil
	})

	// Register list_sandboxes tool
	type ListSandboxesArgs struct{}

	type ListSandboxesResult struct {
		Sandboxes []string `json:"sandboxes"`
		Count     int      `json:"count"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sandboxes",
		Description: "Lists all active sandbox environments",
	}, func(ctx context.Context, request *mcp.CallToolRequest, args ListSandboxesArgs) (*mcp.CallToolResult, ListSandboxesResult, error) {
		sandboxes := sandboxManager.ListSandboxes()
		
		if len(sandboxes) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "No active sandboxes found"},
				},
			}, ListSandboxesResult{Sandboxes: []string{}, Count: 0}, nil
		}

		result := fmt.Sprintf("Active sandboxes (%d):\n", len(sandboxes))
		for i, name := range sandboxes {
			result += fmt.Sprintf("%d. %s\n", i+1, name)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, ListSandboxesResult{Sandboxes: sandboxes, Count: len(sandboxes)}, nil
	})
}

// Run starts the MCP server on stdio transport
func RunServer(server *mcp.Server) error {
	log.Println("Starting MCP server for sandboxed code execution...")
	return server.Run(context.Background(), &mcp.StdioTransport{})
}

// RunServerSSE starts the MCP server with SSE (Server-Sent Events) support
func RunServerSSE(server *mcp.Server, port int) error {
	log.Printf("Starting MCP server with SSE support on port %d...", port)
	
	// Create SSE handler
	mcpSSEHandler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
		return server
	}, &mcp.SSEOptions{})
	
	// Wrap the SSE handler to ensure proper headers
	sseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Cache-Control")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Delegate to the MCP SSE handler
		mcpSSEHandler.ServeHTTP(w, r)
	})
	
	// Set up HTTP routes
	http.Handle("/sse", sseHandler)
	
	// Add health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "sandboxed-mcp"}`))
	})
	
	// Add CORS support for web clients
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Serve simple info page
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Sandboxed MCP Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 5px 0; border-radius: 4px; }
        .method { font-weight: bold; color: #0066cc; }
    </style>
</head>
<body>
    <h1>Sandboxed MCP Server</h1>
    <p>This is an MCP (Model Context Protocol) server providing sandbox management tools.</p>
    <h2>Available Tools:</h2>
    <ul>
        <li><strong>create_sandbox</strong> - Create a new sandbox environment</li>
        <li><strong>run_code</strong> - Execute code in an existing sandbox</li>
        <li><strong>destroy_sandbox</strong> - Destroy a sandbox and clean up resources</li>
        <li><strong>list_sandboxes</strong> - List all active sandboxes</li>
    </ul>
    <h2>Endpoints:</h2>
    <div class="endpoint">
        <span class="method">GET/POST</span> <code>/sse</code> - SSE endpoint for MCP communication
    </div>
    <div class="endpoint">
        <span class="method">GET</span> <code>/health</code> - Health check endpoint
    </div>
    <h2>Usage:</h2>
    <p>Connect using an MCP client with SSE transport to: <code>http://localhost:` + fmt.Sprintf("%d", port) + `/sse</code></p>
    <p>Example using curl to test the health endpoint:</p>
    <pre>curl http://localhost:` + fmt.Sprintf("%d", port) + `/health</pre>
</body>
</html>
			`))
			return
		}
		
		http.NotFound(w, r)
	})
	
	addr := fmt.Sprintf(":%d", port)
	log.Printf("MCP SSE server listening on %s", addr)
	return http.ListenAndServe(addr, nil)
}