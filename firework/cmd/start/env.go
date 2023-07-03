package start

import (
	"context"
	"os"

	"github.com/jlkiri/firework/internal/config"
	"golang.org/x/exp/slog"
)

func prepareEnvironment() error {
	// Create firework data directory and subdirectories
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return err
	}

	if err := os.RemoveAll(config.VmDataDir); err != nil {
		return err
	}

	if err := os.MkdirAll(config.KernelDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(config.RootFsDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(config.MiscDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(config.VmDataDir, 0755); err != nil {
		return err
	}

	ctx := context.TODO()
	err := ensureKernel(ctx, config.KernelUrl, config.KernelPath())
	if err != nil {
		return err
	}

	err = ensureSquashFs(ctx, config.SquashFsUrl, config.RootFsPath())
	if err != nil {
		return err
	}

	return nil
}

func ensureKernel(ctx context.Context, kernelUrl, kernelPath string) error {
	if _, err := os.Stat(kernelPath); os.IsNotExist(err) {
		f, err := os.Create(kernelPath)
		if err != nil {
			return err
		}

		w := &progressWriter{
			inner:   f,
			written: 0,
			total:   0,
		}

		slog.Info("Downloading kernel...")
		err = download(ctx, kernelUrl, w)
		if err != nil {
			return err
		}

		err = os.Chmod(kernelPath, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func ensureSquashFs(ctx context.Context, rootFsUrl, rootFsPath string) error {
	if _, err := os.Stat(rootFsPath); os.IsNotExist(err) {
		f, err := os.Create(rootFsPath)
		if err != nil {
			return err
		}

		w := &progressWriter{
			inner:   f,
			written: 0,
			total:   0,
		}

		slog.Info("Downloading squashfs rootfs image...")
		err = download(ctx, rootFsUrl, w)
		if err != nil {
			return err
		}
	}

	return nil
}
