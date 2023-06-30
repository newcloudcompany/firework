#!/usr/bin/env bash

set -euxo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

source vars.sh

mkdir -p "$data_dir"
mkdir -p "$downloads_dir"

rootfs_url="https://pub-444e85d973ae48759d2f74f88dc82acc.r2.dev/bullseye_rootfs.tar.gz"
# rootfs_url="https://cloud-images.ubuntu.com/minimal/releases/jammy/release/ubuntu-22.04-minimal-cloudimg-amd64-root.tar.xz"
# rootfs_url="https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-minirootfs-3.18.0-x86_64.tar.gz"
firecracker_url="https://github.com/firecracker-microvm/firecracker/releases/download/v1.3.3/firecracker-v1.3.3-x86_64.tgz"

rootfs_base="$downloads_dir/rootfs.tar.xz"

function ensure_rootfs {
    if [[ ! -e "$rootfs_base" ]]; then
        curl -L "$rootfs_url" -o "$rootfs_base"
    fi
}

function ensure_firecracker {
    if ! command -v firecracker; then
        curl -L "$firecracker_url" -o "$downloads_dir/firecracker.tgz"
        tar -xf "$downloads_dir/firecracker.tgz" -C "$downloads_dir"
        cp "$downloads_dir/release-v1.3.3-x86_64/firecracker-v1.3.3-x86_64" /usr/bin/firecracker
    fi
}

function ensure_kernel {
    mkdir -p "$kernel_dir"
    cp "$script_dir/kernel/vmlinux" "$kernel_path"
}

ensure_rootfs
ensure_firecracker
ensure_kernel
