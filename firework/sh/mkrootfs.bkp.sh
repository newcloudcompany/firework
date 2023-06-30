#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

source vars.sh

mount_dir="$script_dir/rootfs"
rootfs="$rootfs_dir/rootfs.ext4"
base_rootfs="$downloads_dir/rootfs.tar.xz"
tmp_rootfs="$script_dir/tmp_rootfs"

rm -rf "$rootfs" || true

function cleanup {
    # Unmount the disk image and remove the temporary mount directory   
    {
        umount --lazy "$mount_dir" &> /dev/null || true
        rm -rf "$mount_dir"
        rm -rf "$tmp_rootfs"
    } || true
}

cleanup
trap cleanup EXIT

mkdir -p "$rootfs_dir"

# Create an empty file
echo "Allocating an empty 2GB file..."
truncate -s 2G "$rootfs"

# Create an ext4 filesystem on the file
echo "Creating an ext4 filesystem on the file..."
mkfs -t ext4 "$rootfs"

mkdir -p $mount_dir
mount -o loop "$rootfs" $mount_dir
cp -r debian-bullseye/* $mount_dir/

buildctl build --no-cache --frontend=dockerfile.v0 --local context=. --local dockerfile=. --output type=local,dest="$mount_dir"


# Zip the rootfs
echo "Zipping the rootfs..."
umount --lazy "$mount_dir"
rm -rf "$mount_dir"

gzip -c "$rootfs" > "$HOME/rootfs.ext4.gz"

echo "Done."