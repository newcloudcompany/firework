package connect

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk/vsock"
	"github.com/jlkiri/firework/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func NewConnectCommand() *cobra.Command {
	connectCmd := &cobra.Command{
		Use:   "connect <name>",
		Short: "Connect to a VM",
		Long:  `Connect to a VM`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vmName := args[0]
			return runConnect(vmName)
		},
	}

	return connectCmd
}

func runConnect(vmName string) error {
	socket := config.VsockPath(vmName)
	conn, err := vsock.DialContext(context.Background(), socket, config.VSOCK_LISTENER_PORT, vsock.WithDialTimeout(time.Second*5))
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", socket, err)
	}

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)

	syncResize(conn, ch)

	ch <- syscall.SIGWINCH                        // Initial resize.
	defer func() { signal.Stop(ch); close(ch) }() // Cleanup signals when done.

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to make terminal raw: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() { _, _ = io.Copy(conn, os.Stdin) }()
	_, err = io.Copy(os.Stdout, conn)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	return nil
}

func syncResize(conn io.Writer, ch chan os.Signal) {
	go func() {
		for range ch {
			width, height, err := getTerminalSize(int(os.Stdin.Fd()))
			if err != nil {
				fmt.Printf("failed to get terminal size: %s\n", err)
				continue
			}

			io.WriteString(conn, fmt.Sprintf("RESIZE,%d,%d,\n", width, height))
		}
	}()
}

func getTerminalSize(fd int) (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}

	return int(ws.Col), int(ws.Row), nil
}
