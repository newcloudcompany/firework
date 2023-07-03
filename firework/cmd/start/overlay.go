package start

import (
	"os/exec"

	"github.com/jlkiri/firework/internal/config"
)

func createOverlayDrive(vmId string) (string, error) {
	path := config.OverlayDrivePath(vmId)

	// Create 2GB sparse file and format it as ext4
	_, err := exec.Command("truncate", "-s", "2G", path).CombinedOutput()
	if err != nil {
		return "", err
	}

	_, err = exec.Command("mkfs.ext4", path).CombinedOutput()
	if err != nil {
		return "", err
	}

	return path, nil
}
