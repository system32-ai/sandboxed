# Sandboxed

Create sanboxed environment for running generated code

## Getting Started

### Prerequisites
- Go 1.19 or later

### Running the Application
```bash
# Run the main command
go run main.go

# Show help
go run main.go --help

# Show version
go run main.go version

# Execute shell commands
go run main.go exec "ls -la"
go run main.go exec "echo Hello" --dir /tmp
go run main.go exec "echo $MY_VAR" --env MY_VAR=value

### Building the Application
```bash
go build -o sandboxed
./sandboxed --help
```

## Available Commands

- **version**: Display application version
- **exec [command]**: Execute shell commands with optional directory and environment settings
- **code [path]**: Open files or directories in a code editor (VS Code by default)
- **help**: Show help for any command


## Dependencies

- [Cobra](https://github.com/spf13/cobra) - A library for creating powerful modern CLI applications
- [Kubernetes client-go](https://github.com/kubernetes/client-go) - Go client for Kubernetes API

## Prerequisites

For Kubernetes functionality:
- Access to a Kubernetes cluster
- Valid kubeconfig file (usually at `~/.kube/config`)
- Appropriate RBAC permissions for pod operations
