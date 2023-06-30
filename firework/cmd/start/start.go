package start

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jlkiri/firework/internal/ipam"
	"github.com/jlkiri/firework/internal/network"
	"github.com/jlkiri/firework/internal/vm"
	"github.com/jlkiri/firework/sources"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

var cacheDir = filepath.Join(sources.DataDir, "cache")
var kernelDir = filepath.Join(cacheDir, "kernel")
var rootFsDir = filepath.Join(cacheDir, "rootfs")
var vmDataDir = filepath.Join(sources.DataDir, "vm")
var miscDir = filepath.Join(sources.DataDir, "misc")
var dbPath = filepath.Join(miscDir, "ips.db")

type Node struct {
	Name string `json:"name"`
}

type Config struct {
	Nodes []Node `json:"nodes"`
}

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

func cleanup() {
	if err := network.Cleanup(); err != nil {
		log.Fatalf("Failed to cleanup network: %v", err)
	}

	if err := os.Remove(filepath.Join(miscDir, "ips.db")); err != nil {
		log.Println("Failed to remove ips.db:", err)
	}

	if err := os.RemoveAll(filepath.Join(vmDataDir)); err != nil {
		log.Println("Failed to remove vm data dir:", err)
	}
}

func runStart() error {
	defer func() {
		if err := recover(); err != nil {
			cleanup()
		}
	}()

	defer cleanup()

	// TODO: Remove this
	os.Remove(dbPath)

	if err := prepareEnvironment(); err != nil {
		return err
	}
	slog.Debug("Prepared environment for execution.")

	ipamDb, err := ipam.NewIPAM(dbPath, "10.0.0.240/28")
	if err != nil {
		return err
	}
	slog.Debug("Created and populated IPAM database.")

	bridge, err := network.NewBridgeNetwork()
	if err != nil {
		return err
	}
	slog.Debug("Created a bridge network.")

	config, err := readConfigFromJson("config.json")
	if err != nil {
		return err
	}
	slog.Debug("Read config.json.")

	ctx := context.TODO()
	mg, err := createMachineGroup(ctx, config.Nodes, bridge, ipamDb)
	if err != nil {
		return fmt.Errorf("failed to create machine group: %w", err)
	}
	slog.Debug("Created machine group from config.")

	if err := mg.Start(ctx); err != nil {
		return fmt.Errorf("failed to start machine group: %w", err)
	}

	vm.InstallSignalHandlers(ctx, mg)
	slog.Debug("Installed signal handlers.")

	if err := mg.Wait(ctx); err != nil {
		return fmt.Errorf("an error occurred while waiting for the machine group to exit: %w", err)
	}

	slog.Info("Graceful shutdown successful.")

	return nil
}

func createMachineGroup(ctx context.Context, nodes []Node, bridge *network.BridgeNetwork, ipamDb *ipam.IPAM) (*vm.MachineGroup, error) {
	kernelPath := filepath.Join(kernelDir, "vmlinux")
	rootFsPath := filepath.Join(rootFsDir, "rootfs.ext4")

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

		rootFsCopyPath, err := createRootFsCopy(rootFsPath, vmDataDir, id)
		if err != nil {
			return nil, err
		}
		slog.Debug("Created a copy of base rootfs", "node", node.Name)

		socketPath := filepath.Join(vmDataDir, fmt.Sprintf("%s.sock", id))
		fifoPath := filepath.Join(miscDir, fmt.Sprintf("%s.fifo", id))
		ipConfig, err := vm.NewMachineIpConfig(bridge.GetIPAddr(), addr, tap.Name)
		if err != nil {
			return nil, err
		}

		machine, err := vm.CreateMachine(ctx, vm.MachineOptions{
			RootFsPath:      rootFsCopyPath,
			KernelImagePath: kernelPath,
			SocketPath:      socketPath,
			FifoPath:        fifoPath,
			Id:              id,
			Cid:             cid,
			VsockPath:       filepath.Join(vmDataDir, fmt.Sprintf("%s-v.sock", node.Name)),
			IpConfig:        ipConfig,
		})
		if err != nil {
			return nil, err
		}

		mg.AddMachine(machine, cid)
		slog.Debug("Created and added the machine config to the machine group")
	}

	return mg, nil
}

func generateCid() uint32 {
	randomCid := rand.Intn(991)
	randomCid += 10
	return uint32(randomCid)
}
