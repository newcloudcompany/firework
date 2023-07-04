package vm

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"os/signal"
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

func (m *Machine) Ipv4() string {
	return m.inner.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.String()
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

type Metadata struct {
	Cid      uint32            `json:"cid"`
	Ipv4     string            `json:"ipv4"`
	Hostname string            `json:"hostname"`
	Hosts    map[string]string `json:"hosts"`
}

func NewMachineGroup() *MachineGroup {
	return &MachineGroup{
		machines: make([]Machine, 0),
		eg:       new(errgroup.Group),
		pidTable: make(PidTable),
	}
}

func (mg *MachineGroup) Start(ctx context.Context) error {
	// Populate hosts map
	hosts := make(map[string]string)
	for _, m := range mg.machines {
		ip, _, err := net.ParseCIDR(m.Ipv4())
		if err != nil {
			return err
		}

		hosts[m.name] = ip.String()
	}

	for _, m := range mg.machines {
		machine := m
		mg.eg.Go(func() error {
			if err := machine.inner.Start(ctx); err != nil {
				return err
			}

			meta, err := createMetadata(Metadata{
				Cid:      machine.cid,
				Ipv4:     machine.Ipv4(),
				Hostname: machine.name,
				Hosts:    hosts,
			})
			if err != nil {
				return err
			}

			if err := machine.inner.SetMetadata(ctx, meta); err != nil {
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

func createMetadata(metadata Metadata) (map[string]interface{}, error) {
	jsonString, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	var apiMetadata map[string]interface{}
	err = json.Unmarshal(jsonString, &apiMetadata)
	if err != nil {
		return nil, err
	}

	return apiMetadata, nil
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
