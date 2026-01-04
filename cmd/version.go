package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of sandboxed",
	Long:  `All software has versions. This is sandboxed's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sandboxed v1.0.8")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
