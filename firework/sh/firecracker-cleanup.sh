#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

source vars.sh

mkdir -p $tmp_vm_dir
mkdir -p $socket_dir

pattern="^fctapvm"

# List all network interfaces and filter them based on the pattern
interfaces=$(ip --brief link | awk '{print $1}' | grep -E "$pattern" || true)

function cleanup {
  pkill -f "firecracker --api-sock" || true
  
  # Iterate through the filtered list and delete each matching interface
  for interface in $interfaces; do
      ip link delete "$interface"
  done

  echo "DONE: Delete all firecracker tap interfaces"
  
  rm -f $socket_dir/*.socket
  rm -f $tmp_vm_dir/*.ext4
  rm -f $tmp_vm_dir/*.json
  rm -f $tmp_vm_dir/ssh/*
  rm -f $tmp_vm_dir/v.sock

  umount --lazy "$tmp_vm_dir/mnt" &> /dev/null || true

  echo "DONE: Delete all firecracker related files"
}

cleanup
