# Homebrew Tap for Sandboxed

This is the official Homebrew tap for [Sandboxed](https://github.com/system32-ai/sandboxed).

## Installation

```bash
# Add the tap
brew tap system32-ai/sandboxed

# Install sandboxed
brew install sandboxed
```

## Usage

```bash
# Show version
sandboxed version

# Start REST API server
sandboxed server

# Start MCP server for AI integration
sandboxed mcp

# Get help
sandboxed --help
```

## Requirements

Sandboxed requires access to a Kubernetes cluster. Make sure you have:

1. kubectl installed and configured
2. Proper RBAC permissions for pod management
3. Access to a Kubernetes cluster

For detailed setup instructions, see the [main repository](https://github.com/system32-ai/sandboxed).

## Updating

```bash
# Update the tap
brew update

# Upgrade sandboxed
brew upgrade sandboxed
```
