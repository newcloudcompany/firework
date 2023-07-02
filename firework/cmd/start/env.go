package start

import (
	"fmt"
	"io"
	"net/http"
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

	err := ensureKernel(sources.KernelUrl, filepath.Join(sources.KernelDir, "vmlinux"))
	if err != nil {
		return err
	}

	err = ensureSquashFs(sources.SquashFsUrl, sources.RootFsDir)
	if err != nil {
		return err
	}

	return nil
}

func ensureKernel(kernelUrl, kernelPath string) error {
	if _, err := os.Stat(kernelPath); os.IsNotExist(err) {
		err := download(kernelUrl, kernelPath)
		if err != nil {
			return err
		}

		err = os.Chmod(kernelPath, 0755)
		if err != nil {
			return err
		}
		slog.Info("Downloaded kernel.")
	}

	return nil
}

func ensureSquashFs(rootFsUrl, rootFsDir string) error {
	if _, err := os.Stat(filepath.Join(rootFsDir, "rootfs.squashfs")); os.IsNotExist(err) {

		err := download(rootFsUrl, filepath.Join(rootFsDir, "rootfs.squashfs"))
		if err != nil {
			return err
		}
		slog.Info("Downloaded rootfs squashfs image.")
	}

	return nil
}

func download(url string, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
