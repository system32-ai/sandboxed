package cmd

import (
	"log"

	"github.com/altgen-ai/sandboxed/pkg/mcp"
	"github.com/spf13/cobra"
)

var (
	sseMode bool
	ssePort int
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server",
	Long: `Start an MCP server that provides sandbox management tools including:
- create_sandbox: Create a new sandbox environment for code execution
- run_code: Execute code in an existing sandbox environment  
- destroy_sandbox: Destroy a sandbox and clean up resources
- list_sandboxes: List all active sandbox environments

The server can run in two modes:
1. stdio transport (default) - for direct MCP client integration
2. SSE (Server-Sent Events) mode - for web-based clients via HTTP

Examples:
  # Start in stdio mode (default)
  sandboxed mcp

  # Start in SSE mode on port 8080
  sandboxed mcp --sse

  # Start in SSE mode on custom port
  sandboxed mcp --sse --port 9000`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create MCP server
		server := mcp.NewServer()
		
		if sseMode {
			// Start SSE server
			log.Printf("Starting MCP server in SSE mode on port %d", ssePort)
			if err := mcp.RunServerSSE(server, ssePort); err != nil {
				log.Fatalf("MCP SSE server failed: %v", err)
			}
		} else {
			// Start stdio server
			log.Println("Starting MCP server in stdio mode")
			if err := mcp.RunServer(server); err != nil {
				log.Fatalf("MCP server failed: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	// Add flags for SSE mode
	mcpCmd.Flags().BoolVar(&sseMode, "sse", false, "Start server in SSE (Server-Sent Events) mode for web clients")
	mcpCmd.Flags().IntVar(&ssePort, "port", 8080, "Port to listen on when in SSE mode")
}