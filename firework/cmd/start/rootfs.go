package start

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func createRootFsCopy(rootFsPath, vmDataDirPath, id string) (string, error) {
	vmRootFsPath := filepath.Join(vmDataDirPath, fmt.Sprintf("%s.ext4", id))
	sourceFile, err := os.Open(rootFsPath)
	if err != nil {
		return "", err
	}
	defer sourceFile.Close()

	vmRootFsFile, err := os.Create(vmRootFsPath)
	if err != nil {
		return "", err
	}
	defer vmRootFsFile.Close()

	_, err = io.Copy(vmRootFsFile, sourceFile)
	if err != nil {
		return "", err
	}

	// Sync to ensure that the contents are written to disk
	err = vmRootFsFile.Sync()
	if err != nil {
		return "", err
	}

	return vmRootFsFile.Name(), nil
}
