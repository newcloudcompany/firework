package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Node struct {
	Name       string `json:"name"`
	Vcpu       int64  `json:"vcpu"`
	Memory     int64  `json:"memory"`
	RootFsPath string `json:"rootfs_path"`
}

type Config struct {
	Nodes      []Node `json:"nodes"`
	SubnetCidr string `json:"subnet_cidr"`
	Gateway    string `json:"gateway"`
}

func Read(path string) (Config, error) {
	wd, _ := os.Getwd()
	absPath := filepath.Join(wd, path)
	file, err := os.ReadFile(absPath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
