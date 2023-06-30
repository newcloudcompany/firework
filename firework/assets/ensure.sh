#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

rootfs_base="debian-bullseye-rootfs"
mkdir -p "$rootfs_base"

function ensure_rootfs {
    if [[ ! -e "$rootfs_base" ]]; then
        debootstrap --arch=amd64 --variant=minbase bullseye $rootfs_base http://deb.debian.org/debian/
    fi
}

time ensure_rootfs