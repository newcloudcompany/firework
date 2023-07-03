package start

import (
	"context"
	"os"
	"path/filepath"

	"github.com/jlkiri/firework/sources"
	"golang.org/x/exp/slog"
)

func prepareEnvironment() error {
	// Create firework data directory and subdirectories
	if err := os.MkdirAll(sources.DataDir, 0755); err != nil {
		return err
	}

	if err := os.RemoveAll(sources.VmDataDir); err != nil {
		return err
	}

	if err := os.MkdirAll(sources.KernelDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(sources.RootFsDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(sources.MiscDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(sources.VmDataDir, 0755); err != nil {
		return err
	}

	ctx := context.TODO()
	err := ensureKernel(ctx, sources.KernelUrl, filepath.Join(sources.KernelDir, "vmlinux"))
	if err != nil {
		return err
	}

	err = ensureSquashFs(ctx, sources.SquashFsUrl, sources.RootFsDir)
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

func ensureSquashFs(ctx context.Context, rootFsUrl, rootFsDir string) error {
	rootFsPath := filepath.Join(rootFsDir, "rootfs.squashfs")
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

func getRootFsPath() string {
	envRootFsPath := os.Getenv("ROOTFS_PATH")
	if envRootFsPath != "" {
		return envRootFsPath
	}
	return filepath.Join(sources.RootFsDir, "rootfs.squashfs")
}

func getKernelPath() string {
	return filepath.Join(sources.KernelDir, "vmlinux")
}

func getSocketPath(vmId string) string {
	return filepath.Join(sources.VmDataDir, vmId+".sock")
}

func getLogFifoPath(vmId string) string {
	return filepath.Join(sources.MiscDir, "log-"+vmId+".fifo")
}

func getMetricsFifoPath(vmId string) string {
	return filepath.Join(sources.MiscDir, "metrics-"+vmId+".fifo")
}

func getVsockPath(name string) string {
	return filepath.Join(sources.VmDataDir, name+"-v.sock")
}
