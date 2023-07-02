package start

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/jlkiri/firework/sources"
)

func createOverlayDrive(vmId string) (string, error) {
	path := filepath.Join(sources.VmDataDir, fmt.Sprintf("%s-overlay.ext4", vmId))

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
