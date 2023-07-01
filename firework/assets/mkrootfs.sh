#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

rootfs_base="debian-bullseye-rootfs"
squashfs_img="rootfs.squashfs"
packages="procps iproute2 ca-certificates curl dnsutils iputils-ping cpu-checker"

mkdir -p "$rootfs_base"

function debootstrap_rootfs {
    if [[ ! -e "$rootfs_base" ]]; then
        echo "Creating debian bullseye rootfs..."
        debootstrap \
            --arch=amd64 \
            --variant=minbase \
            --include=${packages// /,} \
            bullseye "$rootfs_base" \
            http://deb.debian.org/debian/    
    fi
    
    echo "Copying init to the rootfs..."
    mkdir -p "$rootfs_base/overlay" "$rootfs_base/mnt" "$rootfs_base/rom"
    cp init "$rootfs_base/sbin/init"
    cp overlay-init "$rootfs_base/sbin/overlay-init"
}

function mkroot_squashfs {
    if [[ ! -e "$squashfs_img" ]]; then
        echo "Creating squashfs image of the debian rootfs..."
        mksquashfs "$rootfs_base" "$squashfs_img" -noappend
    fi
}

time debootstrap_rootfs
time mkroot_squashfs

