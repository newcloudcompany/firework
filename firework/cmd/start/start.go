package start

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/google/uuid"
	"github.com/jlkiri/firework/internal/config"
	"github.com/jlkiri/firework/internal/ipam"
	"github.com/jlkiri/firework/internal/network"
	"github.com/jlkiri/firework/internal/vm"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

func NewStartCommand() *cobra.Command {
	isDaemon := false

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start a VM cluster from config",
		Long:  `Start a VM cluster from config`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(isDaemon)
		},
	}

	// Add a run in background flag to start command
	startCmd.Flags().BoolVarP(&isDaemon, "daemon", "d", false, "Run in background")
	return startCmd
}

func runStart(isDaemon bool) error {
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

	vmmLogFile, err := createVmmLogFile(config.VmmLogPath)
	if err != nil {
		return fmt.Errorf("failed to create VMM log fifo: %w", err)
	}

	defer vmmLogFile.Close()
	slog.Debug("Created VMM log fifo", "path", config.VmmLogPath)

	mg, err := createMachineGroup(ctx, conf.Nodes, bridge, ipamDb, vmmLogFile)
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

func createMachineGroup(ctx context.Context, nodes []config.Node, bridge *network.BridgeNetwork, ipamDb *ipam.IPAM, fifoLogWriter io.Writer) (*vm.MachineGroup, error) {
	kernelPath := config.KernelPath()
	// rootFsPath := config.RootFsPath()

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

		overlayDrivePath, err := createOverlayDrive(id, node.Disk)
		if err != nil {
			return nil, err
		}

		stdio, err := createStdioWriter(id)
		if err != nil {
			return nil, err
		}

		machine, err := vm.CreateMachine(ctx, vm.MachineOptions{
			Id:                    id,
			RootFsPath:            node.RootFsPath,
			KernelImagePath:       kernelPath,
			SocketPath:            socketPath,
			InstanceLogFifoPath:   logFifoPath,
			InstanceFifoLogWriter: fifoLogWriter,
			Stdio:                 stdio,
			MetricsFifoPath:       metricsFifoPath,
			OverlayDrivePath:      overlayDrivePath,
			VmmLogPath:            config.VmmLogPath,
			VsockPath:             config.VsockPath(node.Name),
			Cid:                   cid,
			Vcpu:                  node.Vcpu,
			Memory:                node.Memory,
			IpConfig:              ipConfig,
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

func createVmmLogFile(vmmLogPath string) (*os.File, error) {
	f, err := os.Create(vmmLogPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func createStdioWriter(vmId string) (*os.File, error) {
	f, err := os.Create(config.StdioPath(vmId))
	if err != nil {
		return nil, err
	}

	return f, nil
}

func cleanup() {
	// _ = os.Remove(config.VmmLogPath)
}
