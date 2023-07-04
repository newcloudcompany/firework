package config

import (
	"os"
	"path/filepath"
)

const KernelUrl = "https://pub-1a5aeef625fc45b4a4bef89ee141047f.r2.dev/vmlinux"
const SquashFsUrl = "https://pub-1a5aeef625fc45b4a4bef89ee141047f.r2.dev/rootfs.squashfs"

const DataDir = "/var/lib/firework"
const CacheDir = "/var/lib/firework/cache"
const VmDataDir = "/var/lib/firework/vm"

const KernelDir = "/var/lib/firework/cache/kernel"
const RootFsDir = "/var/lib/firework/cache/rootfs"
const MiscDir = "/var/lib/firework/misc"
const DbPath = "/var/lib/firework/misc/ips.db"
const VmmLogPath = "/var/lib/firework/vmm.log"

func RootFsPath() string {
	envRootFsPath := os.Getenv("ROOTFS_PATH")
	if envRootFsPath != "" {
		return envRootFsPath
	}
	return filepath.Join(RootFsDir, "rootfs.squashfs")
}

func KernelPath() string {
	return filepath.Join(KernelDir, "vmlinux")
}

func SocketPath(vmId string) string {
	return filepath.Join(VmDataDir, vmId+".sock")
}

func LogFifoPath(vmId string) string {
	return filepath.Join(MiscDir, "log-"+vmId+".fifo")
}

func MetricsFifoPath(vmId string) string {
	return filepath.Join(MiscDir, "metrics-"+vmId+".fifo")
}

func VsockPath(name string) string {
	return filepath.Join(VmDataDir, name+"-v.sock")
}

func OverlayDrivePath(vmId string) string {
	return filepath.Join(VmDataDir, vmId+"-overlay.ext4")
}

func PidTablePath() string {
	return filepath.Join(MiscDir, "pid_table.json")
}
