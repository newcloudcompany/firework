package stop

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/jlkiri/firework/internal/config"
	"github.com/jlkiri/firework/internal/network"
	"github.com/jlkiri/firework/internal/vm"
	"github.com/spf13/cobra"
)

func NewStopCommand() *cobra.Command {
	cleanupCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a VM cluster from config",
		Long:  `Stop a VM cluster from config`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop()
		},
	}

	return cleanupCmd
}

func cleanup() {
	conf, err := config.Read("config.json")
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	if err := network.Cleanup(conf.SubnetCidr); err != nil {
		log.Fatalf("Failed to cleanup network: %v", err)
	}

	if err := os.Remove(filepath.Join(config.MiscDir, "ips.db")); err != nil {
		log.Println("Failed to remove ips.db:", err)
	}

	if err := os.RemoveAll(config.VmDataDir); err != nil {
		log.Println("Failed to remove vm data dir:", err)
	}
}

func runStop() error {
	defer cleanup()

	pidTablePath := config.PidTablePath()
	pidTableFile, err := os.ReadFile(pidTablePath)
	if err != nil {
		return err
	}

	var pidTable vm.PidTable
	if err := json.Unmarshal(pidTableFile, &pidTable); err != nil {
		return err
	}

	for _, entry := range pidTable {
		socketPath := config.SocketPath(entry.VmId)
		if _, err := os.Stat(socketPath); os.IsNotExist(err) {
			continue
		}

		m, err := firecracker.NewMachine(context.TODO(), firecracker.Config{
			SocketPath: socketPath,
		})
		if err != nil {
			return err
		}

		if err := m.Shutdown(context.TODO()); err != nil {
			return err
		}
	}

	return nil
}
