#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

rootfs_base="debian-bookworm-rootfs-systemd"
squashfs_img="rootfs.squashfs"
packages="procps iproute2 ca-certificates curl dnsutils iptables iputils-ping cpu-checker git gnupg systemd"
archive="debian-bookworm-rootfs.tar.gz"

# mkdir -p "$rootfs_base"

function cleanup {
    echo "Cleanup..."
    # rm -rf "$rootfs_base"
}

trap cleanup EXIT

function debootstrap_rootfs {
    if [[ ! -e "$rootfs_base" ]]; then
        debootstrap \
            --arch=amd64 \
            --variant=minbase \
            --include=${packages// /,} \
            bookworm "$rootfs_base" \
            http://deb.debian.org/debian/    
    fi
}

echo "Creating debian bookworm rootfs using debootstrap..."
time debootstrap_rootfs &> /dev/null

echo "Creating compressed tar archive of the rootfs..."
time tar -czvf "$archive" -C "$rootfs_base" . &> /dev/null
