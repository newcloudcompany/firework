package vm

import (
	"context"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
)

func createFirecrackerVM(ctx context.Context, cfg firecracker.Config, binPath, socketPath string) (*firecracker.Machine, error) {
	cmd := firecracker.VMCommandBuilder{}.
		WithSocketPath(socketPath).
		WithBin(binPath).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		Build(ctx)

	m, err := firecracker.NewMachine(ctx, cfg, firecracker.WithProcessRunner(cmd))
	if err != nil {
		return nil, err
	}

	// cmd.Stdout = os.Stdout

	return m, nil
}
