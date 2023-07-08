#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

variant="minimal"

squashfs_img="rootfs-$variant.squashfs"
squashfs_target_dir="systemd-rootfs-$variant-squashfs"

rootfs_archive_path=$1

function cleanup {
    echo "Cleanup..."
    rm -rf "$squashfs_target_dir"
}

trap cleanup EXIT

function mkroot_squashfs {
    echo "Creating squashfs image of the debian rootfs..."
    mksquashfs "$squashfs_target_dir" "$squashfs_img" -noappend
}

if [[ ! -e "$squashfs_img" ]]; then
    buildctl build --frontend=dockerfile.v0 \
        --local context=. \
        --local dockerfile=. \
        --output "type=local,dest=$squashfs_target_dir" \
        --opt build-arg:ROOTFS_ARCHIVE_PATH="$rootfs_archive_path" \
        --opt "filename=Dockerfile.$variant"
    
    mkroot_squashfs
fi
