package vm

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

type MachineOptions struct {
	KernelImagePath string
	RootFsPath      string
	SocketPath      string
	FifoPath        string
	VsockPath       string
	Id              string
	Cid             uint32
	IpConfig        *machineIpConfig
}

type machineIpConfig struct {
	GatewayIp net.IP
	IpAddr    net.IPNet // The IP field of IPNet must be an actual IP and not the network number
	TapDevice string
}

func generateMACAddress() (string, error) {
	// MAC address is 6 bytes long
	mac := make([]byte, 6)

	// Read 6 random bytes
	_, err := rand.Read(mac)
	if err != nil {
		return "", err
	}

	// Set the locally administered bit, and clear the multicast bit to make it a valid MAC address
	// Locally administered bit is the second least significant bit of the first byte
	// Multicast bit is the least significant bit of the first byte
	mac[0] = (mac[0] & 0xfe) | 0x02

	// Format the MAC address
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]), nil
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
		KernelArgs:      "console=ttyS0 noapic reboot=k panic=1 pci=off i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd",
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(2),
			MemSizeMib: firecracker.Int64(1024),
			Smt:        firecracker.Bool(false),
		},
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("root"),
				IsRootDevice: firecracker.Bool(true),
				IsReadOnly:   firecracker.Bool(false),
				PathOnHost:   firecracker.String(opts.RootFsPath),
			},
		},
		FifoLogWriter: os.Stdout,
		LogFifo:       opts.FifoPath,
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
