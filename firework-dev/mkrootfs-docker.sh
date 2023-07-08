#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

squashfs_img="rootfs.squashfs"
squashfs_target_dir="debian-bookworm-rootfs-squashfs"
rootfs_archive_path="debian-bookworm-rootfs.tar.gz"

function cleanup {
    echo "Cleanup..."
    rm -rf "$squashfs_target_dir"
}

trap cleanup EXIT

buildctl build --frontend=dockerfile.v0 \
    --local context=docker \
    --local dockerfile=docker \
    --output type=local,dest=debian-bookworm-rootfs-squashfs \
    --opt build-arg:ROOTFS_ARCHIVE_PATH="$rootfs_archive_path"

function mkroot_squashfs {
    echo "Creating squashfs image of the debian rootfs..."
    mksquashfs "$squashfs_target_dir" "$squashfs_img" -noappend
}

mkroot_squashfs