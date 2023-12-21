package start

import (
	"os"
	"os/exec"

	units "github.com/docker/go-units"
	"github.com/jlkiri/firework/internal/config"
)

func createOverlayDrive(vmId string, capacity int64) (string, error) {
	path := config.OverlayDrivePath(vmId)

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	defer f.Close()

	if err := f.Truncate(units.GiB * capacity); err != nil {
		return "", err
	}

	_, err = exec.Command("mkfs.ext4", path).CombinedOutput()
	if err != nil {
		return "", err
	}

	return path, nil
}
