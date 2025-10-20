package cmd

import (
	"log"

	"github.com/altgen-ai/sandboxed/pkg/mcp"
	"github.com/spf13/cobra"
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

The server runs on stdio transport and communicates via JSON-RPC 2.0 protocol.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create and start MCP server
		server := mcp.NewServer()
		
		if err := mcp.RunServer(server); err != nil {
			log.Fatalf("MCP server failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)


	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mcpCmd.PersistentFlags().String("foo", "", "A help for foo")
	
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mcpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}