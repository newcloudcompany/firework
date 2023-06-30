package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"golang.org/x/sync/errgroup"
)

type Machine struct {
	inner *firecracker.Machine
	cid   uint32
}

type MachineGroup struct {
	machines []Machine
	eg       *errgroup.Group
}

func NewMachineGroup() *MachineGroup {
	return &MachineGroup{
		machines: make([]Machine, 0),
		eg:       new(errgroup.Group),
	}
}

func (mg *MachineGroup) Start(ctx context.Context) error {
	for _, m := range mg.machines {
		machine := m
		mg.eg.Go(func() error {
			if err := machine.inner.Start(ctx); err != nil {
				return err
			}

			type Metadata map[string]interface{}

			jsonMetadata := fmt.Sprintf(`
			{
				"latest": {
					"meta-data": {
						"cid": "%s"
					}
				}
			}
			`, strconv.Itoa(int(machine.cid)))

			var metadata Metadata

			err := json.Unmarshal([]byte(jsonMetadata), &metadata)
			if err != nil {
				return err
			}

			if err := machine.inner.SetMetadata(ctx, metadata); err != nil {
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

func (mg *MachineGroup) Wait(ctx context.Context) error {
	return mg.eg.Wait()
}

func (mg *MachineGroup) Shutdown(ctx context.Context) error {
	for _, m := range mg.machines {
		if err := m.inner.Shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (mg *MachineGroup) AddMachine(machine *firecracker.Machine, cid uint32) error {
	mg.machines = append(mg.machines, Machine{machine, cid})
	return nil
}

func InstallSignalHandlers(ctx context.Context, mg *MachineGroup) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	for sig := range c {
		log.Printf("Caught signal: %s, requesting clean shutdown", sig.String())
		if sig == syscall.SIGTERM || sig == os.Interrupt {
			if err := mg.Shutdown(ctx); err != nil {
				log.Println("An error occurred while shutting down Firecracker VMM", err)
			}
			log.Printf("Clean shutdown successful!")
			break
		}
	}
}
