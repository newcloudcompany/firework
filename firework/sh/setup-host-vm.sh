#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd "$script_dir"

public_if="${1:-enp87s0}"
vm_bridge="fcbr0"
vm_subnet="10.10.0.0/16"

function cleanup {
    ip link del "$vm_bridge" || true
    iptables -t nat -D POSTROUTING ! -o "$vm_bridge" -s "$vm_subnet" -j MASQUERADE || true
    iptables -t filter -D FORWARD -i "$vm_bridge" ! -o "$vm_bridge" -j ACCEPT || true
    iptables -t filter -D FORWARD -i "$vm_bridge" -o "$vm_bridge" -j ACCEPT || true
    iptables -t filter -D FORWARD -o "$vm_bridge" -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT || true
    systemctl stop dnsmasq || true
}

function add_default_iptables {
    set +e
    iptables -t nat -C POSTROUTING ! -o "$vm_bridge" -s "$vm_subnet" -j MASQUERADE &> /dev/null

    # If the rule does not exist, add it
    if [ $? -ne 0 ]; then
        iptables -t nat -A POSTROUTING ! -o "$vm_bridge" -s "$vm_subnet" -j MASQUERADE &> /dev/null
        iptables -t filter -A FORWARD -i "$vm_bridge" ! -o "$vm_bridge" -j ACCEPT &> /dev/null
        iptables -t filter -A FORWARD -i "$vm_bridge" -o "$vm_bridge" -j ACCEPT &> /dev/null
        iptables -t filter -A FORWARD -o "$vm_bridge" -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT &> /dev/null
    fi
    set -e
}

cleanup

add_default_iptables
echo "DONE: Setup masquerade iptables for default namespace"

FC_BRIDGE_NAME="$vm_bridge"
FC_BRIDGE_IPv4_ADDR_CIDR="10.10.0.1/16"
FC_BRIDGE_IPv4_ADDR="10.10.0.1"
DHCP_RANGE_START="10.10.0.2"
DNSMASQ_FC_CONF="/etc/dnsmasq.conf"

# Create dnsmasq configuration file
cat > $DNSMASQ_FC_CONF <<EOF
interface=fcbr0
bind-interfaces

cache-size=2048
no-resolv
no-poll
bogus-priv

server=8.8.8.8

dhcp-range=$DHCP_RANGE_START,10.10.255.254,255.255.0.0,12h
dhcp-option=option:router,$FC_BRIDGE_IPv4_ADDR
dhcp-option=option:dns-server,$FC_BRIDGE_IPv4_ADDR
dhcp-authoritative

log-facility=/var/log/dnsmasq.log
log-dhcp
EOF

ip link add $FC_BRIDGE_NAME type bridge
ip addr add $FC_BRIDGE_IPv4_ADDR_CIDR dev $FC_BRIDGE_NAME
ip link set $FC_BRIDGE_NAME up

echo "DONE: Setup bridge with a dnsmasq DHCP server. IPv4: $FC_BRIDGE_IPv4_ADDR_CIDR"

systemctl restart dnsmasq
