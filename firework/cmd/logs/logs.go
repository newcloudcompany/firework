package logs

import (
	"bufio"
	"os"

	"github.com/jlkiri/firework/internal/config"
	"github.com/spf13/cobra"
)

func NewLogsCommand() *cobra.Command {
	logsCmd := &cobra.Command{
		Use:   "logs <VMID>",
		Short: "View VMM logs or logs of a running VM",
		Long:  `View VMM logs or logs of a running VM`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return runVmmLogs()
			}
			return runVmLogs(args[0])
		},
	}

	return logsCmd
}

func runVmmLogs() error {
	return followLogs(config.VmmLogPath)
}

func runVmLogs(vmId string) error {
	return followLogs(config.LogFifoPath(vmId))
}

func followLogs(path string) error {
	logFile, err := os.Open(path)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(logFile)
	for scanner.Scan() {
		println(scanner.Text())
	}

	return nil
}
