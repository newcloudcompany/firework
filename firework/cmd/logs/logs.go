package logs

import (
	"github.com/spf13/cobra"
)

func NewLogsCommand() *cobra.Command {
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "View logs of a running VM",
		Long:  `View logs of a running VM`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogs()
		},
	}

	return logsCmd
}

func runLogs() error {
	// logsFifoPath := config.LogsFifoPath()

	return nil
}
