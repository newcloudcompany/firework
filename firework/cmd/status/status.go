package status

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/jlkiri/firework/internal/config"
	"github.com/jlkiri/firework/internal/vm"
	"github.com/spf13/cobra"
)

func NewStatusCommand() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "View status of running VMs",
		Long:  `View status of running VMs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}

	return statusCmd
}

func runStatus() error {
	pidTablePath := config.PidTablePath()
	pidTableFile, err := os.ReadFile(pidTablePath)
	if err != nil {
		return err
	}

	var pidTable vm.PidTable
	if err := json.Unmarshal(pidTableFile, &pidTable); err != nil {
		return err
	}

	ctx := context.Background()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "VMID\tNAME\tSTATUS")

	var vmData [][]string

	for name, entry := range pidTable {
		socketPath := config.SocketPath(entry.VmId)
		if _, err := os.Stat(socketPath); os.IsNotExist(err) {
			continue
		}

		m, err := firecracker.NewMachine(ctx, firecracker.Config{
			SocketPath: socketPath,
		})
		if err != nil {
			return err
		}

		instance, err := m.DescribeInstanceInfo(ctx)
		if err != nil {
			return err
		}

		vmData = append(vmData, []string{entry.VmId, name, *instance.State})
	}

	// Write the data to the tabwriter.Writer
	for _, data := range vmData {
		fmt.Fprintf(w, "%s\t%s\t%s\n", data[0], data[1], data[2])
	}

	// Flush the writer to render the output
	w.Flush()
	return nil
}
