package stop

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/jlkiri/firework/internal/network"
	"github.com/jlkiri/firework/internal/vm"
	"github.com/jlkiri/firework/sources"
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
	if err := network.Cleanup(); err != nil {
		log.Fatalf("Failed to cleanup network: %v", err)
	}

	if err := os.Remove(filepath.Join(sources.MiscDir, "ips.db")); err != nil {
		log.Println("Failed to remove ips.db:", err)
	}

	if err := os.RemoveAll(sources.VmDataDir); err != nil {
		log.Println("Failed to remove vm data dir:", err)
	}
}

func runStop() error {
	defer cleanup()

	pidTablePath := filepath.Join(sources.MiscDir, "pid_table.json")
	pidTableFile, err := os.ReadFile(pidTablePath)
	if err != nil {
		return err
	}

	var pidTable vm.PidTable
	if err := json.Unmarshal(pidTableFile, &pidTable); err != nil {
		return err
	}

	for _, entry := range pidTable {
		proc, err := os.FindProcess(entry.Pid)
		if err != nil {
			return err
		}

		// if err := proc.Signal(syscall.SIGTERM); err != nil {
		// 	return err
		// }

		socketPath := filepath.Join(sources.VmDataDir, fmt.Sprintf("%s.sock", entry.VmId))
		cmd := firecracker.VMCommandBuilder{}.
			Build(context.TODO())

		cmd.Process = proc

		m, err := firecracker.NewMachine(context.TODO(), firecracker.Config{
			SocketPath: socketPath,
		}, firecracker.WithProcessRunner(cmd))
		if err != nil {
			return err
		}

		if err := m.Shutdown(context.TODO()); err != nil {
			return err
		}
	}

	return nil
}