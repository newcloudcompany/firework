package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "firework",
	Short: "A tool for launching local Firecracker VM clusters",
	Long:  `A tool for launching local Firecracker VM clusters`,
}

func Execute() {
	AddCommands(rootCmd)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
