package vm

import (
	"context"
	"os"
	"syscall"

	"github.com/firecracker-microvm/firecracker-go-sdk"
)

func createFirecrackerVM(ctx context.Context, cfg firecracker.Config, binPath, socketPath string) (*firecracker.Machine, error) {
	// Command interface automatically connects stdout/stderr/stdin to /dev/null if not specified.
	cmd := firecracker.VMCommandBuilder{}.
		WithSocketPath(socketPath).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithBin(binPath).
		Build(ctx)

	// Copy parent environment variables to the child process.
	cmd.Env = os.Environ()

	// Detach from controlling terminal so that the signals originating from the terminal
	// do not get independently sent to the child process.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	m, err := firecracker.NewMachine(ctx, cfg, firecracker.WithProcessRunner(cmd))
	if err != nil {
		return nil, err
	}

	return m, nil
}
