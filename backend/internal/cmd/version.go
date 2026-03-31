package cmd

import (
	"fmt"

	"zipic/internal/version"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Display the version, build date, and git commit of Zipic.`,
	Run: func(cmd *cobra.Command, args []string) {
		info := version.Info()
		fmt.Printf("Zipic %s\n", info["version"])
		fmt.Printf("Build Date: %s\n", info["build_date"])
		fmt.Printf("Git Commit: %s\n", info["git_commit"])
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}