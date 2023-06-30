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

	// startCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.firework.yaml)")

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

	ipamDb, err := ipam.NewIPAM(dbPath, "10.0.0.240/28")
	if err != nil {
		return err
	}

	bridge, err := network.NewBridgeNetwork()
	if err != nil {
		return err
	}

	config, err := readConfigFromJson("config.json")
	if err != nil {
		return err
	}

	ctx := context.TODO()

	mg, err := createMachineGroup(ctx, config.Nodes, bridge, ipamDb)
	if err != nil {
		return fmt.Errorf("failed to run VMM: %w", err)
	}

	if err := mg.Wait(ctx); err != nil {
		log.Println("An error occurred while waiting for the Firecracker VMM to exit: ", err)
	}

	log.Println("Gracefully shutdown")

	return nil
}

func createMachineGroup(ctx context.Context, nodes []Node, bridge *network.BridgeNetwork, ipamDb *ipam.IPAM) (*vm.MachineGroup, error) {
	kernelPath := filepath.Join(kernelDir, "vmlinux")
	rootFsPath := filepath.Join(rootFsDir, "rootfs.ext4")

	mg := vm.NewMachineGroup()

	for _, node := range nodes {
		id := uuid.NewString()

		tap, err := bridge.CreateTapDevice(id)
		if err != nil {
			return nil, err
		}

		log.Println("Allocating a free IPv4 address...")
		addr, err := ipamDb.AllocateFreeIPAddress(id)
		if err != nil {
			return nil, err
		}

		log.Println("Allocated:", addr)

		rootFsCopyPath, err := createRootFsCopy(rootFsPath, vmDataDir, id)
		if err != nil {
			return nil, err
		}

		socketPath := filepath.Join(vmDataDir, fmt.Sprintf("%s.sock", id))
		fifoPath := filepath.Join(miscDir, fmt.Sprintf("%s.fifo", id))

		ipConfig, err := vm.NewMachineIpConfig(bridge.GetIPAddr(), addr, tap.Name)
		if err != nil {
			return nil, err
		}

		cid := generateCid()

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
	}

	if err := mg.Start(ctx); err != nil {
		return nil, err
	}

	vm.InstallSignalHandlers(ctx, mg)

	return mg, nil
}

func generateCid() uint32 {
	randomCid := rand.Intn(991)
	randomCid += 10
	return uint32(randomCid)
}
