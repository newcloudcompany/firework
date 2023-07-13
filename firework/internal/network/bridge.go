package network

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

func createBridge(name string) (*netlink.Bridge, error) {
	la := netlink.NewLinkAttrs()
	la.Name = name
	bridge := &netlink.Bridge{LinkAttrs: la}
	err := netlink.LinkAdd(bridge)
	if err != nil {
		return nil, fmt.Errorf("could not create %s: %w", la.Name, err)
	}

	return bridge, nil
}

func createTapDevice(name string) (*netlink.Tuntap, error) {
	la := netlink.NewLinkAttrs()
	la.Name = name
	dev := &netlink.Tuntap{
		LinkAttrs: la,
		Mode:      0,
	}

	if err := netlink.LinkAdd(dev); err != nil {
		return nil, fmt.Errorf("could not create %s: %w", la.Name, err)
	}

	if err := netlink.LinkSetUp(dev); err != nil {
		return nil, fmt.Errorf("could not create %s: %w", la.Name, err)
	}

	return dev, nil
}

type BridgeNetwork struct {
	bridge *netlink.Bridge
	ipAddr net.IP
}

func NewBridgeNetwork(subnetCidr string, gateway string) (*BridgeNetwork, error) {
	if link, err := netlink.LinkByName(VM_BRIDGE_NAME); err == nil {
		br, ok := link.(*netlink.Bridge)
		if !ok {
			return nil, fmt.Errorf("link %s is not a bridge", VM_BRIDGE_NAME)
		}

		// Get the IP address of the bridge
		addrs, err := netlink.AddrList(br, 0)
		if err != nil || len(addrs) == 0 {
			return nil, fmt.Errorf("failed to get IP address of bridge %s: %w", VM_BRIDGE_NAME, err)
		}

		if err := setupIptables(subnetCidr); err != nil {
			return nil, fmt.Errorf("failed to set up iptables: %w", err)
		}

		// Assume that the route to VM_SUBNET is already added and iptables rules are already set up
		return &BridgeNetwork{br, addrs[0].IP}, nil
	}

	bridge, err := createBridge(VM_BRIDGE_NAME)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge %s: %w", VM_BRIDGE_NAME, err)
	}

	bridgeIpAddr, err := netlink.ParseAddr(gateway)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bridge IP address %s: %w", gateway, err)
	}

	if err := netlink.AddrAdd(bridge, bridgeIpAddr); err != nil {
		return nil, fmt.Errorf("failed to add IP address %s to bridge %s: %w", bridgeIpAddr, VM_BRIDGE_NAME, err)
	}

	if err := netlink.LinkSetUp(bridge); err != nil {
		return nil, fmt.Errorf("failed to set up bridge %s: %w", VM_BRIDGE_NAME, err)
	}

	if err := setupIptables(subnetCidr); err != nil {
		return nil, fmt.Errorf("failed to set up iptables: %w", err)
	}

	return &BridgeNetwork{bridge, bridgeIpAddr.IP}, nil
}

func (n *BridgeNetwork) CreateTapDevice(id string) (*netlink.Tuntap, error) {
	ifaceName := fmt.Sprintf("%s-%s", VM_TAP_PREFIX, id)
	tap, err := createTapDevice(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create tap %s: %w", ifaceName, err)
	}

	if err := netlink.LinkSetMaster(tap, n.bridge); err != nil {
		return nil, fmt.Errorf("failed to set master for tap %s: %w", ifaceName, err)
	}

	return tap, nil
}

func (n *BridgeNetwork) GetIPAddr() net.IP {
	return n.ipAddr
}
