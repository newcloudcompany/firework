package connect

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk/vsock"
	"github.com/jlkiri/firework/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewConnectCommand() *cobra.Command {
	connectCmd := &cobra.Command{
		Use:   "connect <name>",
		Short: "Connect to a VM",
		Long:  `Connect to a VM`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vmName := args[0]
			return runConnect(vmName)
		},
	}

	return connectCmd
}

func runConnect(vmName string) error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to make terminal raw: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	socket := config.VsockPath(vmName)
	conn, err := vsock.DialContext(context.Background(), socket, config.VSOCK_LISTENER_PORT, vsock.WithDialTimeout(time.Second*5))
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", socket, err)
	}

	go func() { _, _ = io.Copy(conn, os.Stdin) }()
	_, err = io.Copy(os.Stdout, conn)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	return nil
}
