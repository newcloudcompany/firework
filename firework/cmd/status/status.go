package status

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/jlkiri/firework/internal/config"
	"github.com/jlkiri/firework/internal/vm"
	"github.com/sirupsen/logrus"
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

type Table struct {
	header []string
	rows   [][]string
}

func (t *Table) SetHeader(header []string) {
	t.header = header
}

func (t *Table) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

func (t *Table) Print() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, strings.Join(t.header, "\t"))
	fmt.Fprintln(w, "\t\t\t")

	for _, row := range t.rows {
		str := strings.Join(row, "\t")
		fmt.Fprintln(w, str)
	}

	w.Flush()
}

func runStatus() error {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Logger that logs to /dev/null to hide Firecracker binary output
	logrus.SetOutput(devNull)

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

	table := &Table{}
	table.SetHeader([]string{"VMID", "NAME", "STATUS"})

	for name, entry := range pidTable {
		socketPath := config.SocketPath(entry.VmId)
		if _, err := os.Stat(socketPath); os.IsNotExist(err) {
			continue
		}

		m, err := firecracker.NewMachine(ctx, firecracker.Config{
			SocketPath: socketPath,
		}, firecracker.WithLogger(logrus.NewEntry(logrus.StandardLogger())))
		if err != nil {
			return err
		}

		instance, err := m.DescribeInstanceInfo(ctx)
		if err != nil {
			return err
		}

		table.AddRow([]string{entry.VmId, name, *instance.State})
	}

	table.Print()
	return nil
}
