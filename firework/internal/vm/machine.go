package vm

import (
	"context"
	"net"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

type MachineOptions struct {
	KernelImagePath  string
	RootFsPath       string
	SocketPath       string
	LogFifoPath      string
	MetricsFifoPath  string
	VsockPath        string
	InitrdPath       string
	OverlayDrivePath string
	Id               string
	Cid              uint32
	Memory           int64
	Vcpu             int64
	IpConfig         *machineIpConfig
}

type machineIpConfig struct {
	GatewayIp net.IP
	IpAddr    net.IPNet // The IP field of IPNet must be an actual IP and not the network number
	TapDevice string
}

func CreateMachine(ctx context.Context, opts MachineOptions) (*firecracker.Machine, error) {
	mac, err := generateMACAddress()
	if err != nil {
		return nil, err
	}

	networkInterface := firecracker.NetworkInterface{
		StaticConfiguration: &firecracker.StaticNetworkConfiguration{
			HostDevName: opts.IpConfig.TapDevice,
			MacAddress:  mac,
			IPConfiguration: &firecracker.IPConfiguration{
				IfName:      "eth0",
				IPAddr:      opts.IpConfig.IpAddr,
				Gateway:     opts.IpConfig.GatewayIp,
				Nameservers: []string{"8.8.8.8"},
			},
		},
		AllowMMDS: true,
	}

	cfg := firecracker.Config{
		SocketPath:      opts.SocketPath,
		KernelImagePath: opts.KernelImagePath,
		KernelArgs:      "console=ttyS0 noapic reboot=k panic=1 pci=off overlay_root=vdb i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd init=/sbin/overlay-init",
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(opts.Vcpu),
			MemSizeMib: firecracker.Int64(opts.Memory),
			Smt:        firecracker.Bool(false),
		},
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("root"),
				IsRootDevice: firecracker.Bool(true),
				IsReadOnly:   firecracker.Bool(false),
				PathOnHost:   firecracker.String(opts.RootFsPath),
			},
			{
				DriveID:      firecracker.String("overlayfs"),
				IsRootDevice: firecracker.Bool(false),
				IsReadOnly:   firecracker.Bool(false),
				PathOnHost:   firecracker.String(opts.OverlayDrivePath),
			},
		},
		FifoLogWriter: os.Stdout,
		LogFifo:       opts.LogFifoPath,
		MetricsFifo:   opts.MetricsFifoPath,
		LogLevel:      "Debug",
		VsockDevices: []firecracker.VsockDevice{
			{
				CID:  opts.Cid,
				Path: opts.VsockPath,
			},
		},
		VMID:              opts.Id,
		MmdsVersion:       firecracker.MMDSv2,
		ForwardSignals:    []os.Signal{},
		NetworkInterfaces: []firecracker.NetworkInterface{networkInterface},
		// InitrdPath:        opts.InitrdPath,
	}

	machine, err := createFirecrackerVM(ctx, cfg, "/bin/firecracker", opts.SocketPath)
	if err != nil {
		return nil, err
	}

	return machine, nil
}

func NewMachineIpConfig(gatewayIp net.IP, ipAddr string, tapDevice string) (*machineIpConfig, error) {
	ip, ipnet, err := net.ParseCIDR(ipAddr)
	if err != nil {
		return nil, err
	}

	return &machineIpConfig{
		GatewayIp: gatewayIp,
		IpAddr: net.IPNet{
			IP:   ip,
			Mask: ipnet.Mask,
		},
		TapDevice: tapDevice,
	}, nil
}
