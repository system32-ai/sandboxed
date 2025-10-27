# Sandboxed

A comprehensive sandbox platform for secure code execution in Kubernetes environments. Provides REST API, MCP (Model Context Protocol) server with SSE support, and Go SDK for running code in isolated containers.
<img width="3267" height="632" alt="Group 17 (3)" src="https://github.com/user-attachments/assets/0ce72715-9509-474e-9e8f-e09f5c88c466" />

## Features

- 🔒 **Secure Execution**: Code runs in isolated Kubernetes pods with RBAC controls
- 🌐 **Multiple Interfaces**: REST API, MCP server (stdio + SSE), and Go SDK
- 🐍 **Multi-Language Support**: Python, Go, Node.js, Java, Ruby, PHP, Rust
- 🤖 **AI Integration**: Built-in MCP server for AI assistants and coding agents
- 📡 **Production Ready**: Cross-platform builds, Docker images, and CI/CD automation
- 🛠️ **Flexible Configuration**: Customizable namespaces, labels, and resource limits
- ⚡ **SSE Support**: Server-Sent Events transport for web-based MCP clients
- 📦 **Enhanced Code Execution**: File-based script execution with language-specific interpreters
## Getting Started

### Prerequisites
- Go 1.24 or later
- Kubernetes cluster access (for sandbox execution)
- Docker (optional, for containerized deployment)



### Building the Application
```bash
go build -o sandboxed
./sandboxed --help
```

## CLI Commands

The `sandboxed` CLI provides several commands for different use cases:

### Core Commands

- **`mcp`**: Start MCP (Model Context Protocol) server for AI assistant integration
- **`server`**: Start REST API server for HTTP-based sandbox management
- **`exec [command]`**: Execute shell commands with optional directory and environment settings
- **`code [path]`**: Open files or directories in a code editor (VS Code by default)
- **`version`**: Display application version
- **`help`**: Show help for any command

### Command Examples

```bash
# Show version information
./sandboxed version

# Start REST API server on custom port
./sandboxed server --port 9000

# Start MCP server for AI integration
./sandboxed mcp

# Get help for any command
./sandboxed server --help
./sandboxed mcp --help
```


## Using the Go SDK

The Go SDK provides programmatic access to sandbox functionality.

### Basic Usage

```go
package main

import (
	"log"

	"github.com/system32-ai/sandboxed/pkg/sdk"
)

func main() {
	// Create a Python sandbox
	sandbox, err := sdk.CreateSandbox("my-python-sandbox", "python")
	if err != nil {
		log.Fatalf("failed to create sandbox: %v", err)
	}

	// Always clean up resources
	defer sandbox.Destroy()

	// Execute Python code
	code := `
import json
import sys

data = {"message": "Hello, World!", "python_version": sys.version}
print(json.dumps(data, indent=2))
`

	output, err := sandbox.Exec(code)
	if err != nil {
		log.Fatalf("failed to run code: %v", err)
	}

	log.Printf("Output: %s", output.Result)
	log.Printf("Exit Code: %d", output.ExitCode)
}
```

### Advanced Usage with Options

```go
package main

import (
	"log"

	"github.com/system32-ai/sandboxed/pkg/sdk"
)

func main() {
	// Create sandbox with custom options
	opts := []sdk.SandboxOption{
		{Name: "namespace", Value: "development"},
		{Name: "labels", Value: map[string]string{
			"project":     "my-app",
			"environment": "dev",
			"owner":       "my-team",
		}},
	}

	sandbox, err := sdk.CreateSandbox("advanced-sandbox", "python", opts...)
	if err != nil {
		log.Fatalf("failed to create sandbox: %v", err)
	}
	defer sandbox.Destroy()

	// Install packages and run code
	setupCode := `
pip install requests numpy
python -c "import requests, numpy; print('Packages installed successfully')"
`

	output, err := sandbox.Run(setupCode)
	if err != nil {
		log.Fatalf("setup failed: %v", err)
	}
	log.Printf("Setup result: %s", output.Result)

	// Use the installed packages
	mainCode := `
import requests
import numpy as np

# Create a simple array
arr = np.array([1, 2, 3, 4, 5])
print(f"Array: {arr}")
print(f"Mean: {np.mean(arr)}")

# Make a simple HTTP request (if network is available)
try:
    resp = requests.get("https://httpbin.org/json", timeout=5)
    print(f"HTTP Status: {resp.status_code}")
except:
    print("Network request failed (expected in isolated environment)")
`

	output, err = sandbox.Run(mainCode)
	if err != nil {
		log.Fatalf("main code failed: %v", err)
	}
	log.Printf("Main result: %s", output.Result)
}
```

### Multiple Language Support

```go
package main

import (
	"log"

	"github.com/system32-ai/sandboxed/pkg/sdk"
)

func runLanguageExample(name, language, code string) {
	log.Printf("=== %s Example ===", name)
	
	sandbox, err := sdk.CreateSandbox(name+"-sandbox", language)
	if err != nil {
		log.Printf("Failed to create %s sandbox: %v", name, err)
		return
	}
	defer sandbox.Destroy()

	output, err := sandbox.Exec(code)
	if err != nil {
		log.Printf("Failed to run %s code: %v", name, err)
		return
	}

	log.Printf("Output: %s", output.Result)
	log.Printf("Exit Code: %d\n", output.ExitCode)
}

func main() {
	// Python example
	pythonCode := `
print("Hello from Python!")
import sys
print(f"Python version: {sys.version}")
`
	runLanguageExample("Python", "python", pythonCode)

	// JavaScript/Node.js example
	jsCode := `
console.log("Hello from Node.js!");
console.log("Node version:", process.version);
console.log("Platform:", process.platform);
`
	runLanguageExample("JavaScript", "javascript", jsCode)

	// Go example
	goCode := `
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("Hello from Go!")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
`
	runLanguageExample("Go", "go", goCode)
}
```

### Error Handling

```go
package main

import (
	"log"

	"github.com/system32-ai/sandboxed/pkg/sdk"
)

func main() {
	sandbox, err := sdk.CreateSandbox("error-handling-example", "python")
	if err != nil {
		log.Fatalf("failed to create sandbox: %v", err)
	}
	defer sandbox.Destroy()

	// Example of code that will fail
	badCode := `
print("This will work")
undefined_variable_that_causes_error
print("This won't be reached")
`

	output, err := sandbox.Run(badCode)
	if err != nil {
		log.Printf("Execution failed: %v", err)
		return
	}

	// Check exit code to determine success/failure
	if output.ExitCode != 0 {
		log.Printf("Code execution failed with exit code: %d", output.ExitCode)
		log.Printf("Output: %s", output.Result)
	} else {
		log.Printf("Code executed successfully: %s", output.Result)
	}
}
```

### SDK API Reference

#### Supported Languages

The SDK automatically detects the appropriate container image and execution method for each language:

- `python`: Python 3.9 with pip (`python3 script.py`)
- `go`: Go 1.24 compiler and tools (`go run script.go`)
- `node`: Node.js 14 with npm (`node script.js`)
- `java`: OpenJDK 11 with compilation (`javac + java`)
- `rust`: Rust 1.56 compiler (`rustc + executable`)
- `ruby`: Ruby 2.7 interpreter (`ruby script.rb`)
- `php`: PHP 8.0 interpreter (`php script.php`)

#### Enhanced Execution

The `Exec()` method now writes code to temporary files with proper language extensions and executes them using language-specific interpreters for better error handling and multi-line code support.

## REST API Server

The REST API server provides HTTP endpoints for sandbox management.

### Starting the REST Server

```bash
# Start the REST server (default port 8080)
./sandboxed server

# Or using go run
go run main.go server
```

### API Endpoints

#### POST /execute

Execute code in a temporary sandbox environment.

**Request Body:**
```json
{
  "language": "python",
  "code": "print('Hello, World!')",
  "namespace": "default",
  "labels": {
    "project": "api-test"
  }
}
```

**Response:**
```json
{
  "result": "Hello, World!\n",
  "error": "",
  "exit_code": 0,
  "execution_time_ms": 1250
}
```

#### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-20T10:30:00Z"
}
```

### API Usage Examples

#### Using curl

```bash
# Execute Python code
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "language": "python",
    "code": "import sys\nprint(f\"Python version: {sys.version}\")"
  }'

# Execute JavaScript code
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "language": "javascript",
    "code": "console.log(\"Hello from Node.js!\");"
  }'
```

#### Using Go HTTP client

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type ExecuteRequest struct {
    Language string `json:"language"`
    Code     string `json:"code"`
}

type ExecuteResponse struct {
    Result         string `json:"result"`
    Error          string `json:"error"`
    ExitCode       int    `json:"exit_code"`
    ExecutionTimeMs int   `json:"execution_time_ms"`
}

func main() {
    req := ExecuteRequest{
        Language: "python",
        Code:     "print('Hello from REST API!')",
    }
    
    jsonData, _ := json.Marshal(req)
    
    resp, err := http.Post(
        "http://localhost:8080/execute",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    var result ExecuteResponse
    json.NewDecoder(resp.Body).Decode(&result)
    
    fmt.Printf("Output: %s\n", result.Result)
    fmt.Printf("Exit Code: %d\n", result.ExitCode)
}
```

### SSE (Server-Sent Events) Support

For web-based applications, you can connect to the MCP server using SSE transport:

```javascript
// Connect to MCP server via SSE
const sseUrl = 'http://localhost:8080/sse';

// Example using a hypothetical MCP JavaScript client
const mcpClient = new MCPClient({
  transport: 'sse',
  url: sseUrl
});

// Create sandbox and run code
await mcpClient.callTool('create_sandbox', {
  name: 'web-sandbox',
  language: 'python'
});

const result = await mcpClient.callTool('run_code', {
  sandbox_name: 'web-sandbox',
  code: 'print("Hello from web client!")'
});

console.log(result.output);
```


## MCP (Model Context Protocol) Server

The MCP server provides sandbox management tools for AI assistants and other clients that implement the Model Context Protocol. The server runs on stdio transport and communicates via JSON-RPC 2.0.

### Starting the MCP Server

The MCP server supports two transport modes:

#### 1. Stdio Mode (Default)
For AI assistants and command-line MCP clients:
```bash
# Start the MCP server (stdio mode)
./sandboxed mcp

# Or using go run
go run main.go mcp
```

#### 2. SSE (Server-Sent Events) Mode
For web-based clients and HTTP transport:
```bash
# Start in SSE mode on default port 8080
./sandboxed mcp --sse

# Start in SSE mode on custom port
./sandboxed mcp --sse --port 9000

# Or using go run
go run main.go mcp --sse --port 8080
```

The SSE mode provides a web interface at `http://localhost:8080` with API documentation and connection details.

### Available MCP Tools

The server provides the following tools for sandbox management:

#### 1. create_sandbox

Creates a new sandbox environment for code execution.

**Parameters:**
- `name` (string, required): Unique name for the sandbox
- `language` (string, required): Programming language for the sandbox (e.g., "python", "javascript", "go", "java")
- `namespace` (string, optional): Kubernetes namespace (defaults to "default")
- `labels` (object, optional): Additional labels for the sandbox pod

**Example:**
```json
{
  "name": "python-sandbox-1",
  "language": "python",
  "namespace": "default",
  "labels": {
    "project": "ai-assistant",
    "environment": "development"
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Sandbox created successfully"
}
```

#### 2. run_code

Executes code in an existing sandbox environment.

**Parameters:**
- `sandbox_name` (string, required): Name of the sandbox to run code in
- `code` (string, required): Code to execute in the sandbox

**Example:**
```json
{
  "sandbox_name": "python-sandbox-1",
  "code": "print('Hello from sandbox!')\nresult = 2 + 2\nprint(f'2 + 2 = {result}')"
}
```

**Response:**
```json
{
  "success": true,
  "output": "Hello from sandbox!\n2 + 2 = 4\n",
  "exit_code": 0
}
```

#### 3. destroy_sandbox

Destroys an existing sandbox environment and cleans up resources.

**Parameters:**
- `sandbox_name` (string, required): Name of the sandbox to destroy

**Example:**
```json
{
  "sandbox_name": "python-sandbox-1"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Sandbox destroyed successfully"
}
```

#### 4. list_sandboxes

Lists all active sandbox environments.

**Parameters:** None

**Response:**
```json
{
  "sandboxes": ["python-sandbox-1", "javascript-sandbox-2"],
  "count": 2
}
```

### MCP Client Integration

To integrate with the MCP server, use any MCP-compatible client. Here's an example using the Go MCP SDK:

```go
package main

import (
    "context"
    "log"
    "os/exec"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    // Create MCP client
    client := mcp.NewClient(&mcp.Implementation{
        Name: "sandbox-client", 
        Version: "1.0.0"
    }, nil)
    
    // Connect to sandboxed MCP server
    transport := &mcp.CommandTransport{
        Command: exec.Command("./sandboxed", "mcp"),
    }
    
    session, err := client.Connect(context.Background(), transport, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close()
    
    // Create a sandbox
    createResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
        Name: "create_sandbox",
        Arguments: map[string]any{
            "name": "my-python-sandbox",
            "language": "python",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Create result: %v", createResult)
    
    // Run some code
    runResult, err := session.CallTool(context.Background(), &mcp.CallToolParams{
        Name: "run_code",
        Arguments: map[string]any{
            "sandbox_name": "my-python-sandbox",
            "code": "print('Hello from MCP!')",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Run result: %v", runResult)
    
    // Clean up
    _, err = session.CallTool(context.Background(), &mcp.CallToolParams{
        Name: "destroy_sandbox",
        Arguments: map[string]any{
            "sandbox_name": "my-python-sandbox",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### Supported Languages

The sandbox supports the following programming languages with enhanced execution:

- **Python 3.9**: Full environment with pip package management
- **Go 1.24**: Complete Go development environment with compiler
- **Node.js 14**: JavaScript runtime with npm package management
- **Java 11**: OpenJDK with compilation and execution support
- **Rust 1.56**: Rust compiler with cargo build tools
- **Ruby 2.7**: Ruby interpreter with gem support
- **PHP 8.0**: PHP interpreter with composer support

Each language uses optimized container images and language-specific execution methods for better performance and error handling.

### Error Handling

All MCP tools return structured responses with success/failure indicators:

- **Success**: `success: true` with relevant data
- **Failure**: `success: false` with error message in `error` or `message` field

### Security Considerations

- Sandboxes run in isolated Kubernetes pods
- Network access is limited based on cluster configuration
- Resource limits can be applied via Kubernetes resource quotas
- Sandbox cleanup is automatic when tools complete

## Deployment

### Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster with proper RBAC permissions
- **kubectl**: Configured to connect to your cluster
- **Container Runtime**: Docker or other Kubernetes-compatible runtime
- **Go 1.24+**: For building from source

### Kubernetes RBAC Setup

Create the necessary RBAC permissions for the sandboxed service:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sandboxed
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: sandboxed
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["create", "delete", "get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: sandboxed
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: sandboxed
subjects:
- kind: ServiceAccount
  name: sandboxed
  namespace: default
```

### Container Deployment

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o sandboxed .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/sandboxed .

EXPOSE 8080
CMD ["./sandboxed", "server"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sandboxed
  labels:
    app: sandboxed
spec:
  replicas: 3
  selector:
    matchLabels:
      app: sandboxed
  template:
    metadata:
      labels:
        app: sandboxed
    spec:
      serviceAccountName: sandboxed
      containers:
      - name: sandboxed
        image: your-registry/sandboxed:latest
        ports:
        - containerPort: 8080
        env:
        - name: KUBECONFIG
          value: ""
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 250m
            memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: sandboxed-service
spec:
  selector:
    app: sandboxed
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Configuration

### Environment Variables

- `KUBECONFIG`: Path to kubeconfig file (optional if running in-cluster)
- `DEFAULT_NAMESPACE`: Default namespace for sandbox pods (default: "default")
- `SANDBOX_TIMEOUT`: Default timeout for sandbox operations (default: "120s")
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

### Language Container Images

The system uses predefined container images for different languages in `pkg/k8sclient/templates/lang.go`:

```go
var languageImages = map[string]string{
    "go":     "golang:1.24",
    "python": "python:3.9",
    "node":   "node:14",
    "java":   "openjdk:11",
    "ruby":   "ruby:2.7",
    "php":    "php:8.0",
    "rust":   "rust:1.56",
}
```

## Troubleshooting

### Common Issues

#### 1. Pod Creation Fails

**Error**: `failed to create sandbox: pods is forbidden`

**Solution**: Ensure proper RBAC permissions are configured (see Deployment section).

#### 2. Sandbox Timeout

**Error**: `timeout waiting for pod to be ready`

**Solution**: 
- Check if container images are available and can be pulled
- Increase timeout values
- Verify cluster resources are sufficient

#### 3. Code Execution Fails

**Error**: `failed to execute code: command not found`

**Solution**: 
- Verify the correct language container image is being used
- Check if required tools are installed in the container
- Ensure the language syntax is correct

#### 4. Network Issues

**Error**: `network requests fail in sandbox`

**Solution**:
- Network access is limited by default for security
- Configure NetworkPolicies if external access is needed
- Use appropriate cluster configuration for your networking requirements

#### 5. SSE Connection Issues

**Error**: `SSE connection fails or times out`

**Solution**:
- Ensure MCP server is running in SSE mode (`--sse` flag)
- Check firewall settings and port accessibility
- Verify CORS headers if connecting from a web browser
- Use the web interface at `http://localhost:8080` to test connectivity

### Debug Mode

Enable debug logging to troubleshoot issues:

```bash
export LOG_LEVEL=debug
./sandboxed mcp
```

### Health Checks

```bash
# Check if server is running
curl http://localhost:8080/health

# List active pods (requires kubectl access)
kubectl get pods -l created-by=sandboxed-sdk

# Check logs
kubectl logs -l app=sandboxed -f
```

## Release and Distribution

The project includes automated release pipelines for cross-platform distribution:

### Automated Releases

Create releases automatically by pushing tags:

```bash
# Create and push a release tag
git tag v1.0.0
git push origin v1.0.0

# Or use the release script
./scripts/release.sh v1.0.0
```

### Available Distributions

Each release provides:
- **Cross-platform binaries**: Linux, macOS, Windows (AMD64, ARM64)
- **Docker images**: Multi-arch containers via GitHub Container Registry
- **Automated changelog**: Generated from commit history
- **GitHub Actions**: Complete CI/CD pipeline

### Docker Images

```bash
# Pull the latest image
docker pull ghcr.io/altgen-ai/sandboxed:latest

# Run the container
docker run -p 8080:8080 ghcr.io/altgen-ai/sandboxed:latest
```

### Manual Building

```bash
# Build for current platform
go build -o sandboxed .

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o sandboxed-linux-amd64 .
GOOS=darwin GOARCH=arm64 go build -o sandboxed-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o sandboxed-windows-amd64.exe .
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/system32-ai/sandboxed.git
cd sandboxed

# Install dependencies
go mod download

# Run tests
go test ./...

# Build locally
go build -o sandboxed .

# Run with development flags
./sandboxed mcp --sse --port 8080
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

