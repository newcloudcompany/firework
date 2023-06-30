#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd "$script_dir"

public_if="${1:-enp87s0}"
vm_bridge="fcbr0"
vm_subnet="10.0.0.0/16"

function cleanup {
    ip link del "$vm_bridge" || true
    iptables -t nat -D POSTROUTING ! -o "$vm_bridge" -s "$vm_subnet" -j MASQUERADE || true
    iptables -t filter -D FORWARD -i "$vm_bridge" ! -o "$vm_bridge" -j ACCEPT || true
    iptables -t filter -D FORWARD -i "$vm_bridge" -o "$vm_bridge" -j ACCEPT || true
    iptables -t filter -D FORWARD -o "$vm_bridge" -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT || true
    systemctl stop dnsmasq || true
}

cleanup