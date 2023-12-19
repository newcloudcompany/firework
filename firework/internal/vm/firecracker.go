package vm

import (
	"context"
	"io"
	"os"
	"syscall"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/sirupsen/logrus"
)

func openFileLogger(vmmLogPath string) (*logrus.Entry, error) {
	logFile, err := os.OpenFile(vmmLogPath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger := logrus.New()
	logger.SetOutput(logFile)
	logger.SetLevel(logrus.DebugLevel)

	return logrus.NewEntry(logger), nil
}

func createFirecrackerVM(ctx context.Context, cfg firecracker.Config, stdio io.Writer, binPath, vmmLogPath, socketPath string) (*firecracker.Machine, error) {
	// Command interface automatically connects stdout/stderr/stdin to /dev/null if not specified.
	cmd := firecracker.VMCommandBuilder{}.
		WithSocketPath(socketPath).
		WithBin(binPath).
		WithStdout(stdio).
		WithStderr(stdio).
		WithStdin(os.Stdin).
		Build(ctx)

	// Copy parent environment variables to the child process.
	cmd.Env = os.Environ()

	// Detach from controlling terminal so that the signals originating from the terminal
	// do not get independently sent to the child process.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	logger, err := openFileLogger(vmmLogPath)
	if err != nil {
		return nil, err
	}

	m, err := firecracker.NewMachine(ctx, cfg,
		firecracker.WithProcessRunner(cmd),
		firecracker.WithLogger(logger),
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}
