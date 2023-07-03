package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/jlkiri/firework/internal/config"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

type Machine struct {
	inner *firecracker.Machine
	name  string
	cid   uint32
}

type MachineGroup struct {
	machines []Machine
	eg       *errgroup.Group
	pidTable PidTable
}

type Entry struct {
	VmId string `json:"vm_id"`
	Pid  int    `json:"pid"`
}

// Machine name -> Entry
type PidTable map[string]Entry

func NewMachineGroup() *MachineGroup {
	return &MachineGroup{
		machines: make([]Machine, 0),
		eg:       new(errgroup.Group),
		pidTable: make(PidTable),
	}
}

func (mg *MachineGroup) Start(ctx context.Context) error {
	for _, m := range mg.machines {
		machine := m
		mg.eg.Go(func() error {
			if err := machine.inner.Start(ctx); err != nil {
				return err
			}

			ipAddr := machine.inner.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.String()
			metadata, err := createMetadata(machine.cid, ipAddr)
			if err != nil {
				return err
			}

			if err := machine.inner.SetMetadata(ctx, metadata); err != nil {
				return err
			}

			pid, err := machine.inner.PID()
			if err != nil {
				return err
			}

			vmId := machine.inner.Cfg.VMID
			slog.Debug("Machine started with", "name", machine.name, "vmId", vmId, "pid", pid)

			mg.pidTable[machine.name] = Entry{
				VmId: vmId,
				Pid:  pid,
			}

			if err := mg.updatePidTable(); err != nil {
				return err
			}

			if err := machine.inner.Wait(ctx); err != nil {
				return err
			}

			return nil
		})
	}

	return nil
}

func (mg *MachineGroup) updatePidTable() error {
	// Update the pid table file and create if it does not exist
	bytes, err := json.Marshal(mg.pidTable)
	if err != nil {
		return err
	}

	pidTablePath := config.PidTablePath()
	if err := os.WriteFile(pidTablePath, bytes, 0644); err != nil {
		return err
	}

	return nil
}

func (mg *MachineGroup) Wait(ctx context.Context) error {
	return mg.eg.Wait()
}

func (mg *MachineGroup) Shutdown(ctx context.Context) error {
	for _, m := range mg.machines {
		// info := m.inner.DescribeInstanceInfo()
		if err := m.inner.Shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (mg *MachineGroup) AddMachine(machine *firecracker.Machine, name string, cid uint32) error {
	mg.machines = append(mg.machines, Machine{machine, name, cid})
	return nil
}

func createMetadata(cid uint32, ipAddr string) (map[string]interface{}, error) {
	jsonMetadata := fmt.Sprintf(`
	{
		"cid": "%s",
		"ipv4": "%s"
	}
	`, strconv.Itoa(int(cid)), ipAddr)

	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(jsonMetadata), &metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func InstallSignalHandlers(ctx context.Context, mg *MachineGroup) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case sig := <-c:
				slog.Info("Caught signal: %s, requesting clean shutdown", sig.String())
				if sig == syscall.SIGTERM || sig == os.Interrupt {
					if err := mg.Shutdown(ctx); err != nil {
						slog.Error("an error occurred while shutting down Firecracker VMM", err)
					}
					return
				}
			}
		}
	}()
}
