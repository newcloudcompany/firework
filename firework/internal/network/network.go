package network

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/coreos/go-iptables/iptables"
	"github.com/vishvananda/netlink"
)

const (
	VM_BRIDGE_NAME = "firework0"
	VM_TAP_PREFIX  = "tap-firework"
)

type Chain string

const (
	ChainForward     Chain = "FORWARD"
	ChainPostrouting Chain = "POSTROUTING"
)

type Table string

const (
	TableNat    Table = "nat"
	TableFilter Table = "filter"
)

type Target string

const (
	TargetAccept     Target = "ACCEPT"
	TargetMasquerade Target = "MASQUERADE"
)

func cleanupIptables(subnetCidr string) error {
	path, ok := os.LookupEnv("PATH")
	if !ok {
		return fmt.Errorf("PATH not set")
	}

	fmt.Println("PATH", path)

	ipt, err := iptables.New()
	if err != nil {
		return err
	}

	if err := ipt.DeleteIfExists(string(TableNat), string(ChainPostrouting), "!", "-o", VM_BRIDGE_NAME, "-s", subnetCidr, "-j", string(TargetMasquerade)); err != nil {
		return err
	}
	if err := ipt.DeleteIfExists(string(TableFilter), string(ChainForward), "-i", VM_BRIDGE_NAME, "!", "-o", VM_BRIDGE_NAME, "-j", string(TargetAccept)); err != nil {
		return err
	}
	if err := ipt.DeleteIfExists(string(TableFilter), string(ChainForward), "-i", VM_BRIDGE_NAME, "-o", VM_BRIDGE_NAME, "-j", string(TargetAccept)); err != nil {
		return err

	}
	if err := ipt.DeleteIfExists(string(TableFilter), string(ChainForward), "-o", VM_BRIDGE_NAME, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", string(TargetAccept)); err != nil {
		return err
	}

	return nil
}

func setupIptables(subnetCidr string) error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}

	// Add default iptables
	if err := ipt.AppendUnique(string(TableNat), string(ChainPostrouting), "!", "-o", VM_BRIDGE_NAME, "-s", subnetCidr, "-j", string(TargetMasquerade)); err != nil {
		return err
	}
	if err := ipt.AppendUnique(string(TableFilter), string(ChainForward), "-i", VM_BRIDGE_NAME, "!", "-o", VM_BRIDGE_NAME, "-j", string(TargetAccept)); err != nil {
		return err
	}
	if err := ipt.AppendUnique(string(TableFilter), string(ChainForward), "-i", VM_BRIDGE_NAME, "-o", VM_BRIDGE_NAME, "-j", string(TargetAccept)); err != nil {
		return err
	}
	if err := ipt.AppendUnique(string(TableFilter), string(ChainForward), "-o", VM_BRIDGE_NAME, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", string(TargetAccept)); err != nil {
		return err
	}

	return nil
}

func Cleanup(subnetCidr string) error {
	if err := cleanupIptables(subnetCidr); err != nil {
		return fmt.Errorf("failed to cleanup iptables: %w", err)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("failed to get interfaces: %w", err)
	}

	for _, iface := range ifaces {
		if strings.HasPrefix(iface.Name, VM_TAP_PREFIX) {
			link, err := netlink.LinkByName(iface.Name)
			if err != nil {
				return fmt.Errorf("failed to get link %s: %w", iface.Name, err)
			}

			if err := netlink.LinkDel(link); err != nil {
				return fmt.Errorf("failed to delete link %s: %w", iface.Name, err)
			}
		}
	}

	return nil
}
