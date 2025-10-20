package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec [command]",
	Short: "Execute shell commands",
	Long: `Execute shell commands with optional directory and environment variable settings.
	
Examples:
  sandboxed exec "ls -la"
  sandboxed exec "echo Hello World" --dir /tmp
  sandboxed exec "echo $MY_VAR" --env MY_VAR=value
  sandboxed exec -f script.sh
  sandboxed exec -f ../commands.txt --dir /tmp`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		dir, _ := cmd.Flags().GetString("dir")
		envVars, _ := cmd.Flags().GetStringSlice("env")
		shell, _ := cmd.Flags().GetString("shell")
		file, _ := cmd.Flags().GetString("file")
		
		var commands []string
		
		// If file flag is provided, read commands from file
		if file != "" {
			// Convert to absolute path
			absPath, err := filepath.Abs(file)
			if err != nil {
				fmt.Printf("Error resolving file path: %v\n", err)
				os.Exit(1)
			}
			
			// Check if file exists
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				fmt.Printf("File does not exist: %s\n", absPath)
				os.Exit(1)
			}
			
			// Read commands from file
			fileHandle, err := os.Open(absPath)
			if err != nil {
				fmt.Printf("Error opening file: %v\n", err)
				os.Exit(1)
			}
			defer fileHandle.Close()
			
			scanner := bufio.NewScanner(fileHandle)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				// Skip empty lines and comments
				if line != "" && !strings.HasPrefix(line, "#") {
					commands = append(commands, line)
				}
			}
			
			if err := scanner.Err(); err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				os.Exit(1)
			}
			
			if len(commands) == 0 {
				fmt.Println("No valid commands found in file")
				os.Exit(1)
			}
		} else {
			// Use command line arguments
			if len(args) == 0 {
				fmt.Println("Error: either provide a command or use -f to specify a file")
				os.Exit(1)
			}
			commands = []string{strings.Join(args, " ")}
		}
		
		// Execute each command
		for i, command := range commands {
			if len(commands) > 1 {
				fmt.Printf("\n=== Executing command %d/%d ===\n", i+1, len(commands))
			}
			
			// Create the command
			var execCmd *exec.Cmd
			if shell != "" {
				execCmd = exec.Command(shell, "-c", command)
			} else {
				// Default to sh on Unix systems
				execCmd = exec.Command("sh", "-c", command)
			}
			
			// Set working directory if specified
			if dir != "" {
				execCmd.Dir = dir
			}
			
			// Set environment variables
			execCmd.Env = os.Environ()
			for _, env := range envVars {
				execCmd.Env = append(execCmd.Env, env)
			}
			
			// Set up stdout and stderr
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			
			// Execute the command
			fmt.Printf("Executing: %s\n", command)
			if dir != "" {
				fmt.Printf("In directory: %s\n", dir)
			}
			fmt.Println("---")
			
			err := execCmd.Run()
			if err != nil {
				fmt.Printf("Error executing command: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	
	// Add flags
	execCmd.Flags().StringP("file", "f", "", "Execute commands from a file")
	execCmd.Flags().StringP("dir", "d", "", "Working directory for the command")
	execCmd.Flags().StringSliceP("env", "e", []string{}, "Environment variables (format: KEY=value)")
	execCmd.Flags().StringP("shell", "s", "", "Shell to use (default: sh)")
}