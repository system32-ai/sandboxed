package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// codeCmd represents the code command
var codeCmd = &cobra.Command{
	Use:   "code [path]",
	Short: "Open files or directories in a code editor",
	Long: `Open files or directories in a code editor (VS Code by default).
	
Examples:
  sandboxed code                    # Open current directory
  sandboxed code file.go           # Open specific file
  sandboxed code /path/to/project  # Open specific directory
  sandboxed code . --editor vim    # Open with different editor`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		editor, _ := cmd.Flags().GetString("editor")
		newWindow, _ := cmd.Flags().GetBool("new-window")
		wait, _ := cmd.Flags().GetBool("wait")
		
		// Determine the path to open
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		
		// Convert to absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			os.Exit(1)
		}
		
		// Check if path exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			fmt.Printf("Path does not exist: %s\n", absPath)
			os.Exit(1)
		}
		
		// Build command based on editor
		var execCmd *exec.Cmd
		switch editor {
		case "vscode", "code":
			args := []string{absPath}
			if newWindow {
				args = append([]string{"--new-window"}, args...)
			}
			if wait {
				args = append([]string{"--wait"}, args...)
			}
			execCmd = exec.Command("code", args...)
		case "vim":
			execCmd = exec.Command("vim", absPath)
		case "nano":
			execCmd = exec.Command("nano", absPath)
		case "emacs":
			execCmd = exec.Command("emacs", absPath)
		case "subl", "sublime":
			execCmd = exec.Command("subl", absPath)
		default:
			// Default to VS Code
			args := []string{absPath}
			if newWindow {
				args = append([]string{"--new-window"}, args...)
			}
			if wait {
				args = append([]string{"--wait"}, args...)
			}
			execCmd = exec.Command("code", args...)
		}
		
		// Execute the command
		fmt.Printf("Opening %s with %s\n", absPath, editor)
		
		// For terminal editors, we want to pass through stdin/stdout
		if editor == "vim" || editor == "nano" || editor == "emacs" {
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
		}
		
		err = execCmd.Run()
		if err != nil {
			fmt.Printf("Error opening editor: %v\n", err)
			fmt.Printf("Make sure %s is installed and available in your PATH\n", editor)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(codeCmd)
	
	// Add flags
	codeCmd.Flags().StringP("editor", "e", "code", "Editor to use (code, vim, nano, emacs, subl)")
	codeCmd.Flags().BoolP("new-window", "n", false, "Open in new window (VS Code only)")
	codeCmd.Flags().BoolP("wait", "w", false, "Wait for editor to close (VS Code only)")
}