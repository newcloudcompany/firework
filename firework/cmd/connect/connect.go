package connect

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/jlkiri/firework/sources"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var vmDataDir = filepath.Join(sources.DataDir, "vm")

func NewConnectCommand() *cobra.Command {
	connectCmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to a VM",
		Long:  `Connect to a VM`,
		RunE: func(cmd *cobra.Command, args []string) error {
			vmName := args[0]
			return runConnect(vmName)
		},
	}

	return connectCmd
}

func cleanup() {
}

func runConnect(vmName string) error {
	defer func() {
		if err := recover(); err != nil {
			cleanup()
		}
	}()

	defer cleanup()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to make terminal raw: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	socket := filepath.Join(vmDataDir, fmt.Sprintf("%s-v.sock", vmName))
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", socket, err)
	}

	if _, err := io.WriteString(conn, "CONNECT 10000\n"); err != nil {
		return fmt.Errorf("failed to write CONNECT command: %w", err)
	}

	go func() { _, _ = io.Copy(conn, os.Stdin) }()
	_, err = io.Copy(os.Stdout, conn)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	return nil
}
