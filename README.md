# Sandboxed

Create sanboxed environment for running generated code.
## Getting Started

### Prerequisites
- Go 1.23 or later

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
go run main.go exec "echo Hello" 
go run main.go exec "echo $MY_VAR"
```

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
- **server**: Run as a REST server
- **mcp**: Run as mcp server


## Using the SDK

```
package main

import (
	"log"

	"github.com/altgen-ai/sandboxed/pkg/sdk"
)

func main() {

	sandbox, err := sdk.CreateSandbox("debug-generated-code", "python")
	if err != nil {
		log.Fatalf("failed to create sandbox: %v", err)
	}

	defer sandbox.Destroy()

	code := `python -c 'print("Hello, World!")'`

	output, err := sandbox.Run(code)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Output: %s", output.Result)

	code = `python --version`
	output, err = sandbox.Run(code)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Output: %s", output.Result)
}

```

## Using the REST APIs


## Using the MCP server

