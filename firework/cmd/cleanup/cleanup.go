package cleanup

import (
	"log"
	"os"
	"path/filepath"

	"github.com/jlkiri/firework/internal/network"
	"github.com/jlkiri/firework/sources"
	"github.com/spf13/cobra"
)

var cacheDir = filepath.Join(sources.DataDir, "cache")
var kernelDir = filepath.Join(cacheDir, "kernel")
var rootFsDir = filepath.Join(cacheDir, "rootfs")
var vmDataDir = filepath.Join(sources.DataDir, "vm")
var miscDir = filepath.Join(sources.DataDir, "misc")
var dbPath = filepath.Join(miscDir, "ips.db")

func NewCleanupCommand() *cobra.Command {
	cleanupCmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup",
		Long:  `Cleanup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanup()
		},
	}

	return cleanupCmd
}

func cleanup() {
	if err := network.Cleanup(); err != nil {
		log.Fatalf("Failed to cleanup network: %v", err)
	}

	if err := os.Remove(filepath.Join(miscDir, "ips.db")); err != nil {
		log.Println("Failed to remove ips.db:", err)
	}

	if err := os.RemoveAll(filepath.Join(vmDataDir)); err != nil {
		log.Println("Failed to remove vm data dir:", err)
	}
}

func runCleanup() error {
	cleanup()
	return nil
}
