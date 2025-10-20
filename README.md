# Sandboxed Go CLI Application

A command-line application built with Go and Cobra CLI framework.

## Getting Started

### Prerequisites
- Go 1.19 or later

### Running the Application
```bash
# Run the main command
go run main.go

# Show help
go run main.go --help

# Greet someone
go run main.go greet Alice

# Greet someone in uppercase
go run main.go greet Bob --uppercase

# Show version
go run main.go version

# Execute shell commands
go run main.go exec "ls -la"
go run main.go exec "echo Hello" --dir /tmp
go run main.go exec "echo $MY_VAR" --env MY_VAR=value

# Open files/directories in code editor
go run main.go code                    # Open current directory in VS Code
go run main.go code main.go           # Open specific file
go run main.go code --editor vim .    # Open with vim

# Kubernetes operations
go run main.go k8s create-pod my-pod --image nginx
go run main.go k8s create-pod test-pod --image busybox --command "echo" --args "Hello"
go run main.go k8s list-pods
go run main.go k8s get-pod my-pod --logs
go run main.go k8s delete-pod my-pod
```

### Building the Application
```bash
go build -o sandboxed
./sandboxed --help
```

## Available Commands

- **Root command**: Basic welcome message
- **greet [name]**: Greet someone by name (optional --uppercase flag)
- **version**: Display application version
- **exec [command]**: Execute shell commands with optional directory and environment settings
- **code [path]**: Open files or directories in a code editor (VS Code by default)
- **k8s**: Kubernetes operations (create-pod, delete-pod, list-pods, get-pod)
- **help**: Show help for any command

## Project Structure
```
.
├── main.go          # Main application entry point
├── go.mod           # Go module file with dependencies
├── cmd/             # Cobra command definitions
│   ├── root.go      # Root command and CLI setup
│   ├── greet.go     # Greet command implementation
│   ├── version.go   # Version command implementation
│   ├── exec.go      # Shell command execution
│   ├── code.go      # Code editor integration
│   └── k8s.go       # Kubernetes operations
├── pkg/             # Internal packages
│   └── k8sclient/   # Kubernetes client wrapper
│       └── client.go
├── Makefile         # Build and development tasks
├── README.md        # This file
└── .gitignore       # Git ignore file
```

## Dependencies

- [Cobra](https://github.com/spf13/cobra) - A library for creating powerful modern CLI applications
- [Kubernetes client-go](https://github.com/kubernetes/client-go) - Go client for Kubernetes API

## Prerequisites

For Kubernetes functionality:
- Access to a Kubernetes cluster
- Valid kubeconfig file (usually at `~/.kube/config`)
- Appropriate RBAC permissions for pod operations

## Kubernetes Features

The CLI provides comprehensive Kubernetes pod management:

- **Create pods** with custom images, commands, and labels
- **Delete pods** by name
- **List pods** in any namespace with status information
- **Get pod details** including logs
- **Wait for pod readiness** with configurable timeout
- **Namespace support** for all operations