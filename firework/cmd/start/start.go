package start

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"syscall"

	"github.com/google/uuid"
	"github.com/jlkiri/firework/internal/config"
	"github.com/jlkiri/firework/internal/ipam"
	"github.com/jlkiri/firework/internal/network"
	"github.com/jlkiri/firework/internal/vm"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

func NewStartCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start a VM cluster from config",
		Long:  `Start a VM cluster from config`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart()
		},
	}

	return startCmd
}

func runStart() error {
	defer cleanup()

	// TODO: Remove this
	os.Remove(config.DbPath)

	if err := prepareEnvironment(); err != nil {
		return err
	}
	slog.Debug("Prepared environment for execution.")

	conf, err := config.Read("config.json")
	if err != nil {
		return err
	}
	slog.Debug("Read config.json.", "config", conf)

	ipamDb, err := ipam.NewIPAM(config.DbPath, conf.SubnetCidr)
	if err != nil {
		return err
	}
	slog.Debug("Created and populated IPAM database.")

	bridge, err := network.NewBridgeNetwork(conf.SubnetCidr, conf.Gateway)
	if err != nil {
		return err
	}
	slog.Debug("Created a bridge network.", "cidr", conf.SubnetCidr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmmLogFifo, err := createVmmLogFifo(config.VmmLogPath)
	if err != nil {
		return fmt.Errorf("failed to create VMM log fifo: %w", err)
	}

	defer vmmLogFifo.Close()
	slog.Debug("Created VMM log fifo", "path", config.VmmLogPath)

	mg, err := createMachineGroup(ctx, conf.Nodes, bridge, ipamDb)
	if err != nil {
		return fmt.Errorf("failed to create machine group: %w", err)
	}
	slog.Debug("Created machine group from config:", "config", conf)

	if err := mg.Start(ctx); err != nil {
		return fmt.Errorf("failed to start machine group: %w", err)
	}

	slog.Debug("Installing SIGTERM and SIGINT signal handlers.")
	vm.InstallSignalHandlers(ctx, mg)

	if err := mg.Wait(ctx); err != nil {
		cancel() // Stop signal handlers
		return fmt.Errorf("an error occurred while waiting for the machine group to exit: %w", err)
	}

	slog.Info("Graceful shutdown successful.")
	return nil
}

func createMachineGroup(ctx context.Context, nodes []config.Node, bridge *network.BridgeNetwork, ipamDb *ipam.IPAM) (*vm.MachineGroup, error) {
	kernelPath := config.KernelPath()
	rootFsPath := config.RootFsPath()

	mg := vm.NewMachineGroup()

	for _, node := range nodes {
		cid := generateCid()
		id := uuid.NewString()

		slog.Info("Generated CID", "node", node.Name, "cid", cid)
		slog.Info("Generated ID", "node", node.Name, "id", id)

		tap, err := bridge.CreateTapDevice(id)
		if err != nil {
			return nil, err
		}

		addr, err := ipamDb.AllocateFreeIPAddress(id)
		if err != nil {
			return nil, err
		}
		slog.Info("Allocated free IP address", "node", node.Name, "addr", addr)

		socketPath := config.SocketPath(id)
		logFifoPath := config.LogFifoPath(id)
		metricsFifoPath := config.MetricsFifoPath(id)
		ipConfig, err := vm.NewMachineIpConfig(bridge.GetIPAddr(), addr, tap.Name)
		if err != nil {
			return nil, err
		}

		overlayDrivePath, err := createOverlayDrive(id)
		if err != nil {
			return nil, err
		}

		machine, err := vm.CreateMachine(ctx, vm.MachineOptions{
			Id:               id,
			RootFsPath:       rootFsPath,
			KernelImagePath:  kernelPath,
			SocketPath:       socketPath,
			LogFifoPath:      logFifoPath,
			MetricsFifoPath:  metricsFifoPath,
			OverlayDrivePath: overlayDrivePath,
			VmmLogPath:       config.VmmLogPath,
			VsockPath:        config.VsockPath(node.Name),
			Cid:              cid,
			Vcpu:             node.Vcpu,
			Memory:           node.Memory,
			IpConfig:         ipConfig,
		})
		if err != nil {
			return nil, err
		}

		mg.AddMachine(machine, node.Name, cid)
		slog.Debug("Created and added the machine config to the machine group")
	}

	return mg, nil
}

func generateCid() uint32 {
	randomCid := rand.Intn(991)
	randomCid += 10
	return uint32(randomCid)
}

func createVmmLogFifo(vmmLogPath string) (*os.File, error) {
	if _, err := os.Stat(vmmLogPath); os.IsNotExist(err) {
		if err := syscall.Mkfifo(vmmLogPath, 0666); err != nil {
			return nil, err
		}
	}

	// Needs to be open for reading and writing.
	f, err := os.OpenFile(config.VmmLogPath, os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func cleanup() {
	_ = os.Remove(config.VmmLogPath)
}
